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
	exit24IsExit0 := flag.Bool("exit_24_is_exit_0",
		false,
		// e.g.: directory has vanished: "/home/michael/.local/share/containers/storage/overlay/2dcb7ef2c46242b1584effa327b6faf276075d76637d508775befdef4415e2c0"
		"rsync exits with status code 24 when a file or directory vanishes between listing and transferring it. this can be expected (when doing a full backup while working with docker containers, for example) or cause for concern (when replicating an ever-growing data set). when this flag is enabled, rsync-prom treats exit code 24 like exit code 0 (expected)")
	flag.Parse()

	ctx := context.Background()
	var rsync *exec.Cmd
	var stdoutPipe io.ReadCloser
	start := func(ctx context.Context, args []string) (io.Reader, error) {
		rsync = exec.CommandContext(ctx, args[0], args[1:]...)
		rsync.Stderr = os.Stderr
		rc, err := rsync.StdoutPipe()
		if err != nil {
			return nil, err
		}
		stdoutPipe = rc

		log.Printf("Starting rsync %q", rsync.Args)
		if err := rsync.Start(); err != nil {
			return nil, err
		}
		return rc, nil
	}
	wait := func() int {
		defer stdoutPipe.Close()
		if err := rsync.Wait(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					code := status.ExitStatus()
					if *exit24IsExit0 && code == 24 {
						code = 0
					}
					return code
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
