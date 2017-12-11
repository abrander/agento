package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/abrander/agento/api"
	"github.com/abrander/agento/client"
	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/core"
	"github.com/abrander/agento/logger"
	"github.com/abrander/agento/monitor"
	"github.com/abrander/agento/plugins"
	_ "github.com/abrander/agento/plugins/agents/cpustats"
	_ "github.com/abrander/agento/plugins/agents/diskstats"
	_ "github.com/abrander/agento/plugins/agents/diskusage"
	_ "github.com/abrander/agento/plugins/agents/entropy"
	_ "github.com/abrander/agento/plugins/agents/hostname"
	_ "github.com/abrander/agento/plugins/agents/http"
	_ "github.com/abrander/agento/plugins/agents/linuxhost"
	_ "github.com/abrander/agento/plugins/agents/loadstats"
	_ "github.com/abrander/agento/plugins/agents/memorystats"
	_ "github.com/abrander/agento/plugins/agents/mysql"
	_ "github.com/abrander/agento/plugins/agents/netfilter"
	_ "github.com/abrander/agento/plugins/agents/netstat"
	_ "github.com/abrander/agento/plugins/agents/null"
	_ "github.com/abrander/agento/plugins/agents/openfiles"
	_ "github.com/abrander/agento/plugins/agents/phpfpm"
	_ "github.com/abrander/agento/plugins/agents/snmpstats"
	_ "github.com/abrander/agento/plugins/agents/socketstats"
	_ "github.com/abrander/agento/plugins/agents/tcpport"
	_ "github.com/abrander/agento/plugins/transports/local"
	_ "github.com/abrander/agento/plugins/transports/ssh"
	"github.com/abrander/agento/server"
	"github.com/abrander/agento/timeseries"
	"github.com/abrander/agento/userdb"
)

var configPath = "/etc/agento.conf"
var config = configuration.Configuration{}

func init() {
	// This should not be used for crypto, time.Now() is enough.
	rand.Seed(time.Now().UnixNano())
}

func main() {
	rootCommand := &cobra.Command{
		Use:   os.Args[0],
		Short: "Agento is a Client/server platform collecting near realtime metrics.",
	}

	gendocCommand := &cobra.Command{
		Use:    "gendoc",
		Short:  "Generate markdown documentation for all agents",
		Run:    gendoc,
		Args:   cobra.NoArgs,
		Hidden: true,
	}
	rootCommand.AddCommand(gendocCommand)

	runCommand := &cobra.Command{
		Use:   "run",
		Short: "Run an Agento node",
		Run:   run,
		Args:  cobra.NoArgs,
	}
	rootCommand.AddCommand(runCommand)

	rootCommand.PersistentFlags().StringVar(&configPath, "config", configPath, "The configuration file to use")
	rootCommand.Execute()
}

func gendoc(_ *cobra.Command, _ []string) {
	d := plugins.GetDoc()

	for _, doc := range d {
		fmt.Printf("| %-40s | %-110s |\n", doc.Info.Description, "")
		fmt.Printf("|------------------------------------------|----------------------------------------------------------------------------------------------------------------|\n")
		for name, description := range doc.Tags {
			fmt.Printf("| %-40s | %-110s |\n", "Tag: "+name, description)
		}
		for name, description := range doc.Measurements {
			fmt.Printf("| %-40s | %-110s |\n", name, description)
		}
		fmt.Printf("\n")
	}
}

func run(_ *cobra.Command, _ []string) {
	err := config.LoadFromFile(configPath)

	if err != nil {
		logger.Red("agento", "Configuration error: %s", err.Error())
		os.Exit(1)
	}

	wg := &sync.WaitGroup{}

	db := userdb.NewSingleUser(config.Server.Secret)
	engine := gin.New()

	var store core.Store

	emitter := core.NewSimpleEmitter()

	// If the user have Mongo enabled, we use that. If not, we read from
	// configuration.
	if config.Mongo.Enabled {
		store, err = monitor.NewMongoStore(config.Mongo, emitter)
		if err != nil {
			logger.Red("agento", "Mongo error: %s", err.Error())
			os.Exit(1)
		}
	} else {
		store, err = monitor.NewConfigurationStore(&config, emitter)
		if err != nil {
			logger.Red("agento", "Configuration error: %s", err.Error())
			os.Exit(1)
		}
	}

	scheduler := monitor.NewScheduler(store, db)

	serv, err := server.NewServer(engine, config.Server, db, store)
	if err != nil {
		logger.Red("agento", "Server error: %s", err.Error())
		os.Exit(1)
	}

	tsdb, err := timeseries.NewInfluxDb(&config.Server.Influxdb)
	if err != nil {
		logger.Red("agento", "InfluxDB error: %s", err.Error())
		os.Exit(1)
	}

	if config.Server.HTTP.Enabled {
		wg.Add(1)
		go serv.ListenAndServe(engine)
	}

	if config.Server.HTTPS.Enabled {
		wg.Add(1)
		go serv.ListenAndServeTLS(engine)
	}

	if config.Server.UDP.Enabled {
		wg.Add(1)
		go serv.ListenAndServeUDP()
	}

	if config.Client.Enabled {
		wg.Add(1)
		go client.GatherAndReport(config.Client)
	}

	wg.Add(1)
	go scheduler.Loop(*wg, tsdb)

	go api.Init(engine.Group("/api"), store, emitter, db)

	wg.Wait()
}
