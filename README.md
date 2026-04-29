# cronwatch

A daemon that monitors cron job execution times and sends alerts when jobs drift outside expected windows.

## Installation

```bash
go install github.com/yourname/cronwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/cronwatch.git && cd cronwatch && go build ./...
```

## Usage

Define your expected job windows in a config file:

```yaml
# cronwatch.yaml
jobs:
  - name: daily-backup
    schedule: "0 2 * * *"
    window: 5m
    alert: slack
  - name: hourly-sync
    schedule: "0 * * * *"
    window: 2m
    alert: email
```

Then start the daemon:

```bash
cronwatch --config cronwatch.yaml
```

Wrap your existing cron jobs to report execution times:

```bash
# In your crontab
0 2 * * * cronwatch exec --job daily-backup -- /usr/local/bin/backup.sh
```

cronwatch will send an alert if a job starts late, runs too long, or fails to execute within the expected window.

### Alert Channels

| Channel | Config Key |
|---------|------------|
| Slack   | `slack`    |
| Email   | `email`    |
| PagerDuty | `pagerduty` |

## Configuration

See [docs/configuration.md](docs/configuration.md) for a full reference of available options.

## License

MIT © [yourname](https://github.com/yourname)