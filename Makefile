all: test
	CGO_ENABLED=0 go install -mod=mod

test:
	go test -mod=mod -v ./...

push:
	# midna uses /home/michael/go/bin/rsync-prom
	scp =rsync-prom funnel:/usr/local/bin/
