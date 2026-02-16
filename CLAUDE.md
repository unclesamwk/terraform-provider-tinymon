# terraform-provider-tinymon

Terraform provider for TinyMon. Manages hosts and checks as infrastructure as code via the TinyMon Push API.

## Project Structure

```
main.go                              Entry point (providerserver.Serve)
internal/provider/
  provider.go                        Provider config (url, api_key), TinyMonClient HTTP helper
  host_resource.go                   tinymon_host resource (CRUD via Push API)
  check_resource.go                  tinymon_check resource (CRUD via Push API)
examples/main.tf                     Example HCL configuration
.goreleaser.yml                      Cross-platform release builds
```

## Tech Stack

- Go 1.24, hashicorp/terraform-plugin-framework v1.17
- No SDKv2 -- uses the modern Plugin Framework exclusively

## Key Concepts

- **Provider config**: `url` + `api_key` (also via env `TINYMON_URL`, `TINYMON_API_KEY`)
- **TinyMonClient**: HTTP client with Bearer auth, `DoJSON()` helper for all API calls
- **tinymon_host**: Identified by `address` (ForceNew). CRUD maps to Push API: POST (upsert), GET, DELETE
- **tinymon_check**: Identified by `host_address` + `type` + `config` (all ForceNew). Same CRUD pattern
- **Upsert pattern**: Create and Update both use POST (Push API upsert behavior)
- **Import**: Host by address, check by `host_address/type` or `host_address/type/config`

## Build & Test

```bash
go build -o terraform-provider-tinymon .
```

Local development via `~/.terraformrc` dev_overrides (no registry publish needed):

```hcl
provider_installation {
  dev_overrides {
    "unclesamwk/tinymon" = "/path/to/terraform-provider-tinymon"
  }
  direct {}
}
```

With dev overrides, `terraform init` is not required.

## Versioning

- Current version: v0.0.1
- Tags: increment +0.0.1
- GitHub: unclesamwk/terraform-provider-tinymon

## Related Repos

- **TinyMon (MiniMon)**: The monitoring application itself. Push API endpoints used by this provider.
- **tinymon-operator**: K8s operator (separate use case -- this provider is for non-K8s infrastructure)
