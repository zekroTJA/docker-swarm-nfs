#!/usr/bin/env bash

set -eo pipefail

VM_NUM_CPU=2
VM_MEM_GB=2
VM_STORAGE_GB=16

NETWORK_NAME=swarm-test
NETWORK_GATEWAY=192.168.69.1/24

LIMACTL_LOG="$PWD/limactl.log"

NFS_HOST_NAME=nfs-host
NFS_HOST_TEMPLATE=template:ubuntu
SWARM_NODE_TEMPLATE=template:docker
SWARM_NODE_NAME_PREFIX=swarm-node-
SWARM_NODES=2

create_vm() {
    local name=$1
    local template=$2

    echo "Creating VM $name ($template)"
    lima create \
        --name "$name" \
        --cpus "$VM_NUM_CPU" \
        --memory "$VM_MEM_GB" \
        --disk "$VM_STORAGE_GB" \
        --mount-only "$PWD" \
        --network "lima:$NETWORK_NAME" \
        --yes \
        "$template"

    echo "Starting VM $name ($template)"
    lima start "$name"
}

stop_vm() {
    local name=$1

    lima stop "$name"
}

remove_vm() {
    local name=$1

    stop_vm "$name"

    echo "Removing VM $name"
    lima delete "$name"
}

remove_network() {
    echo "Removing VM network $NETWORK_NAME"
    lima network delete --force "$NETWORK_NAME"
}

shutdown_all() {
    remove_vm "$NFS_HOST_NAME" &

    for i in $(seq 1 "$SWARM_NODES"); do
        remove_vm "${SWARM_NODE_NAME_PREFIX}${i}" &
    done

    wait

    remove_network

    echo "Finished."
}

getcmd() {
    if ! command -v "$1" 2>/dev/null; then
        echo "error: $1 is not installed or not available in path" >&2
        exit 1
    fi
}

limactl=$(getcmd limactl)

lima() {
    "$limactl" "$@" >>"$LIMACTL_LOG" 2>&1
}

create_nfs_host() {
    create_vm "$NFS_HOST_NAME" "$NFS_HOST_TEMPLATE"

    echo "Setting up NFS host ..."
    lima shell "$NFS_HOST_NAME" sudo apt-get install --yes nfs-kernel-server
    lima shell "$NFS_HOST_NAME" sudo apt-get install --yes nfs-kernel-server
    lima shell "$NFS_HOST_NAME" sudo mkdir -p /media/nfs
    lima shell "$NFS_HOST_NAME" sudo chown nobody:nogroup /media/nfs
    lima shell "$NFS_HOST_NAME" sudo tee -a /etc/exports <<<"/media/nfs   ${NETWORK_GATEWAY}(rw,sync,no_subtree_check)"
    lima shell "$NFS_HOST_NAME" sudo exportfs -arv
}

if [[ "${1:-}" == "--shutdown" ]]; then
    shutdown_all
    exit 0
fi

echo "Creating VM network $NETWORK_NAME ($NETWORK_GATEWAY)"
lima network create "$NETWORK_NAME" --gateway "$NETWORK_GATEWAY"

create_nfs_host &

for i in $(seq 1 "$SWARM_NODES"); do
    create_vm "${SWARM_NODE_NAME_PREFIX}${i}" "$SWARM_NODE_TEMPLATE" &
done

sleep 1
echo "Waiting for finishing creating VMs ..."
wait

echo "Finished."
