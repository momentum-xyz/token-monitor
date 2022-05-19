package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/OdysseyMomentumExperience/token-service/pkg/log"
	"github.com/OdysseyMomentumExperience/token-service/pkg/tokensvc"
	"github.com/prometheus/common/expfmt"
)

const configFileName = "config.dev.yaml"

func main() {
	err := start()
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		os.Exit(1)
	}
}

func start() error {
	configPath, ok := os.LookupEnv("CONFIG_PATH")	
	if !ok {
		configPath = configFileName
	}
	cfg := tokensvc.LoadConfig(configPath)
	cfg.PrettyPrint()

	svc, cleanup, err := tokensvc.NewMQTTTokenService(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = svc.Start(ctx)
	if err != nil {
		return err
	}

	log.Logln(0, "Initialized the Token Service...")
	wait()
	log.Logln(0, "Terminating...")

	if err := tokensvc.DumpMetrics("metrics/metrics.txt", expfmt.FmtText); err != nil {
		return err
	}

	return nil
}
func wait() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
}
