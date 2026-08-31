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

create_vm() {
    local name=$1
    local template=$2

    if output=$("$limactl" list --format "{{.Name}} {{.Status}}" 2>/dev/null | grep "$name"); then
        read -r _ status <<<"$output"
        if [[ $status != "Running" ]]; then
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

getcmd() {
    if ! command -v "$1" 2>/dev/null; then
        echo "error: $1 is not installed or not available in path" >&2
        exit 1
    fi
}

limactl=$(getcmd limactl)
jq=$(getcmd jq)

if ! "$limactl" network list --json | "$jq" -r .name | grep "$NETWORK_NAME" &>/dev/null; then
    echo "Creating VM network $NETWORK_NAME ($NETWORK_GATEWAY)"
    "$limactl" network create "$NETWORK_NAME" --gateway "$NETWORK_GATEWAY"
fi

create_vm "$NFS_HOST_NAME" "$NFS_HOST_TEMPLATE" &

for i in $(seq 1 "$SWARM_NODES"); do
    create_vm "${SWARM_NODE_NAME_PREFIX}${i}" "$SWARM_NODE_TEMPLATE" &
done

wait

echo "Done."
