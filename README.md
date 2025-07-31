# SimplyQ

SimplyQ is a scalable and reliable distributed message queue built on top of Hashicorp's implementation of Raft. 


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