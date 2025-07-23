package unit_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/weilezheng/simplyQ/internal/queue"
)

func TestNewQueueManager(t *testing.T) {
	config := queue.QueueManagerConfig{
		Name: "TestManager",
	}

	qm := queue.NewQueueManager(config)

	if qm.Queues == nil {
		t.Error("Queues map should not be nil")
	}

	if len(qm.Queues) != 0 {
		t.Error("Queues map should be empty initially")
	}
}

func TestCreateQueue(t *testing.T) {
	config := queue.QueueManagerConfig{
		Name: "TestManager",
	}
	qm := queue.NewQueueManager(config)

	queueConfig := queue.QueueConfig{
		Name:              "TestQueue",
		Type:              queue.QueueTypeFIFO,
		RetentionPeriod:   time.Hour,
		VisibilityTimeout: time.Minute,
		MaxReceiveCount:   3,
		MaxMessageSize:    1024,
	}

	// Test successful queue creation
	code := qm.CreateQueue(queueConfig)
	if code != queue.OK {
		t.Errorf("Expected OK, got %v", code)
	}

	expectedID := "queue-" + queueConfig.Name
	if _, exists := qm.Queues[expectedID]; !exists {
		t.Error("Queue should exist in manager")
	}

	// Test creating queue with same name
	code = qm.CreateQueue(queueConfig)
	if code != queue.QUEUE_ALREADY_EXISTS {
		t.Errorf("Expected QUEUE_ALREADY_EXISTS, got %v", code)
	}
}

func TestDeleteQueue(t *testing.T) {
	config := queue.QueueManagerConfig{
		Name: "TestManager",
	}
	qm := queue.NewQueueManager(config)

	queueConfig := queue.QueueConfig{
		Name:              "TestQueue",
		Type:              queue.QueueTypeFIFO,
		RetentionPeriod:   time.Hour,
		VisibilityTimeout: time.Minute,
		MaxReceiveCount:   3,
		MaxMessageSize:    1024,
	}

	// Create a queue first
	qm.CreateQueue(queueConfig)
	queueID := "queue-" + queueConfig.Name

	// Test successful deletion
	code := qm.DeleteQueue(queueID)
	if code != queue.OK {
		t.Errorf("Expected OK, got %v", code)
	}

	if _, exists := qm.Queues[queueID]; exists {
		t.Error("Queue should not exist after deletion")
	}

	// Test deleting non-existent queue
	code = qm.DeleteQueue("non-existent")
	if code != queue.QUEUE_NOT_FOUND {
		t.Errorf("Expected QUEUE_NOT_FOUND, got %v", code)
	}
}

func TestSendMessage(t *testing.T) {
	config := queue.QueueManagerConfig{
		Name: "TestManager",
	}
	qm := queue.NewQueueManager(config)

	queueConfig := queue.QueueConfig{
		Name:              "TestQueue",
		Type:              queue.QueueTypeFIFO,
		RetentionPeriod:   time.Hour,
		VisibilityTimeout: time.Minute,
		MaxReceiveCount:   3,
		MaxMessageSize:    1024,
	}

	// Create a queue
	qm.CreateQueue(queueConfig)
	queueID := "queue-" + queueConfig.Name

	message := queue.Message{
		ID:   "msg-1",
		Body: "Test message",
	}

	req := queue.Request{
		Message: message,
	}

	// Test successful message send
	response := qm.SendMessage(queueID, req)
	if response.Code != queue.OK {
		t.Errorf("Expected OK, got %v", response.Code)
	}

	// Test sending to non-existent queue
	response = qm.SendMessage("non-existent", req)
	if response.Code != queue.QUEUE_NOT_FOUND {
		t.Errorf("Expected QUEUE_NOT_FOUND, got %v", response.Code)
	}
}

func TestPeekMessage(t *testing.T) {
	config := queue.QueueManagerConfig{
		Name: "TestManager",
	}
	qm := queue.NewQueueManager(config)

	queueConfig := queue.QueueConfig{
		Name:              "TestQueue",
		Type:              queue.QueueTypeFIFO,
		RetentionPeriod:   time.Hour,
		VisibilityTimeout: time.Minute,
		MaxReceiveCount:   3,
		MaxMessageSize:    1024,
	}

	// Create a queue
	qm.CreateQueue(queueConfig)
	queueID := "queue-" + queueConfig.Name

	// Test peek on empty queue
	response := qm.PeekMessage(queueID)
	if response.Code != queue.EMPTY_QUEUE {
		t.Errorf("Expected EMPTY_QUEUE, got %v", response.Code)
	}

	// Send a message first
	message := queue.Message{
		ID:   "msg-1",
		Body: "Test message",
	}
	req := queue.Request{Message: message}
	qm.SendMessage(queueID, req)

	// Test successful peek
	response = qm.PeekMessage(queueID)
	if response.Code != queue.OK {
		t.Errorf("Expected OK, got %v", response.Code)
	}
	if response.Message.ID != message.ID {
		t.Errorf("Expected message ID %s, got %s", message.ID, response.Message.ID)
	}

	// Test peek on non-existent queue
	response = qm.PeekMessage("non-existent")
	if response.Code != queue.QUEUE_NOT_FOUND {
		t.Errorf("Expected QUEUE_NOT_FOUND, got %v", response.Code)
	}
}

