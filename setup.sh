#!/usr/bin/env bash

set -eo pipefail

VM_NUM_CPU=2
VM_MEM_GB=2
VM_STORAGE_GB=16

NETWORK_NAME=swarm-test
NETWORK_GATEWAY=192.168.69.1/24

NFS_HOST_NAME=nfs-host
NFS_HOST_TEMPLATE=template:ubuntu
SWARM_NODE_TEMPLATE=template:docker
SWARM_NODE_NAME_PREFIX=swarm-node-
SWARM_NODES=2

vm_info() {
    local name=$1
    "$limactl" list --format "{{.Name}} {{.Status}}" 2>/dev/null | grep -E "^$name "
}

vm_exists() {
    vm_info "$1" &>/dev/null
}

vm_status() {
    local name=$1
    local output status
    output=$(vm_info "$name") || return 1
    read -r _ status <<<"$output"
    echo "$status"
}

network_exists() {
    "$limactl" network list --json | "$jq" -r .name | grep -qx "$NETWORK_NAME"
}

create_vm() {
    local name=$1
    local template=$2

    if vm_exists "$name"; then
        if [[ $(vm_status "$name") != "Running" ]]; then
            echo "Starting VM $name ($template)"
            "$limactl" start "$name" >/dev/null
        else
            echo "VM $name ($template) is already running"
        fi
        return
    fi

    echo "Creating VM $name ($template)"
    "$limactl" create \
        --name "$name" \
        --cpus "$VM_NUM_CPU" \
        --memory "$VM_MEM_GB" \
        --disk "$VM_STORAGE_GB" \
        --mount-only "$PWD" \
        --network "lima:$NETWORK_NAME" \
        --yes \
        "$template" >/dev/null

    echo "Starting VM $name ($template)"
    "$limactl" start "$name" >/dev/null
}

stop_vm() {
    local name=$1

    if ! vm_exists "$name"; then
        return
    fi

    if [[ $(vm_status "$name") == "Running" ]]; then
        echo "Stopping VM $name"
        "$limactl" stop "$name" >/dev/null
    fi
}

remove_vm() {
    local name=$1

    if ! vm_exists "$name"; then
        echo "VM $name does not exist"
        return
    fi

    stop_vm "$name"

    echo "Removing VM $name"
    "$limactl" delete "$name" >/dev/null
}

remove_network() {
    if ! network_exists; then
        echo "VM network $NETWORK_NAME does not exist"
        return
    fi

    echo "Removing VM network $NETWORK_NAME"
    "$limactl" network delete --force "$NETWORK_NAME" >/dev/null
}

shutdown_all() {
    remove_vm "$NFS_HOST_NAME" &

    for i in $(seq 1 "$SWARM_NODES"); do
        remove_vm "${SWARM_NODE_NAME_PREFIX}${i}" &
    done

    wait

    remove_network

    echo "Done."
}

getcmd() {
    if ! command -v "$1" 2>/dev/null; then
        echo "error: $1 is not installed or not available in path" >&2
        exit 1
    fi
}

limactl=$(getcmd limactl)
jq=$(getcmd jq)

if [[ "${1:-}" == "--shutdown" ]]; then
    shutdown_all
    exit 0
fi

if ! network_exists; then
    echo "Creating VM network $NETWORK_NAME ($NETWORK_GATEWAY)"
    "$limactl" network create "$NETWORK_NAME" --gateway "$NETWORK_GATEWAY"
fi

create_vm "$NFS_HOST_NAME" "$NFS_HOST_TEMPLATE" &

for i in $(seq 1 "$SWARM_NODES"); do
    create_vm "${SWARM_NODE_NAME_PREFIX}${i}" "$SWARM_NODE_TEMPLATE" &
done

wait

echo "Done."
