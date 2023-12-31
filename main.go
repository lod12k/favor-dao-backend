package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"favor-dao-backend/internal"
	"favor-dao-backend/internal/conf"
	"favor-dao-backend/internal/routers"
	"favor-dao-backend/internal/service"
	"favor-dao-backend/pkg/debug"
	"favor-dao-backend/pkg/util"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

var (
	deleteAddress     string
	pushSearch        bool
	noDefaultFeatures bool
	features          suites
)

type suites []string

func (s *suites) String() string {
	return strings.Join(*s, ",")
}

func (s *suites) Set(value string) error {
	for _, item := range strings.Split(value, ",") {
		*s = append(*s, strings.TrimSpace(item))
	}
	return nil
}

func init() {
	flagParse()

	conf.Initialize(features, noDefaultFeatures)
	internal.Initialize()
}

func flagParse() {
	flag.StringVar(&deleteAddress, "del-address", "", "Cancellation user and real delete")
	flag.BoolVar(&pushSearch, "push-search", false, "push posts to search")
	flag.BoolVar(&noDefaultFeatures, "no-default-features", false, "whether use default features")
	flag.Var(&features, "features", "use special features")
	flag.Parse()
}

func main() {
	if pushSearch {
		service.PushPostsToSearch()
		service.PushDAOsToSearch()
		return
	}
	if deleteAddress != "" {
		err := service.Cancellation(deleteAddress)
		if err != nil {
			panic(err)
		}
		log.Println("successful")
		return
	}
	go service.CancellationTask()

	gin.SetMode(conf.ServerSetting.RunMode)

	router := routers.NewRouter()
	s := &http.Server{
		Addr:           conf.ServerSetting.HttpIp + ":" + conf.ServerSetting.HttpPort,
		Handler:        router,
		ReadTimeout:    conf.ServerSetting.ReadTimeout,
		WriteTimeout:   conf.ServerSetting.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	util.PrintHelloBanner(debug.VersionInfo())
	fmt.Fprintf(color.Output, "favor dao service listen on %s\n",
		color.GreenString(fmt.Sprintf("http://%s:%s", conf.ServerSetting.HttpIp, conf.ServerSetting.HttpPort)),
	)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Fatalf("app stoped: %s", err)
		}
		sigs <- syscall.SIGTERM
	}()
	select {
	case <-sigs:
	}
	s.Shutdown(context.TODO())
}
