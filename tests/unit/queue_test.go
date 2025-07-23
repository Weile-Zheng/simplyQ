package unit_test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/weilezheng/simplyQ/internal/queue"
)

var config = queue.QueueConfig{
	Name:              "Test Queue",
	Type:              queue.QueueTypeFIFO,
	RetentionPeriod:   time.Hour,
	VisibilityTimeout: time.Minute,
	MaxReceiveCount:   3,
	MaxMessageSize:    1024,
}

func TestMakeQueue(t *testing.T) {
	queueIO := queue.MakeQueue("id", config)
	defer queueIO.Close()

	// Test that queue is created successfully
	if queueIO.SendChan == nil {
		t.Error("SendChan should not be nil")
	}
	if queueIO.End == nil {
		t.Error("End channel should not be nil")
	}
}

func TestInsertQueue(t *testing.T) {
	queueIO := queue.MakeQueue("id", config)
	defer queueIO.Close()

	message := queue.Message{
		ID:   "msg-1",
		Body: "Hello, World!",
	}

	req := queue.Request{
		Message: message,
	}

	response := queueIO.InsertQueue(req)

	if response.Code != queue.OK {
		t.Errorf("Expected OK, got %v", response.Code)
	}
	if response.Message.ID != message.ID {
		t.Errorf("Expected message ID %s, got %s", message.ID, response.Message.ID)
	}
	if response.Message.Body != message.Body {
		t.Errorf("Expected message body %s, got %s", message.Body, response.Message.Body)
	}
}

func TestPeekQueue(t *testing.T) {
	queueIO := queue.MakeQueue("id", config)
	defer queueIO.Close()

	// Test peek on empty queue
	req := queue.Request{}
	response := queueIO.PeekQueue(req)
	if response.Code != queue.EMPTY_QUEUE {
		t.Errorf("Expected EMPTY_QUEUE, got %v", response.Code)
	}

	// Insert a message and test peek
	message := queue.Message{
		ID:   "msg-2",
		Body: "Test message",
	}
	insertReq := queue.Request{Message: message}
	queueIO.InsertQueue(insertReq)

	response = queueIO.PeekQueue(req)
	if response.Code != queue.OK {
		t.Errorf("Expected OK, got %v", response.Code)
	}
	if response.Message.ID != message.ID {
		t.Errorf("Expected message ID %s, got %s", message.ID, response.Message.ID)
	}

	// Peek again to ensure message is still there
	response = queueIO.PeekQueue(req)
	if response.Code != queue.OK {
		t.Errorf("Expected OK on second peek, got %v", response.Code)
	}
}

func TestRemoveQueue(t *testing.T) {
	queueIO := queue.MakeQueue("id", config)
	defer queueIO.Close()

	// Test delete on empty queue
	req := queue.Request{}
	response := queueIO.RemoveQueue(req)
	if response.Code != queue.EMPTY_QUEUE {
		t.Errorf("Expected EMPTY_QUEUE, got %v", response.Code)
	}

	// Insert a message and test delete
	message := queue.Message{
		ID:   "msg-3",
		Body: "To be deleted",
	}
	insertReq := queue.Request{Message: message}
	queueIO.InsertQueue(insertReq)

	response = queueIO.RemoveQueue(req)
	if response.Code != queue.OK {
		t.Errorf("Expected OK, got %v", response.Code)
	}

	// Verify message is deleted by peeking
	peekResponse := queueIO.PeekQueue(req)
	if peekResponse.Code != queue.EMPTY_QUEUE {
		t.Errorf("Expected EMPTY_QUEUE after delete, got %v", peekResponse.Code)
	}
}

func TestRequeueQueue(t *testing.T) {
	queueIO := queue.MakeQueue("id", config)
	defer queueIO.Close()

	// Test requeue on empty dead letter queue
	req := queue.Request{}
	response := queueIO.RequeueQueue(req)
	if response.Code != queue.EMPTY_DEAD_LETTER_QUEUE {
		t.Errorf("Expected EMPTY_DEAD_LETTER_QUEUE, got %v", response.Code)
	}
}

func TestQueueFIFOBehavior(t *testing.T) {
	queueIO := queue.MakeQueue("id", config)
	defer queueIO.Close()

	// Insert multiple messages
	messages := []queue.Message{
		{ID: "msg-1", Body: "First"},
		{ID: "msg-2", Body: "Second"},
		{ID: "msg-3", Body: "Third"},
	}

	for _, msg := range messages {
		req := queue.Request{Message: msg}
		queueIO.InsertQueue(req)
	}

	// Peek should return first message
	peekReq := queue.Request{}
	response := queueIO.PeekQueue(peekReq)
	if response.Message.ID != "msg-1" {
		t.Errorf("Expected first message, got %s", response.Message.ID)
	}

	// Delete should remove first message
	deleteReq := queue.Request{}
	queueIO.RemoveQueue(deleteReq)

	// Peek should now return second message
	response = queueIO.PeekQueue(peekReq)
	if response.Message.ID != "msg-2" {
		t.Errorf("Expected second message after delete, got %s", response.Message.ID)
	}
}

