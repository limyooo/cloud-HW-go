# cloud-HW-go

![CI](https://github.com/limyooo/cloud-HW-go/actions/workflows/ci.yml/badge.svg)

A small REST API written in Go, packaged with Docker, and tested automatically
by a GitHub Actions workflow that posts its results to Slack.

## Project structure

```
.
├── .github/workflows/ci.yml   # CI workflow: tests in Docker + Slack summary
├── Dockerfile.api             # image that runs the API until stopped
├── Dockerfile.test            # image that runs the tests and exits
├── run-api.sh                 # builds and runs the API container
├── run-tests.sh               # builds and runs the test container
├── go.mod                     # module + Go version (dependency management)
├── main.go                    # starts the HTTP server, graceful shutdown
├── handlers.go                # HTTP routes and handlers
├── handlers_test.go           # endpoint tests (status codes, bodies)
├── store.go                   # thread-safe in-memory item store
└── store_test.go              # unit tests for the store
```

## API

The API manages items of the form `{"id": 1, "name": "pen", "price": 1.5}`.

| Method | Path          | Success          | Errors                                 |
|--------|---------------|------------------|----------------------------------------|
| GET    | `/health`     | `200 OK`         |                                        |
| GET    | `/items`      | `200 OK`         |                                        |
| GET    | `/items/{id}` | `200 OK`         | `400` invalid id, `404` not found      |
| POST   | `/items`      | `201 Created`    | `400` invalid body                     |
| PUT    | `/items/{id}` | `200 OK`         | `400` invalid id/body, `404` not found |
| DELETE | `/items/{id}` | `204 No Content` | `400` invalid id, `404` not found      |

Any other method on a known path returns `405 Method Not Allowed`. A valid body
needs a non-empty `name` and a `price` of 0 or more; unknown fields are rejected.

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (running)
- Git
- Optional, for running without Docker: [Go](https://go.dev/dl/) matching `go.mod`

## Running the API

```bash
./run-api.sh
```

This builds `Dockerfile.api` and starts the API on <http://localhost:8080>. It
keeps running until you press **Ctrl+C** (or run `docker stop go-rest-api`).
Use `PORT=9090 ./run-api.sh` to choose a different host port.

Try it from a second terminal:

```bash
curl -i -X POST localhost:8080/items -d '{"name":"pen","price":1.5}'      # 201
curl -i localhost:8080/items                                             # 200
curl -i localhost:8080/items/1                                           # 200
curl -i -X PUT localhost:8080/items/1 -d '{"name":"blue pen","price":2}'  # 200
curl -i -X DELETE localhost:8080/items/1                                 # 204
curl -i localhost:8080/items/1                                           # 404
```

## Running the tests

```bash
./run-tests.sh
echo $?   # 0 = all tests passed, non-zero = a test failed
```

This builds `Dockerfile.test`, runs `go vet` and `go test -v` inside the
container, and exits with the tests' status code.

Without Docker: `go test -v ./...`

## Test-driven development

The endpoint tests in `handlers_test.go` were written first and failed (the
server did not exist yet). `handlers.go` was then written to make them pass.
Each test asserts the expected status code, and where relevant the response
body and headers. Tests use Go's `net/http/httptest`, so no real server or
network port is needed.

## Continuous integration

`.github/workflows/ci.yml` runs on every `push` and `pull_request`, and can be
run manually from **Actions → CI → Run workflow** (`workflow_dispatch`).

1. **Run tests in Docker** runs `./run-tests.sh`, so CI uses exactly the same
   environment as every developer's machine, and writes a summary of the results
   to the run's summary page.
2. **Post summary to Slack** always runs afterwards and posts the result
   (passed / failed), branch, commit, trigger, test counts, and a link to the run.

### Slack setup

1. Create a Slack app at <https://api.slack.com/apps> (**Blank app**), enable
   **Incoming Webhooks**, and add a webhook for the target channel.
2. In the GitHub repository go to **Settings → Secrets and variables → Actions**
   and add a repository secret named `SLACK_WEBHOOK_URL` containing the webhook URL.

If the secret is not set, the Slack step is skipped with a warning instead of
failing the workflow.
