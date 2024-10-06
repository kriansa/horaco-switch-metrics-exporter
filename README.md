# ZX-SWTGW218AS metrics exporter

This is a simple metrics exporter for the ZX-SWTGW218AS switch. Without SNMP support, the only way
to get metrics from the switch is through the web interface. This exporter uses the web interface to
get the metrics and expose them in a prometheus format.

This is a cheap 2.5G managed switch sold on Aliexpress by various sellers under different brands and
names. You can see more information about it on [STH Forums][sth-forum].

[sth-forum]: https://forums.servethehome.com/index.php?threads/horaco-2-5gbe-managed-switch-8-x-2-5gbe-1-10gb-sfp.41571/

## Usage

Build and execute the exporter, passing the environment variables as the parameters to access the
switch:

```shell
SWITCH_URL=http://192.168.1.2 USER=admin PASS=admin switch-exporter
```

By default, it will open a web server on `http://localhost:8080/metrics`, but the binding address
can be configured by setting the `BIND_ADDRESS` environment variable.

## Features

- Pools for the following metrics every second and exposes them in a prometheus format on `/metrics`:
  - `switch_port_enabled`
  - `switch_port_connected`
  - `switch_port_speed`
  - `switch_port_packets_total`
- Uses environment variables to configure the connection for flexibility and ease of use.
- Automatically crashes if the switch is not accessible, to be used with a process supervisor such
  as systemd, docker-compose, supervisord, Kubernetes, etc.
- Distributed as a Docker image for easy deployment.

## License

All the code in this repository is licensed under the [Apache V2](LICENSE).