func TestDeleteMessage(t *testing.T) {
	config := queue.QueueManagerConfig{
		Name: "TestManager",
	}
	qm := queue.NewQueueManager(config)

	queueConfig := queue.QueueConfig{
		Name:              "TestQueue",
		Type:              queue.QueueTypeFIFO,
		RetentionPeriod:   time.Hour,
		VisibilityTimeout: time.Minute,
		MaxReceiveCount:   3,
		MaxMessageSize:    1024,
	}

	// Create a queue
	qm.CreateQueue(queueConfig)
	queueID := "queue-" + queueConfig.Name

	// Test delete on empty queue
	req := queue.Request{}
	response := qm.DeleteMessage(queueID, req)
	if response.Code != queue.EMPTY_QUEUE {
		t.Errorf("Expected EMPTY_QUEUE, got %v", response.Code)
	}

	// Send a message first
	message := queue.Message{
		ID:   "msg-1",
		Body: "Test message",
	}
	sendReq := queue.Request{Message: message}
	qm.SendMessage(queueID, sendReq)

	// Test successful delete
	response = qm.DeleteMessage(queueID, req)
	if response.Code != queue.OK {
		t.Errorf("Expected OK, got %v", response.Code)
	}

	// Verify message is deleted
	peekResponse := qm.PeekMessage(queueID)
	if peekResponse.Code != queue.EMPTY_QUEUE {
		t.Errorf("Expected EMPTY_QUEUE after delete, got %v", peekResponse.Code)
	}

	// Test delete on non-existent queue
	response = qm.DeleteMessage("non-existent", req)
	if response.Code != queue.QUEUE_NOT_FOUND {
		t.Errorf("Expected QUEUE_NOT_FOUND, got %v", response.Code)
	}
}

func TestMultipleQueues(t *testing.T) {
	config := queue.QueueManagerConfig{
		Name: "TestManager",
	}
	qm := queue.NewQueueManager(config)

	// Create multiple queues
	for i := 0; i < 3; i++ {
		queueConfig := queue.QueueConfig{
			Name:              fmt.Sprintf("TestQueue%d", i),
			Type:              queue.QueueTypeFIFO,
			RetentionPeriod:   time.Hour,
			VisibilityTimeout: time.Minute,
			MaxReceiveCount:   3,
			MaxMessageSize:    1024,
		}

		code := qm.CreateQueue(queueConfig)
		if code != queue.OK {
			t.Errorf("Failed to create queue %d", i)
		}
	}

	// Verify all queues exist
	if len(qm.Queues) != 3 {
		t.Errorf("Expected 3 queues, got %d", len(qm.Queues))
	}

	// Send messages to different queues
	for i := 0; i < 3; i++ {
		queueID := fmt.Sprintf("queue-TestQueue%d", i)
		message := queue.Message{
			ID:   fmt.Sprintf("msg-%d", i),
			Body: fmt.Sprintf("Message %d", i),
		}
		req := queue.Request{Message: message}

		response := qm.SendMessage(queueID, req)
		if response.Code != queue.OK {
			t.Errorf("Failed to send message to queue %d", i)
		}
	}

	// Verify messages in different queues
	for i := 0; i < 3; i++ {
		queueID := fmt.Sprintf("queue-TestQueue%d", i)
		response := qm.PeekMessage(queueID)
		if response.Code != queue.OK {
			t.Errorf("Failed to peek message from queue %d", i)
		}
		expectedID := fmt.Sprintf("msg-%d", i)
		if response.Message.ID != expectedID {
			t.Errorf("Expected message ID %s, got %s", expectedID, response.Message.ID)
		}
	}
}

