# AlertManager Webhook Compatible EC2 Operator

This app is designed to be used alongwith [AlertManager](https://prometheus.io/docs/alerting/latest/alertmanager/) ([Webhook](https://prometheus.io/docs/alerting/latest/configuration/#webhook_config) POST payload) or any direct GET requests to start or stop an EC2 instance.

Inspired from tons of wasted dollars on GPU instances where we didn't quite use GPU but kept the instance running hoping we'd need to use it in the next 30 mins. We all know those 30 mins easily become 3 hours and even 3 days.

It supports 2 APIs:
* `operateInstance` - request contains InstanceId of the EC2 instance to start/stop.
* `operateHostname` - request contains resolvable hostname. App will look for its IP address and use that in `private-ip-address` filter to get InstanceId from AWS EC2 APIs.

It also supports 2 different implementations of EC2 Client:
* `native` - Instantiates EC2 go-sdk client and uses session contexts. Useful when you have Access Key and Secret Key and aws profile configured on host machine.
* `cli` - Uses `aws ec2` CLI shell commands. Useful when your host machine has roles / permissions attached to start / stop required EC2 instances.

# Example Prometheus rule to trigger an alert
    additionalPrometheusRulesMap:
        gpu.rules:
            groups:
            - name: gpu.rules
            rules:
            - alert: UnusedGPU
                expr: max_over_time(nvidia_smi_utilization_gpu_ratio{instance="your-hostname:9835"}[5m]) * 100 < 2
                labels:
                    severity: critical
                    hostname: your-hostname
                    action: stop
                    client: cli
                    dryrun: "false"
                annotations:
                    description: Instance {{$labels.instance}} has GPU usage of {{ $value }}%
                    summary: Unused GPU

# Example AlertManager routes and receivers to trigger this API
    alertmanager:
        enabled: true
        config:
            ...
            route:
                ...
                group_by: ['namespace']
                routes:
                ...
                - receiver: 'ec2operator'
                  matchers:
                    - alertname =~ "UnusedGPU"
            receivers:
            - name: 'ec2operator'
              webhook_configs:
              - url: 'https://ec2op.your-domain.com/operateHostname'
                send_resolved: false

# Usage - Local, [Docker](https://hub.docker.com/repository/docker/rahul67/ec2-operator/general) and Kubernetes
Default properties are picked up from ENV:
* `HOST` - `0.0.0.0` (IP on which server listens, you can also change it to `localhost` for testing purposes)
* `PORT` - `8080`

Use Docker to run it locally or on any EC2 instance:

    docker run -p 8080:8080 rahul67/ec2-operator:v0.7

Refer [examples/](https://github.com/rahul67/ec2-operator/tree/master/examples) for kubernetes deployment along with service definition.
