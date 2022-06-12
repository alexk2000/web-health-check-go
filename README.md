# Web endpoints (urls) monitoring tool (web health checks)

Simple web monitoring tool to check web endpoints, for example to ensure that you web site is alive. Alerts are sent to Slack.
This is not a replacement of professional monitoring systems like [Prometheus](https://prometheus.io/), this is just fast solution to start monitoring and get alerts in your team channel in Slack.

**Features:**

- yaml format of config file
- check methods: http status code, response, content type
- separate scheduler per http endpoint
- config auto reload (restart not needed if config changed)
- web server to provide endpoint (/healthz) for Kubernetes liveness probe
- currently only Slack notification method implemented but it's easy to add more methods

[Config file example](/conf/config.yml):
```
port: 8080
tz: Europe/Kiev
timeout: 5
# default scheduler
cron: "*/30 * 9-21 * * *"

notificationMethods:
  slack:
    type: slack
    webhook: https://hooks.slack.com/services/XXX/XXX/XXX
  email:
    type: email
    email: alert@example.com

notifications: 
  - slack

failureThreshold: 6
failureInterval: 12

dataDir: ./data
# in seconds
notificationInterval: 900

checks:
  - name: Prometheus Dev
    url: http://prometheus.dev.company.com/graph
    cron: "*/30 */1 9-21 * * 1-5"
  - name: company.com
    url: https://company.com/home/
    cron: "*/30 */1 9-21 * * 1-5"
  - name: company2.com
    url: https://company2.com/home/
  - name: Gitlab
    url: https://gitlab.company.com/-/liveness?token=XXX
    contentType: application/json
    response: '{"status":"ok"}'
  - name: Prometheus Prod
    url: http://prometheus.prod.company.com/graph
  - name: app1
    url: https://app1.company2.com/health_check
    response: '{"mongodb":"UP","amqp":"UP","status":"UP"}'
  - name: app2
    url: https://app2.prod.company.com/health_check
    response: '{"status":"UP"}'
```

**How to run**:

- create [Slack Incoming Webhooks](https://api.slack.com/messaging/webhooks)
- edit [config file](/conf/config.yml): set Slack Incoming Webhook url, configure needed checks, etc
- build and run:
```
❯ go build cmd/web-health-check.go
❯ ./web-health-check
```
