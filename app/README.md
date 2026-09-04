# Guestbook

A small demo web application for the Docker Swarm setup in this repository.
Visitors can sign a guestbook and see all entries on a single HTML page that
refreshes itself periodically. An optional activity simulator posts synthetic
entries so a running cluster always shows some traffic.

## API

All endpoints speak JSON. No authentication.

| Method | Path            | Description                                                              |
| ------ | --------------- | ---------------------------------------------------------------------- |
| `POST` | `/api/entries`  | Create an entry. Body: `{"name": "...", "message": "..."}`.             |
| `GET`  | `/api/entries`  | List all entries, oldest first. Query param `last_id` returns only entries newer than that ID. |
| `GET`  | `/`             | The static web page (`internal/web/assets/`, embedded into the binary).  |

Entries are stored as one JSON file per entry inside `GUESTBOOK_STORAGE_DIR`.
The file name is the entry ID, a time sortable [xid](https://github.com/rs/xid).

## Configuration

| Environment variable           | Default   | Description                                  |
| ------------------------------ | --------- | ------------------------------------------- |
| `GUESTBOOK_ADDRESS`            | `:8080`   | Listen address of the HTTP server.           |
| `GUESTBOOK_STORAGE_DIR`        | `./data`  | Directory the entry JSON files are stored in. |
| `GUESTBOOK_SIMULATOR_ENABLED`  | `true`    | Whether the synthetic activity simulator runs. |
| `GUESTBOOK_SIMULATOR_INTERVAL` | `5s`      | Delay between two synthetic entries.          |

## Run locally

```sh
go run ./cmd/guestbook
# open http://localhost:8080
```

## Run on Docker Swarm

```sh
docker build -t guestbook:latest .
docker stack deploy -c stack.yml guestbook
```

See the comments in `stack.yml` for pointing the `data` volume at shared
(NFS) storage so all replicas serve the same entries.
