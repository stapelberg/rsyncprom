// Program rsync-prom is a wrapper for rsync which exports prometheus metrics to
// a prometheus push gateway.
//
// See the rsyncprom package documentation for how to start rsync.
package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/stapelberg/rsyncprom"
)

func rsyncprommain() error {
	hostname, err := os.Hostname()
	if err != nil {
		log.Print(err)
	}
	var params rsyncprom.WrapParams
	flag.StringVar(&params.Pushgateway, "prometheus_push_gateway",
		"https://pushgateway.ts.zekjur.net",
		"URL for the https://github.com/prometheus/pushgateway service to push Prometheus metrics to")
	flag.StringVar(&params.Instance, "instance",
		"rsync@"+hostname,
		"prometheus instance label. should be as descriptive as possible")
	flag.StringVar(&params.Job, "job",
		"rsync",
		"prometheus job label")
	flag.Parse()

	ctx := context.Background()
	var rsync *exec.Cmd
	start := func(ctx context.Context, args []string) (io.Reader, error) {
		rsync = exec.CommandContext(ctx, args[0], args[1:]...)
		rc, err := rsync.StdoutPipe()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		log.Printf("Starting rsync %q", rsync.Args)
		if err := rsync.Start(); err != nil {
			return nil, err
		}
		return rc, nil
	}
	wait := func() int {
		if err := rsync.Wait(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					return status.ExitStatus()
				}
			}
			log.Print(err)
			return 1
		}
		return 0
	}
	return rsyncprom.WrapRsync(ctx, &params, flag.Args(), start, wait)
}

func main() {
	if err := rsyncprommain(); err != nil {
		log.Fatal(err)
	}
}
