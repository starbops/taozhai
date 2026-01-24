# Dependency Notes

This document tracks deprecated dependencies and unstable API dependencies used by taozhai.

## Deprecated Dependencies

### github.com/gogo/protobuf v1.3.2

**Status**: Deprecated indirect dependency (transitive via Kubernetes client libraries)

**Why it exists**: The `k8s.io/apimachinery` package depends on gogo/protobuf for
protocol buffer serialization.

**Risk Assessment**: LOW

- This is an indirect dependency; taozhai does not directly use gogo/protobuf
- Kubernetes upstream is actively working on migration (KEP #5589)
- No action required from taozhai maintainers

**Monitoring Strategy**:

1. Track Kubernetes releases for gogo/protobuf removal
2. When updating `k8s.io/client-go`, run `go mod tidy` to check if gogo/protobuf
   is still required
3. Periodically check <https://github.com/kubernetes/enhancements/issues/5589>

**References**:

- <https://github.com/gogo/protobuf> (deprecated notice)
- <https://github.com/kubernetes/kubernetes/issues/96564>
- <https://github.com/kubernetes/enhancements/issues/5589>
- <https://github.com/kubernetes/kubernetes/pull/128653>

**Last Reviewed**: 2026-01-24

## External API Dependencies

### metal.harvesterhci.io/v1alpha1 - Inventory

**Provider**: Harvester Seeder
**Stability**: Alpha (v1alpha1)
**Used in**: `pkg/inventory/finder.go`

The Inventory CR represents a discovered bare-metal machine. This API is part of
the Harvester Seeder project and may change without notice.

**Monitoring**: Check Harvester releases for API version changes.
**Upstream**: <https://github.com/harvester/seeder>

### bmc.tinkerbell.org/v1alpha1 - Job

**Provider**: Tinkerbell Rufio
**Stability**: Alpha (v1alpha1)
**Used in**: `pkg/bmc/manager.go`

The BMC Job CR allows executing out-of-band management operations (power control,
boot device selection). This API is part of the Tinkerbell Rufio project.

**Monitoring**: Check Tinkerbell releases for API version changes.
**Upstream**: <https://github.com/tinkerbell/rufio>
