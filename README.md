# Out-of-band GPU Exporter
This is a simple out-of-band GPU exporter for [Prometheus](https://prometheus.io). The exporter uses the Redfish API to collect information and it supports the regular `/metrics` endpoint to expose metrics from the host passed via the `target` parameter. For example, to scrape metrics from a Redfish instance on the IP address `192.168.1.1` call the following URL address.

```text
http://localhost:9347/metrics?target=192.168.1.1
```

Every time the exporter is called with a new target, it tries to establish a connection to the Redfish API. If the target is unreachable or if the authentication fails, the status code 500 is returned together with an error message.


## Installation
The exporter is written in [Go](https://golang.org) and it can be downloaded and compiled using:

```sh
git clone https://github.com/firmus-public/oob_gpu_exporter
cd oob_gpu_exporter
make
```

### Docker
There is a `Dockerfile` in the repository for building a container image. To build it locally use:

```sh
docker build -t oob_gpu_exporter .
```

Remember to set the listen address to `0.0.0.0` when running inside a container.

## Configuration
There are manya number of [configuration options](sample-config.yml) for the exporter, but most importantly you need to provide a username and password for all remote hosts. By default, the exporter looks for the configuration file in `/etc/oob_gpu_exporter/config.yml` but the path can be specified using the `-config` option.

```yaml
address: 127.0.0.1 # Listen address
port: 9348         # Listen port
timeout: 10        # HTTP timeout (in seconds) for Redfish API calls
hosts:
  default:
    username: user
    password: pass
  192.168.1.1:
    username: user
    password: pass
metrics:
  all: true
```

As shown in the above example, under `hosts` you can specify login information for individual hosts via their IP address or hostname, otherwise the exporter will attempt to use the login information under `default`. The login user only needs read-only permissions. Under `metrics` you can select what kind of metrics that should be returned.

**For a detailed description of the configuration, please see the [sample-config.yml](sample-config.yml) file. In this file you can also find the corresponding environment variables for the different configuration options.**


## List of Metrics
The exporter can expose the metrics described below. For each metric you can see the name and the associated labels.

```text
oob_gpu_exporter_build_info{goversion,revision,version}
oob_gpu_exporter_scrape_errors_total
oob_gpu_num_gpus
oob_gpu_bandwidth_percent{id}
oob_gpu_board_power_supply_status{id,status}
oob_gpu_consumed_power_watt{id}
oob_gpu_health{id,status}
oob_gpu_info{id,manufacturer,model,part_number,serial_number,uuid}
oob_gpu_memory_bandwidth_percent{id}
oob_gpu_memory_operating_speed_mhz{id}
oob_gpu_memory_temperature_celsius{id}
oob_gpu_operating_speed_mhz{id}
oob_gpu_power_brake_status{id,status}
oob_gpu_primary_gpu_temperature_celsius{id}
oob_gpu_state{id,state}
oob_gpu_thermal_alert_status{id,status}
```

## Endpoints
The exporter currently has three different endpoints.

| Endpoint     | Parameters | Description                                         |
| ------------ | ---------- | --------------------------------------------------- |
| `/metrics`   | `target`   | Metrics for the specified target                    |
| `/reset`     | `target`   | Reset internal state for the specified target       |
| `/health`    |            | Returns http status 200 and nothing else            |


## Prometheus Configuration
For the situation where you have a single `oob_gpu_exporter` and multiple hosts to query, the following `prometheus.yml` snippet can be used. Here `192.168.1.1` and `192.168.1.2` are the hosts to query, and `exporter:9348` is the address and port where `oob_gpu_exporter` is running.

```yaml
scrape_configs:
  - job_name: oob_gpu_exporter
    static_configs:
      - targets: ['192.168.1.1', '192.168.1.2']
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: exporter:9347
```
