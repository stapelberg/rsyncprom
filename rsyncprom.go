// Package rsyncprom implements a parser that extracts transfer details from
// rsync standard output output.
//
// This package contains the parser, see cmd/rsync-prom for a wrapper program.
//
// # Rsync Requirements
//
// Start rsync with --verbose (-v) or --stats to enable printing transfer
// totals.
//
// Do not use the --human-readable (-h) flag in your rsync invocation, otherwise
// rsyncprom cannot parse the output!
//
// Run rsync in the C.UTF-8 locale to prevent rsync from localizing decimal
// separators and fractional points in big numbers.
package rsyncprom

import (
	"context"
	"io"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/stapelberg/rsyncparse"
)

// Stats contains all data found in rsync output.
type Stats = rsyncparse.Stats

// Parse reads from the specified io.Reader and scans individual lines. rsync
// transfer totals are extracted when found, and returned in the Stats struct.
func Parse(r io.Reader) (*Stats, error) {
	return rsyncparse.Parse(r)
}

// WrapParams is the configuration struct for the WrapRsync() function.
type WrapParams struct {
	// Address of a Prometheus push gateway. This is passed as url parameter to
	// https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/push#New
	Pushgateway string
	// Prometheus instance label.
	Instance string
	// Prometheus job name.
	Job string
}

// WrapRsync starts one rsync invocation and pushes prometheus metrics about the
// invocation to the Prometheus push gateway specified in the WrapParams.
//
// This function is used by the cmd/rsync-prom wrapper tool, but you can also
// use it programmatically and start rsync remotely via SSH instead of wrapping
// the process, for example.
func WrapRsync(ctx context.Context, params *WrapParams, args []string, start func(context.Context, []string) (io.Reader, error), wait func() int) error {
	log.Printf("push gateway: %q", params.Pushgateway)

	startTimeMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: params.Job + "_start_timestamp_seconds",
		Help: "The timestamp of the rsync start",
	})
	startTimeMetric.SetToCurrentTime()
	pushAll := func(collectors ...prometheus.Collector) {
		p := push.New(params.Pushgateway, params.Job).
			Grouping("instance", params.Instance)
		for _, c := range collectors {
			p.Collector(c)
		}
		if err := p.Add(); err != nil {
			log.Print(err)
		}
	}
	pushAll(startTimeMetric)

	exitCode := 0
	defer func() {
		log.Printf("Pushing exit code %d", exitCode)
		exitCodeMetric := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: params.Job + "_exit_code",
			Help: "The exit code (0 = success, non-zero = failure)",
		})
		exitCodeMetric.Set(float64(exitCode))
		// end timestamp is push_time_seconds
		pushAll(exitCodeMetric)
	}()

	rd, err := start(ctx, args)
	if err != nil {
		exitCode = 254
		return err
	}

	log.Printf("Parsing rsync output")
	parsed, err := Parse(rd)
	if err != nil {
		return err
	}

	if parsed.Found {
		totalWrittenMetric := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "rsync_total_written",
			Help: "Total bytes written for this transfer",
		})
		totalWrittenMetric.Set(float64(parsed.TotalWritten))

		totalReadMetric := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "rsync_total_read",
			Help: "Total bytes read for this transfer",
		})
		totalReadMetric.Set(float64(parsed.TotalRead))

		bytesPerSec := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "rsync_bytes_per_sec",
			Help: "bytes per second",
		})
		bytesPerSec.Set(float64(parsed.BytesPerSec))

		totalSize := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "rsync_total_size",
			Help: "Total size of all processed files, in bytes",
		})
		totalSize.Set(float64(parsed.TotalSize))

		pushAll(totalWrittenMetric, totalReadMetric, totalSize)
	}

	log.Printf("Waiting for rsync to exit")
	exitCode = wait()

	return nil
}
