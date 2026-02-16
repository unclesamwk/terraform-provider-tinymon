# Terraform Provider for TinyMon

Terraform provider to manage [TinyMon](https://github.com/unclesamwk/TinyMon) hosts and checks as infrastructure as code.

## Requirements

- [Go](https://go.dev/) 1.22+
- [Terraform](https://www.terraform.io/) 1.0+

## Installation (local)

Build the provider binary:

```sh
git clone https://github.com/unclesamwk/terraform-provider-tinymon.git
cd terraform-provider-tinymon
go build -o terraform-provider-tinymon .
```

Add the dev override to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "unclesamwk/tinymon" = "/path/to/terraform-provider-tinymon"
  }
  direct {}
}
```

Replace `/path/to/terraform-provider-tinymon` with the absolute path to the directory containing the built binary.

With dev overrides, `terraform init` is not required. Terraform will use the local binary directly.

## Provider Configuration

```hcl
provider "tinymon" {
  url     = "https://mon.example.com"
  api_key = var.tinymon_api_key
}
```

| Attribute | Environment Variable | Description |
|-----------|---------------------|-------------|
| `url` | `TINYMON_URL` | Base URL of the TinyMon instance |
| `api_key` | `TINYMON_API_KEY` | Bearer token for the Push API |

Both attributes can be set via environment variables instead of in the configuration.

## Resources

### tinymon_host

Manages a host in TinyMon.

```hcl
resource "tinymon_host" "webserver" {
  name        = "Webserver"
  address     = "192.168.1.10"
  description = "Main web server"
  topic       = "production/webservers"
  enabled     = true
}
```

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `address` | string | yes | | IP or hostname (forces replacement on change) |
| `name` | string | no | address | Display name |
| `description` | string | no | `""` | Description |
| `topic` | string | no | `""` | Topic path for grouping |
| `enabled` | bool | no | `true` | Whether the host is enabled |
| `id` | int | computed | | Host ID |

Import: `terraform import tinymon_host.webserver 192.168.1.10`

### tinymon_check

Manages a check for an existing host.

```hcl
resource "tinymon_check" "webserver_ping" {
  host_address = tinymon_host.webserver.address
  type         = "ping"
}

resource "tinymon_check" "webserver_http" {
  host_address     = tinymon_host.webserver.address
  type             = "http"
  interval_seconds = 60
}

resource "tinymon_check" "webserver_disk" {
  host_address = tinymon_host.webserver.address
  type         = "disk"
  config       = jsonencode({ mount = "/" })
}
```

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `host_address` | string | yes | | Host address (forces replacement) |
| `type` | string | yes | | Check type (forces replacement) |
| `config` | string | no | `"{}"` | JSON config (forces replacement) |
| `interval_seconds` | int | no | `300` | Check interval in seconds |
| `enabled` | bool | no | `true` | Whether the check is enabled |
| `id` | int | computed | | Check ID |

Check types: `ping`, `http`, `port`, `certificate`, `content`, `content_hash`, `disk`, `disk_health`, `load`, `memory`

Import: `terraform import tinymon_check.webserver_ping 192.168.1.10/ping`

For checks with config: `terraform import tinymon_check.webserver_disk '192.168.1.10/disk/{"mount":"/"}'`

## Full Example

```hcl
terraform {
  required_providers {
    tinymon = {
      source = "unclesamwk/tinymon"
    }
  }
}

provider "tinymon" {
  url     = "https://mon.example.com"
  api_key = var.tinymon_api_key
}

variable "tinymon_api_key" {
  type      = string
  sensitive = true
}

resource "tinymon_host" "nas" {
  name        = "NAS"
  address     = "192.168.1.50"
  description = "Synology DS920+"
  topic       = "home/storage"
}

resource "tinymon_check" "nas_ping" {
  host_address = tinymon_host.nas.address
  type         = "ping"
}

resource "tinymon_check" "nas_http" {
  host_address     = tinymon_host.nas.address
  type             = "http"
  interval_seconds = 60
}

resource "tinymon_check" "nas_cert" {
  host_address = tinymon_host.nas.address
  type         = "certificate"
}
```
