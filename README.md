# Terraform Provider for Flespi

A Terraform provider for managing resources on the [Flespi](https://flespi.com) telematics platform.

Built with the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)

## Provider Configuration

```hcl
terraform {
  required_providers {
    flespi = {
      source = "mixser/flespi"
    }
  }
}

provider "flespi" {
  token = var.flespi_token  # or set FLESPI_TOKEN env var
}
```

The provider requires a Flespi master token. You can create tokens in the [Flespi panel](https://flespi.io).

## Resources

### Gateway

| Resource | Description |
|----------|-------------|
| `flespi_device` | GPS/telematics device |
| `flespi_channel` | Protocol channel (by protocol ID or name) |
| `flespi_stream` | Message stream |
| `flespi_geofence` | Geofence zone (circle, polygon, or corridor) |

### Platform

| Resource | Description |
|----------|-------------|
| `flespi_webhook` | Webhook (single or chained) |
| `flespi_token` | API access token |
| `flespi_subaccount` | Sub-account |
| `flespi_limit` | Resource usage limit set |

### Storage

| Resource | Description |
|----------|-------------|
| `flespi_cdn` | CDN storage bucket |

## Example Usage

```hcl
# Create a device type limit set
resource "flespi_limit" "standard" {
  name         = "standard-plan"
  devices_count = 100
  channels_count = 10
}

# Create a sub-account with that limit
resource "flespi_subaccount" "tenant" {
  name     = "tenant-a"
  limit_id = flespi_limit.standard.id
}

# Create a channel
resource "flespi_channel" "gps" {
  name          = "gps-channel"
  enabled       = true
  protocol_name = "Teltonika"
}

# Create a device
resource "flespi_device" "tracker" {
  name           = "vehicle-tracker-01"
  enabled        = true
  device_type_id = 1
}

# Create a stream
resource "flespi_stream" "kafka" {
  name        = "kafka-stream"
  protocol_id = 7
  enabled     = true
  queue_ttl   = 86400
}

# Create a webhook
resource "flespi_webhook" "notify" {
  name = "event-webhook"
  type = "single-webhook"

  triggers = [
    {
      topic = "gw/devices/+/messages"
    }
  ]

  configurations = [
    {
      type    = "custom-server"
      uri     = "https://example.com/webhook"
      method  = "POST"
      body    = "{\"device\": \"{#device_id}\"}"
      headers = []
    }
  ]
}
```

## Building from Source

```shell
git clone https://github.com/mixser/terraform-provider-flespi
cd terraform-provider-flespi
go build ./...
```

To install locally for development, add a [dev override](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) to your `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "mixser/flespi" = "/path/to/your/GOPATH/bin"
  }
  direct {}
}
```

Then run:

```shell
go install .
```

## Developing the Provider

```shell
# Build
go build ./...

# Generate docs
go generate ./...
```
