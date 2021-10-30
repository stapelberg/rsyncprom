# rsync-prom

[![Go Reference](https://pkg.go.dev/badge/github.com/stapelberg/rsyncprom.svg)](https://pkg.go.dev/github.com/stapelberg/rsyncprom)

An rsync wrapper (or output parser) that pushes metrics to
[prometheus](https://prometheus.io/).

This allows you to then build dashboards and alerting for your rsync batch jobs.

## Installation

```
go install github.com/stapelberg/rsyncprom/cmd/rsync-prom@latest
```

## Setup example: crontab

```
9 9 * * * /home/michael/go/bin/rsync-prom --instance="sync-drive" -- /home/michael/sync-drive.sh
```

## Setup example: systemd

```
[Service]
ExecStart=/home/michael/go/bin/rsync-prom --instance="sync-wiki" -- /usr/bin/rsync --exclude data/cache -av --checksum server:wiki/ /var/cache/wiki
```
