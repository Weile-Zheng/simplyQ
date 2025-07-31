# SimplyQ

SimplyQ is a scalable and reliable distributed message queue built on top of [Hashicorp's implementation of Raft](https://github.com/hashicorp/raft). 

## Features

You are able to create multiple queues each handling messages independently in parallel, with configurable parameters matching that of a standard message queue service such as receive/retry count, retention period, and visibility timeout. 



## Todos
### Queue
- Support Multiple Queue
- Standard & FIFO

### Constraints:
- Max message size
- Max receive count
- VisibilityTimeout
- Retention Period

### Persistency
- Append Only File
- Snapshotting
- Log compaction
- Lazy Cleanup

### Disaster Recovery
- Replication

### Non-Functional
- High concurrency