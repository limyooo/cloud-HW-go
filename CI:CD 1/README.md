# go-devops-course

![CI](https://github.com/YOUR-USERNAME/YOUR-REPO/actions/workflows/ci.yml/badge.svg)

A small Go project that serves as the foundation for this semester's assignments.
It contains an in-memory item store, unit tests for it, Go's dependency
management file (`go.mod`), and a GitHub Actions workflow that runs the tests
automatically on every change and on demand.

## Project structure

```
.
├── .github/workflows/ci.yml   # CI workflow
├── go.mod                     # module name + Go version (dependency management)
├── main.go                    # demo program
├── store.go                   # in-memory item store (List/Get/Create/Update/Delete)
├── store_test.go              # unit tests for the store
├── run.sh                     # runs the demo
└── test.sh                    # runs vet + tests
```

## Prerequisites

- [Go](https://go.dev/dl/) 1.23 or newer (`go version` to check)
- Git

## Running the code

```bash
./run.sh
# or, equivalently
go run .
```

Expected output:

```
Items in store:
  #1  Notebook (large)   $4.75
  #2  Pen                $1.25
```

## Running the tests

```bash
./test.sh
# or, equivalently
go test -v ./...
```

The script exits with status `0` when all tests pass and a non-zero status when
any test fails. To also run Go's race detector (requires a C compiler):

```bash
go test -race ./...
```

## Dependency management

Go uses **modules**. `go.mod` declares the module name and the Go version the
project targets. This project currently uses only the Go standard library, so
there are no third-party requirements. When a dependency is added
(`go get example.com/pkg`), Go records it in `go.mod` and writes its checksums
to `go.sum`; both files are committed so every machine builds with exactly the
same versions. `go mod tidy` removes unused dependencies.

## Continuous integration

The workflow in `.github/workflows/ci.yml` runs on:

| Trigger             | When                                                        |
|---------------------|-------------------------------------------------------------|
| `push`              | any commit pushed to any branch                             |
| `pull_request`      | a pull request is opened or updated                         |
| `workflow_dispatch` | manually: **Actions** tab → **CI** → **Run workflow**       |

Each run checks out the code, installs the Go version from `go.mod`, verifies
module checksums, runs `go vet`, and runs the tests with the race detector.
A failing test fails the run, and GitHub shows a red ✗ next to the commit.
