package main

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/abrander/agento/api"
	"github.com/abrander/agento/client"
	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/logger"
	"github.com/abrander/agento/monitor"
	"github.com/abrander/agento/plugins"
	_ "github.com/abrander/agento/plugins/agents/cpuspeed"
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
	_ "github.com/abrander/agento/plugins/agents/snmpstats"
	_ "github.com/abrander/agento/plugins/agents/socketstats"
	_ "github.com/abrander/agento/plugins/agents/tcpport"
	_ "github.com/abrander/agento/plugins/transports/local"
	"github.com/abrander/agento/plugins/transports/ssh"
	"github.com/abrander/agento/server"
	"github.com/abrander/agento/timeseries"
	"github.com/abrander/agento/userdb"
)

var config = configuration.Configuration{}

func main() {
	// This should not be used for crypto, time.Now() is enough.
	rand.Seed(time.Now().UnixNano())

	if len(os.Args) > 1 && os.Args[1] == "gendoc" {
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
		os.Exit(0)
	}

	err := config.LoadFromFile("/etc/agento.conf")

	if err != nil {
		logger.Red("agento", "Configuration error: %s", err.Error())
		os.Exit(1)
	}

	wg := &sync.WaitGroup{}

	db := userdb.NewSingleUser(config.Server.Secret)
	engine := gin.New()

	var store monitor.Store

	emitter := monitor.NewSimpleEmitter()

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

	var tsdb timeseries.Database
	if config.Server.Http.Enabled || config.Server.Https.Enabled || config.Server.Udp.Enabled {
		tsdb, err = timeseries.NewInfluxDb(&config.Server.Influxdb)
		if err != nil {
			logger.Red("agento", "InfluxDB error: %s", err.Error())
			os.Exit(1)
		}
	}

	if config.Server.Http.Enabled {
		wg.Add(1)
		go serv.ListenAndServe(engine)
	}

	if config.Server.Https.Enabled {
		wg.Add(1)
		go serv.ListenAndServeTLS(engine)
	}

	if config.Server.Udp.Enabled {
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

	// Website for debugging
	templ := template.Must(template.New("web/index.html").Delims("[[", "]]").ParseFiles("web/index.html"))
	engine.SetHTMLTemplate(templ)
	engine.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"sshPublicKey": ssh.PublicKey(),
			"agentoSecret": config.Server.Secret,
		})
	})
	engine.Static("/static", "web/")

	wg.Wait()
}
