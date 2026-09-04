# swarm-test

A self-contained playground for a multi-node **Docker Swarm** cluster running
entirely in local VMs. [`setup.sh`](./setup.sh) spins up the VMs with
[Lima](https://lima-vm.io/), forms a Swarm, provisions an NFS server for shared
volume storage, and deploys a small demo application
([`app/`](./app), the *Guestbook*) as a Swarm stack.

The goal is to have a realistic, throw-away environment to experiment with Swarm
services, rolling updates, and shared (NFS-backed) volumes without touching a
cloud provider or the host's Docker daemon.

## Topology

| VM                | Template               | Role                                              |
| ----------------- | ---------------------- | ------------------------------------------------- |
| `nfs-host`        | `template:ubuntu`      | NFS server, exports `/media/nfs` to the cluster.  |
| `swarm-node-1`    | `template:docker-rootful` | First Swarm manager (runs `docker swarm init`). |
| `swarm-node-2`    | `template:docker-rootful` | Additional Swarm manager (joins node 1).         |

All VMs are attached to a dedicated Lima network `swarm-test`
(`192.168.69.0/24`, gateway `192.168.69.1`). Each VM is provisioned with 2
CPUs, 2 GB RAM, and a 16 GB disk, and mounts the repository directory (read
only).

The demo `guestbook` stack runs 2 replicas behind a shared NFS volume so every
replica serves the same guestbook entries.

## Prerequisites

Install these on the host (macOS or Linux):

- [Lima](https://lima-vm.io/) — provides the `limactl` / `lima` CLI. On macOS:
  `brew install lima`.
- A working VM backend for Lima (on macOS the bundled QEMU/`vz` backend is
  fine; on Linux QEMU is used).
- `bash` (the script uses `set -eo pipefail` and bash arrays) plus standard
  UNIX tools (`awk`, `seq`, `ip` inside the guests).
- Enough free resources for 3 VMs: ~6 CPUs, ~6 GB RAM, ~48 GB disk.

The Lima templates `template:ubuntu` and `template:docker-rootful` are shipped
with Lima and are fetched automatically on first use, so an internet connection
is required for the first run. The demo image
`docker.io/zekro/demo-guestbook` is pulled from Docker Hub by the Swarm.

## What `setup.sh` does

Run from the repository root:

```sh
./setup.sh
```

Step by step:

1. **Checks tooling.** Verifies `limactl` is on `PATH`, aborting otherwise. All
   `limactl` output is appended to `limactl.log`; if any Lima command fails the
   script stops and points at that log.
2. **Creates the network.** `lima network create swarm-test` with gateway
   `192.168.69.1/24`.
3. **Creates the VMs in parallel:**
   - `create_nfs_host` creates and starts `nfs-host` from `template:ubuntu`.
   - `swarm-node-1` and `swarm-node-2` are each created and started from
     `template:docker-rootful`.
   - Every VM is created with `lima create` (2 CPUs / 2 GB / 16 GB), the repo
     directory mounted via `--mount-only`, and attached to the `swarm-test`
     Lima network, then started with `lima start`.
   The script waits for all background VM creations to finish.
4. **Provisions the NFS host.** Inside `nfs-host`:
   - installs `nfs-kernel-server`,
   - creates `/media/nfs`, owned by `nobody:nogroup`,
   - appends an export line for `/media/nfs` to `/etc/exports` allowing the
     `192.168.69.0/24` network (`rw,sync,no_subtree_check`),
   - runs `exportfs -arv` to publish the export.
5. **Initializes the Swarm.** Runs `docker swarm init` on `swarm-node-1` and
   determines that VM's IP address on the `swarm-test` network.
6. **Joins the remaining nodes.** For each further node (`swarm-node-2`), fetches
   a *manager* join token from `swarm-node-1` and runs `docker swarm join`
   against `<swarm-node-1-ip>:2377`. (All nodes join as managers in this demo.)
7. **Deploys the demo stack.** Determines the NFS host's IP, creates
   `/media/nfs/guestbook` on `nfs-host` (owned by `nobody:nogroup`), then on
   `swarm-node-1` runs:

   ```sh
   NFS_ADDRESS=<nfs-host-ip> NFS_SHARE=/media/nfs \
     docker stack deploy -c app/stack.yml guestbook
   ```

   [`app/stack.yml`](./app/stack.yml) defines a single `guestbook` service (2
   replicas, `start-first` rolling updates, restart on failure, port `8080`
   published) with a `data` volume whose driver options mount
   `:${NFS_SHARE}/guestbook` from `${NFS_ADDRESS}` over NFSv4.

### Tearing down

```sh
./setup.sh --shutdown
```

This stops and deletes `nfs-host` and every `swarm-node-*` VM (in parallel),
then removes the `swarm-test` Lima network.

## The demo app

[`app/`](./app) contains *Guestbook*, a small Go web application: visitors sign a
guestbook and see all entries on a self-refreshing HTML page. An optional
activity simulator posts synthetic entries so a running cluster always shows
traffic. Entries are stored as one JSON file per entry in
`GUESTBOOK_STORAGE_DIR` (`/data` in the container, backed by the NFS volume).

See [`app/README.md`](./app/README.md) for the API, configuration variables, and
how to run it locally with `go run ./cmd/guestbook`.

### Trying the demo

After `./setup.sh` finishes:

1. Find a Swarm node's address on the `swarm-test` network, e.g.:

   ```sh
   limactl shell swarm-node-1 ip -4 -o addr show
   ```

2. Open `http://<swarm-node-ip>:8080` in a browser. Because the port is
   published on the Swarm routing mesh, any node's IP works.
3. Inspect the running service:

   ```sh
   limactl shell swarm-node-1 docker service ls
   limactl shell swarm-node-1 docker stack ps guestbook
   ```

4. Redeploy after changing `app/stack.yml`:

   ```sh
   limactl shell swarm-node-1 \
     env NFS_ADDRESS=<nfs-host-ip> NFS_SHARE=/media/nfs \
     docker stack deploy -c app/stack.yml guestbook
   ```
