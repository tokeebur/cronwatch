# cronwatch

Lightweight daemon that monitors cron job execution and sends alerts on failures.

## Installation

```bash
go install github.com/cronwatch/cronwatch@latest
```

Or build from source:

```bash
git clone https://github.com/cronwatch/cronwatch.git && cd cronwatch && make build
```

## Usage

Wrap your cron job with `cronwatch` to monitor its execution:

```bash
# In your crontab
* * * * * cronwatch --job "backup" --alert email /usr/local/bin/backup.sh
```

Configure alerts in `cronwatch.yaml`:

```yaml
alerts:
  email:
    to: ops@example.com
    on_failure: true
    on_timeout: true
timeout: 30m
```

Run as a daemon to monitor multiple jobs:

```bash
cronwatch daemon --config /etc/cronwatch/cronwatch.yaml
```

Check job status:

```bash
cronwatch status
cronwatch status --job backup
```

## Configuration

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to config file | `./cronwatch.yaml` |
| `--job` | Job name for tracking | required |
| `--alert` | Alert channel (email, slack, webhook) | none |
| `--timeout` | Max allowed runtime | `0` (disabled) |

## License

MIT © cronwatch contributors