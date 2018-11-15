package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	_ "net/http/pprof"

	_ "github.com/q3k/statusz"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/exaring/openconfig-streaming-telemetry-exporter/pkg/collector"
	"github.com/exaring/openconfig-streaming-telemetry-exporter/pkg/config"
	"github.com/exaring/openconfig-streaming-telemetry-exporter/pkg/frontend"
)

const version string = "0.0.0"

var (
	showVersion = flag.Bool("version", false, "Print version information.")
	configFile  = flag.String("config.file", "config.yml", "Path to config file")
)

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: openconfig-streaming-telemetry-exporter [ ... ]\n\nParameters:")
		fmt.Println()
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("could not load config file. %v", err)
	}

	col := collector.New(cfg)
	for _, target := range cfg.Targets {
		go func(target *config.Target) {
			t := col.AddTarget(target, cfg.StringValueMapping, true)
			conn, err := grpc.Dial(fmt.Sprintf("%s:%d", target.Hostname, target.Port),
				grpc.WithInsecure(),
				grpc.WithKeepaliveParams(keepalive.ClientParameters{
					Time:    time.Second * time.Duration(target.KeepaliveS),
					Timeout: time.Second * time.Duration(target.TimeoutS),
				}),
			)
			if err != nil {
				log.Errorf("Unable to dial: %v", err)
				return
			}

			t.Serve(conn)
		}(target)
	}

	fe := frontend.New(cfg, col)
	go fe.Start()

	select {}
}

func loadConfig() (*config.Config, error) {
	log.Infoln("Loading config from", *configFile)
	b, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return nil, err
	}

	return config.Load(bytes.NewReader(b))
}

func printVersion() {
	fmt.Println("openconfig_streaming_telemetry_exporter")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("Author(s): Annika Wickert, Oliver Herms")
	fmt.Println("Metric exporter for switches and routers via openconfig streaming telemetry")
}
