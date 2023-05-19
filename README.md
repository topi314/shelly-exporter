[![Go Report](https://goreportcard.com/badge/github.com/TopiSenpai/shelly-exporter)](https://goreportcard.com/report/github.com/TopiSenpai/shelly-exporter)
[![Go Version](https://img.shields.io/github/go-mod/go-version/TopiSenpai/shelly-exporter)](https://golang.org/doc/devel/release.html)
[![KittyBot License](https://img.shields.io/github/license/TopiSenpai/shelly-exporter)](LICENSE)
[![KittyBot Version](https://img.shields.io/github/v/tag/TopiSenpai/shelly-exporter?label=release)](https://github.com/TopiSenpai/shelly-exporter/releases/latest)
[![Docker](https://github.com/TopiSenpai/shelly-exporter/actions/workflows/docker.yml/badge.svg)](https://github.com/TopiSenpai/shelly-exporter/actions/workflows/docker.yml)
[![Discord](https://discordapp.com/api/guilds/608506410803658753/embed.png?style=shield)](https://discord.gg/sD3ABd5)


# shelly-exporter

Prometheus exporter for Shelly Plug S devices.

## Installation

You can either run the exporter directly or use the provided Docker image.

### Docker-Compose

```yaml
version: '3.7'

services:
  prometheus:
    image: prom/prometheus:latest
    ...
  shelly-exporter:
    image: ghcr.io/topisenpai/shelly-exporter:master
    container_name: shelly-exporter
    restart: unless-stopped
    volumes:
      - ./shelly.yml:/var/lib/shelly-collector/config.yml
    expose:
      - 2112
```

## Configuration

The exporter is configured via a YAML file. The default path is `/etc/shelly-exporter/config.yml` but you can change it
with the `--config` flag.

```yaml
# shelly.yml.example
```

## Metrics

Shelly Exporter exposes the following metrics:

| Metric                   | Description                                                               | Labels                    |
|--------------------------|---------------------------------------------------------------------------|---------------------------|
| `shellyplug_power`       | Current real AC power being drawn, in Watts                               | `serial`, `name`, `meter` |
| `shellyplug_power_valid` | Whether power metering self-checks OK (0.0 = not OK, 1.0 = OK)            | `serial`, `name`, `meter` |
| `shellyplug_overpower`   | Value in Watts, on which an overpower condition is detected               | `serial`, `name`, `meter` |
| `shellyplug_total_power` | Total energy consumed by the attached electrical appliance in Watt-minute | `serial`, `name`, `meter` |
| `shellyplug_temperature` | PlugS only internal device temperature in °C                              | `serial`, `name`          |
| `shellyplug_uptime`      | Seconds elapsed since boot                                                | `serial`, `name`          |
| `shellyplug_has_update`  | Whether an update is available (0.0 = no update, 1.0 = update)            | `serial`, `name`          |

## Details

The exporter uses the Shelly API to get the current power consumption of the devices. You can find more information [here](https://shelly-api-docs.shelly.cloud/gen1/#status) and [here](https://shelly-api-docs.shelly.cloud/gen1/#shelly-plug-plugs-status)

## License

Shelly Exporter is licensed under the [Apache License 2.0](/LICENSE).

---

## Contributing

Contributions are always welcome! Just open a pull request or discussion and I will take a look at it.

## Contact

- [Discord](https://discord.gg/sD3ABd5)
- [Twitter](https://twitter.com/TopiSenpai)
- [Email](mailto:git@topi.wtf)
