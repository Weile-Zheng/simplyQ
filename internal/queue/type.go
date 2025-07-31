package queue

type QueueType string

const (
	QueueTypeStandard QueueType = "standard"
	QueueTypeFIFO     QueueType = "fifo"
)

type OpType int

const (
	INSERT OpType = iota
	PEEK
	DELETE
	REQUEUE
	GET_CONFIG
	UPDATE_CONFIG
)

type Code int

const (
	OK Code = iota
	EMPTY_QUEUE
	EMPTY_DEAD_LETTER_QUEUE
	QUEUE_NOT_FOUND
	QUEUE_ALREADY_EXISTS
)