func TestConcurrentOperations(t *testing.T) {
	queueIO := queue.MakeQueue("id", config)

	defer queueIO.Close()

	// Test concurrent inserts
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			message := queue.Message{
				ID:   fmt.Sprintf("msg-%d", id),
				Body: fmt.Sprintf("Message %d", id),
			}
			req := queue.Request{Message: message}
			response := queueIO.InsertQueue(req)
			if response.Code != queue.OK {
				t.Errorf("Insert failed for message %d", id)
			}
			done <- true
		}(i)
	}

	// Wait for all inserts to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify we can peek and get a message
	peekReq := queue.Request{}
	response := queueIO.PeekQueue(peekReq)
	if response.Code != queue.OK {
		t.Errorf("Expected OK after concurrent inserts, got %v", response.Code)
	}

	// Count messages to verify all were inserted
	messageCount := 0
	for {
		peekResp := queueIO.PeekQueue(queue.Request{})
		if peekResp.Code == queue.EMPTY_QUEUE {
			break
		}
		messageCount++
		deleteResp := queueIO.RemoveQueue(queue.Request{})
		if deleteResp.Code != queue.OK {
			t.Errorf("Failed to delete message %d", messageCount)
		}
	}

	if messageCount != 10 {
		t.Errorf("Expected 10 messages, got %d", messageCount)
	}
}

func TestConcurrentMixedOperations(t *testing.T) {
	queueIO := queue.MakeQueue("id", config)

	defer queueIO.Close()

	// Pre-populate with some messages
	for i := 0; i < 5; i++ {
		msg := queue.Message{ID: fmt.Sprintf("init-%d", i), Body: "initial"}
		queueIO.InsertQueue(queue.Request{Message: msg})
	}

	var insertCount, deleteCount, peekCount int32
	done := make(chan bool, 30)

	// Concurrent inserts
	for i := 0; i < 10; i++ {
		go func(id int) {
			message := queue.Message{
				ID:   fmt.Sprintf("concurrent-%d", id),
				Body: fmt.Sprintf("Body %d", id),
			}
			response := queueIO.InsertQueue(queue.Request{Message: message})
			if response.Code == queue.OK {
				atomic.AddInt32(&insertCount, 1)
			}
			done <- true
		}(i)
	}

	// Concurrent deletes
	for i := 0; i < 10; i++ {
		go func() {
			response := queueIO.RemoveQueue(queue.Request{})
			if response.Code == queue.OK {
				atomic.AddInt32(&deleteCount, 1)
			}
			done <- true
		}()
	}

	// Concurrent peeks
	for i := 0; i < 10; i++ {
		go func() {
			response := queueIO.PeekQueue(queue.Request{})
			if response.Code == queue.OK {
				atomic.AddInt32(&peekCount, 1)
			}
			done <- true
		}()
	}

	// Wait for all operations
	for i := 0; i < 30; i++ {
		<-done
	}

	// Verify queue is still functional
	testMsg := queue.Message{ID: "final-test", Body: "test"}
	response := queueIO.InsertQueue(queue.Request{Message: testMsg})
	if response.Code != queue.OK {
		t.Error("Queue not functional after concurrent operations")
	}
}

func TestCloseQueue(t *testing.T) {
	queueIO := queue.MakeQueue("id", config)

	// Insert a message
	message := queue.Message{ID: "msg-1", Body: "Test"}
	req := queue.Request{Message: message}
	response := queueIO.InsertQueue(req)
	if response.Code != queue.OK {
		t.Errorf("Expected OK, got %v", response.Code)
	}

	// Close the queue
	queueIO.Close()

	// Give some time for goroutine to process close
	time.Sleep(10 * time.Millisecond)
}

func TestMaxReceive(t *testing.T) {
	queueIO := queue.MakeQueue("id", config)

	defer queueIO.Close()

	// Insert a message
	message := queue.Message{ID: "msg-1", Body: "Test"}
	req := queue.Request{Message: message}
	response := queueIO.InsertQueue(req)
	if response.Code != queue.OK {
		t.Errorf("Expected OK, got %v", response.Code)
	}

	// Peek the message multiple times
	for i := 0; i < int(config.MaxReceiveCount); i++ {
		response = queueIO.PeekQueue(queue.Request{})
		if response.Code != queue.OK {
			t.Errorf("Expected OK, got %v", response.Code)
		}
	}

	// Now the message should be in the dead letter queue
	response = queueIO.PeekQueue(queue.Request{})
	if response.Code != queue.EMPTY_QUEUE {
		t.Errorf("Expected EMPTY_QUEUE, got %v", response.Code)
	}
}
