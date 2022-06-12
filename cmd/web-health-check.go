package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"time"

	"github.com/go-co-op/gocron"

	"web-health-check/pkg/checker"
	"web-health-check/pkg/config"
	"web-health-check/pkg/web"
)

var scheduler *gocron.Scheduler

func runCronJobs(conf config.Config) {
	tz, err := time.LoadLocation(conf.TZ)
	if err != nil {
		log.Fatal(err)
	}
	scheduler = gocron.NewScheduler(tz)

	log.Println("in runCronJobs", conf.Checks)
	for _, check := range conf.Checks {
		// use default cron schedule if not set for explicitly for check
		if check.Cron != "" {
			scheduler.CronWithSeconds(check.Cron).SingletonMode().Do(checker.DoCheck, check, conf)
		} else {
			scheduler.CronWithSeconds(conf.Cron).SingletonMode().Do(checker.DoCheck, check, conf)
		}
	}

	log.Printf("Number of jobs %v", len(scheduler.Jobs()))
	scheduler.StartAsync()
}

func reloader() {
	for conf := range config.Conf.SubscribeOnChange() {
		log.Println("Reloading jobs")
		scheduler.Stop()
		scheduler.Clear()
		runCronJobs(conf)
	}
}

func main() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	runCronJobs(config.Conf.Get())
	go reloader()

	web.StartAsync()

	<-sigs
}