func TestConcurrentQueueOperations(t *testing.T) {
	config := queue.QueueManagerConfig{
		Name: "TestManager",
	}
	qm := queue.NewQueueManager(config)

	queueConfig := queue.QueueConfig{
		Name:              "ConcurrentTestQueue",
		Type:              queue.QueueTypeFIFO,
		RetentionPeriod:   time.Hour,
		VisibilityTimeout: time.Minute,
		MaxReceiveCount:   3,
		MaxMessageSize:    1024,
	}

	// Create a queue
	qm.CreateQueue(queueConfig)
	queueID := "queue-" + queueConfig.Name

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent sends
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			message := queue.Message{
				ID:   fmt.Sprintf("concurrent-msg-%d", id),
				Body: fmt.Sprintf("Concurrent message %d", id),
			}
			req := queue.Request{Message: message}
			response := qm.SendMessage(queueID, req)
			if response.Code != queue.OK {
				t.Errorf("Failed to send concurrent message %d", id)
			}
		}(i)
	}

	wg.Wait()

	// Verify messages were sent
	messageCount := 0
	for {
		response := qm.PeekMessage(queueID)
		if response.Code == queue.EMPTY_QUEUE {
			break
		}
		messageCount++
		deleteResponse := qm.DeleteMessage(queueID, queue.Request{})
		if deleteResponse.Code != queue.OK {
			t.Errorf("Failed to delete message %d", messageCount)
		}
	}

	if messageCount != numGoroutines {
		t.Errorf("Expected %d messages, got %d", numGoroutines, messageCount)
	}
}

func TestConcurrentQueueCreationDeletion(t *testing.T) {
	config := queue.QueueManagerConfig{
		Name: "TestManager",
	}
	qm := queue.NewQueueManager(config)

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent queue creation
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			queueConfig := queue.QueueConfig{
				Name:              fmt.Sprintf("ConcurrentQueue%d", id),
				Type:              queue.QueueTypeFIFO,
				RetentionPeriod:   time.Hour,
				VisibilityTimeout: time.Minute,
				MaxReceiveCount:   3,
				MaxMessageSize:    1024,
			}

			code := qm.CreateQueue(queueConfig)
			if code != queue.OK {
				t.Errorf("Failed to create concurrent queue %d", id)
			}
		}(i)
	}

	wg.Wait()

	// Verify all queues were created
	if len(qm.Queues) != numGoroutines {
		t.Errorf("Expected %d queues, got %d", numGoroutines, len(qm.Queues))
	}

	// Concurrent queue deletion
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			queueID := fmt.Sprintf("queue-ConcurrentQueue%d", id)
			code := qm.DeleteQueue(queueID)
			if code != queue.OK {
				t.Errorf("Failed to delete concurrent queue %d", id)
			}
		}(i)
	}

	wg.Wait()

	// Verify all queues were deleted
	if len(qm.Queues) != 0 {
		t.Errorf("Expected 0 queues after deletion, got %d", len(qm.Queues))
	}
}

func TestMixedConcurrentOperations(t *testing.T) {
	config := queue.QueueManagerConfig{
		Name: "TestManager",
	}
	qm := queue.NewQueueManager(config)

	// Pre-create some queues
	for i := 0; i < 3; i++ {
		queueConfig := queue.QueueConfig{
			Name:              fmt.Sprintf("MixedQueue%d", i),
			Type:              queue.QueueTypeFIFO,
			RetentionPeriod:   time.Hour,
			VisibilityTimeout: time.Minute,
			MaxReceiveCount:   3,
			MaxMessageSize:    1024,
		}
		qm.CreateQueue(queueConfig)
	}

	var wg sync.WaitGroup
	numOperations := 50

	wg.Add(numOperations)
	for i := 0; i < numOperations; i++ {
		go func(id int) {
			defer wg.Done()

			queueID := fmt.Sprintf("queue-MixedQueue%d", id%3)

			switch id % 4 {
			case 0: // Send message
				message := queue.Message{
					ID:   fmt.Sprintf("mixed-msg-%d", id),
					Body: fmt.Sprintf("Mixed message %d", id),
				}
				req := queue.Request{Message: message}
				qm.SendMessage(queueID, req)
			case 1: // Peek message
				qm.PeekMessage(queueID)
			case 2: // Delete message
				qm.DeleteMessage(queueID, queue.Request{})
			case 3: // Create new queue (might fail if exists)
				newQueueConfig := queue.QueueConfig{
					Name:              fmt.Sprintf("NewMixedQueue%d", id),
					Type:              queue.QueueTypeFIFO,
					RetentionPeriod:   time.Hour,
					VisibilityTimeout: time.Minute,
					MaxReceiveCount:   3,
					MaxMessageSize:    1024,
				}
				qm.CreateQueue(newQueueConfig)
			}
		}(i)
	}

	wg.Wait()

	// Verify queue manager is still functional
	testQueueConfig := queue.QueueConfig{
		Name:              "FinalTestQueue",
		Type:              queue.QueueTypeFIFO,
		RetentionPeriod:   time.Hour,
		VisibilityTimeout: time.Minute,
		MaxReceiveCount:   3,
		MaxMessageSize:    1024,
	}

	code := qm.CreateQueue(testQueueConfig)
	if code != queue.OK {
		t.Error("Queue manager not functional after mixed concurrent operations")
	}
}
