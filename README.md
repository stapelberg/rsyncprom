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

## Code example: SSH wrapper

Here’s an example for code which uses the `x/crypto/ssh` package to trigger
rsync on a remote machine and parses the output:

https://github.com/stapelberg/zkj-nas-tools/blob/02d46d718df60c413844d9218f6dd702ad94e5f1/dornroeschen/sshutil.go#L134-L139
