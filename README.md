# SimplyQ

SimplyQ is a scalable and reliable distributed message queue built on top of [Hashicorp's implementation of Raft](https://github.com/hashicorp/raft). 

## Features (in progress)

You are able to create multiple queues each handling messages independently in parallel, with cooresponding DLQ and configurable parameters matching that of a standard message queue service such as receive/retry count, retention period, and visibility timeout. 