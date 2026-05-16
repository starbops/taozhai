# AGENTS.md

This file provides context and guidelines for AI coding agents working on the **taozhai** repository.

## Project Overview

`taozhai` is a Kubernetes-native CLI tool written in Go that detects IP address hijackers in bare-metal environments and reclaims the IP by powering off the offending server via its Baseboard Management Controller (BMC).

### The 3-Phase Workflow

1. **MAC Address Discovery**: Spawns a transient Kubernetes pod (Alpine image configured with a `macvlan` network annotation) to `ping` the target IP and parse the ARP table (`ip neigh`) to discover the MAC address.
2. **Inventory Search**: Queries Harvester Seeder `Inventory` Custom Resources (CRs) using a dynamic Kubernetes client to match the MAC address and identify the hardware node.
3. **BMC Power Control**: Issues a power-off command by creating a Tinkerbell Rufio BMC `Job` CR targeting the identified machine. 

## Repository Structure

- `cmd/taozhai/`: CLI application entry point (`main.go`). Handles CLI flags and workflow orchestration.
- `pkg/bmc/`: Tinkerbell Rufio BMC Job management (`bmc.tinkerbell.org/v1alpha1`).
- `pkg/client/`: Kubernetes client initialization, wrapping both `kubernetes.Clientset` and `dynamic.Interface`.
- `pkg/discovery/`: IP-to-MAC resolution via temporary discovery K8s pods.
- `pkg/inventory/`: Harvester Seeder Inventory CR queries (`metal.harvesterhci.io/v1alpha1`).
- `pkg/pod/`: K8s pod lifecycle management wrapper (create, wait, fetch logs, delete).
- `pkg/version/`: Build-time version variable injection.
- `internal/testutil/`: Testing utilities, including fake dynamic clients and object fixtures.
- `docs/`: Dependency notes and architectural documentation.

## Build and Development

The repository uses a `Makefile` for standard development tasks:

- **Build binary**: `make build` (Outputs to `bin/taozhai`)
- **Run tests**: `make test` or `make test-unit`
- **Code formatting**: `make fmt` and `make vet`
- **Linting**: `make lint` (Requires `golangci-lint`)
- **Full pre-commit check**: `make check` (Runs fmt, vet, lint, and tests)

## Guidelines for AI Agents

1. **Unstructured K8s Clients**: `taozhai` intentionally uses the Kubernetes `DynamicClient` and `unstructured.Unstructured` types to interact with external Harvester and Tinkerbell CRDs. **Do not** import the upstream Go types for these CRDs. This deliberate architectural choice prevents Go module version conflicts and keeps the binary lean.
2. **Error Handling**: Use wrapped errors with `fmt.Errorf("...: %w", err)` to preserve the original error chain.
3. **Testing Standard**: The project uses `Ginkgo` and `Gomega` for its test suites. When adding or modifying logic in `pkg/`, ensure that corresponding unit tests are updated or created. Use the fake dynamic K8s client defined in `internal/testutil` to mock Kubernetes interactions.
4. **Clean Code**: Follow standard idiomatic Go conventions. All code must pass `make lint` and `make vet` cleanly.
5. **No Dangerous Automations Without Prompts**: The CLI requires explicit confirmation before powering off hardware unless the `--force` flag is provided. When modifying workflow logic, always respect the dry-run default.
