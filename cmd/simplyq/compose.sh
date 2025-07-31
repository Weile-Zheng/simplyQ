#!/bin/bash
# filepath: /Users/weilezheng/CompSci/simplyQ/cmd/simplyq/compose.sh

# Default values
NODE_COUNT=3
BASE_PORT=10000
HTTP_BASE_PORT=8080
DATA_DIR="./raftdata"
COMMAND="start"

# Print usage information
usage() {
    echo "Usage: $0 [options] [start|stop|clean]"
    echo "Options:"
    echo "  -n, --nodes NUMBER    Number of nodes in the cluster (default: 3)"
    echo "  -p, --port PORT       Base port for Raft communication (default: 10000)"
    echo "  -h, --http-port PORT  Base port for HTTP API (default: 8080)"
    echo "  -d, --data-dir DIR    Directory for storing data (default: ./data)"
    echo "Commands:"
    echo "  start                 Start the cluster (default)"
    echo "  stop                  Stop the cluster"
    echo "  clean                 Clean up data directories"
    echo "  status                Check status of the cluster"
    echo "  join                  Join nodes to the leader"
    exit 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--nodes)
            NODE_COUNT="$2"
            shift 2
            ;;
        -p|--port)
            BASE_PORT="$2"
            shift 2
            ;;
        -h|--http-port)
            HTTP_BASE_PORT="$2"
            shift 2
            ;;
        -d|--data-dir)
            DATA_DIR="$2"
            shift 2
            ;;
        start|stop|clean|status|join)
            COMMAND="$1"
            shift
            ;;
        *)
            usage
            ;;
    esac
done

# Ensure data directory exists
mkdir -p "$DATA_DIR"

# Function to get PID file path
get_pid_file() {
    local node_id="$1"
    echo "/tmp/simplyq_node${node_id}.pid"
}

# Function to start a node
start_node() {
    local node_id="$1"
    local raft_port="$2"
    local http_port="$3"
    local peers="$4"
    local node_data_dir="${DATA_DIR}/node${node_id}"
    
    mkdir -p "$node_data_dir"

    go build -o "${node_data_dir}/simplyq" ../../cmd/simplyq/main.go

    echo "Starting node${node_id} (Raft port: ${raft_port}, HTTP port: ${http_port})"
    
    # Run the node and save PID
    NODE_ID="node${node_id}" \
    BIND_ADDR="127.0.0.1" \
    RAFT_PORT="${raft_port}" \
    HTTP_PORT="${http_port}" \
    DATA_DIR="${node_data_dir}" \
    PEERS="${peers}" \
    "${node_data_dir}/simplyq" > "${node_data_dir}/node.log" 2>&1 &
    
    echo $! > "$(get_pid_file "$node_id")"
    echo "Node${node_id} started with PID $!"
}

# Function to stop a node
stop_node() {
    local node_id="$1"
    local pid_file="$(get_pid_file "$node_id")"
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        echo "Stopping node${node_id} (PID: ${pid})"
        kill -TERM "$pid" 2>/dev/null || true
        rm -f "$pid_file"
    else
        echo "Node${node_id} is not running"
    fi
}

# Function to check cluster status
check_status() {
    echo "Checking cluster status..."
    for i in $(seq 1 "$NODE_COUNT"); do
        local http_port=$((HTTP_BASE_PORT + i - 1))
        echo "Node${i} status:"
        curl -s "http://127.0.0.1:${http_port}/raft/status" || echo "Node${i} is not responding"
        echo
    done
}

# Function to join nodes to the leader
join_nodes() {
    echo "Joining nodes to the leader..."
    local leader_port=$HTTP_BASE_PORT
    
    # Wait for leader to be ready
    echo "Waiting for leader to be ready..."
    while ! curl -s "http://127.0.0.1:${leader_port}/ping" > /dev/null; do
        sleep 1
    done
    
    # Join other nodes to the leader
    for i in $(seq 2 "$NODE_COUNT"); do
        local raft_port=$((BASE_PORT + i - 1))
        echo "Joining node${i} to the cluster..."
        curl -X POST "http://127.0.0.1:${leader_port}/raft/join?id=node${i}&address=127.0.0.1:${raft_port}"
        echo
        sleep 1
    done
}

case "$COMMAND" in
    start)
        echo "Starting SimplyQ cluster with $NODE_COUNT nodes..."
        
        # Start nodes
        for i in $(seq 1 "$NODE_COUNT"); do
            raft_port=$((BASE_PORT + i - 1))
            http_port=$((HTTP_BASE_PORT + i - 1))
            
            # For first node, no peers
            if [ "$i" -eq 1 ]; then
                peers=""
            else
                # For other nodes, set the first node as peer
                peers="127.0.0.1:$BASE_PORT"
            fi
            
            start_node "$i" "$raft_port" "$http_port" "$peers"
            sleep 2
        done
        
        echo "All nodes started. Use '$0 join' to join nodes to the cluster."
        ;;
        
    stop)
        echo "Stopping SimplyQ cluster..."
        for i in $(seq 1 "$NODE_COUNT"); do
            stop_node "$i"
        done
        echo "All nodes stopped."
        ;;
        
    clean)
        echo "Cleaning up data directories..."
        # First stop all nodes
        for i in $(seq 1 "$NODE_COUNT"); do
            stop_node "$i"
        done
        # Then remove data directories
        rm -rf "${DATA_DIR}/node"*
        echo "Data directories cleaned."
        ;;
        
    status)
        check_status
        ;;
        
    join)
        join_nodes
        ;;
        
    *)
        usage
        ;;
esac

echo "Done."