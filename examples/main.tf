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

resource "tinymon_host" "webserver" {
  name        = "Webserver"
  address     = "192.168.1.10"
  description = "Main web server"
  topic       = "production/webservers"
}

resource "tinymon_check" "webserver_ping" {
  host_address = tinymon_host.webserver.address
  type         = "ping"
}

resource "tinymon_check" "webserver_http" {
  host_address     = tinymon_host.webserver.address
  type             = "http"
  interval_seconds = 60
}
