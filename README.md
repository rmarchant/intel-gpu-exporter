THIS IS A FORK WITH BASIC CODE CHANGES TO COLLECT FROM SRIOV DEVICE on PROXMOX.

Not meant to run in Docker on proxmox host, as running docker on proxmox host is not recommended and could break the networking of the host.

Instead, simply compile the go program, and then execute on the host to expose the metrics.

# intel-gpu-exporter
[![release](https://img.shields.io/github/v/tag/clambin/intel-gpu-exporter?color=green&label=release&style=plastic)](https://github.com/clambin/intel-gpu-exporter/releases)
[![codecov](https://img.shields.io/codecov/c/gh/clambin/intel-gpu-exporter?style=plastic)](https://app.codecov.io/gh/clambin/intel-gpu-exporter)
[![build](https://github.com/clambin/intel-gpu-exporter/workflows/build/badge.svg)](https://github.com/clambin/intel-gpu-exporter/actions)
[![go report card](https://goreportcard.com/badge/github.com/clambin/intel-gpu-exporter)](https://goreportcard.com/report/github.com/clambin/intel-gpu-exporter)
[![license](https://img.shields.io/github/license/clambin/intel-gpu-exporter?style=plastic)](LICENSE.md)

Exports GPU statistics for Intel Quick Sync Video GPUs. 
# Metrics


| metric | type |  labels | help                                               |
| --- | --- |  --- |----------------------------------------------------|
| gpumon_clients_count | GAUGE | | Number of active clients (currently not supported) |
| gpumon_engine_usage | GAUGE | attrib, engine| Usage statistics for the different GPU engines     |
| gpumon_power | GAUGE | type| Power consumption by type                          |

## Authors

* **Christophe Lambin**

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
