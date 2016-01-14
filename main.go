package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sync"

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
	_ "github.com/abrander/agento/plugins/agents/netstat"
	_ "github.com/abrander/agento/plugins/agents/openfiles"
	_ "github.com/abrander/agento/plugins/agents/snmpstats"
	_ "github.com/abrander/agento/plugins/agents/socketstats"
	_ "github.com/abrander/agento/plugins/transports/local"
	"github.com/abrander/agento/plugins/transports/ssh"
	"github.com/abrander/agento/server"
)

var config = configuration.Configuration{}

func main() {
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

	engine := gin.New()

	if config.Server.Http.Enabled || config.Server.Https.Enabled {
		server.Init(engine, config)
	}

	if config.Server.Http.Enabled {
		wg.Add(1)
		go server.ListenAndServe(engine)
	}

	if config.Server.Https.Enabled {
		wg.Add(1)
		go server.ListenAndServeTLS(engine)
	}

	if config.Server.Udp.Enabled {
		wg.Add(1)
		go server.ListenAndServeUDP()
	}

	if config.Client.Enabled {
		wg.Add(1)
		go client.GatherAndReport(config.Client)
	}

	if config.Monitor.Enabled {
		monitor.Init(config.Monitor)
		emitter := monitor.NewSimpleEmitter()
		scheduler := monitor.NewScheduler(emitter)
		wg.Add(1)
		go scheduler.Loop(*wg)

		go api.Init(engine.Group("/api"), scheduler, emitter)

		// Website for debugging
		templ := template.Must(template.New("web/index.html").Delims("[[", "]]").ParseFiles("web/index.html"))
		engine.SetHTMLTemplate(templ)
		engine.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"sshPublicKey": ssh.PublicKey,
			})
		})
		engine.Static("/static", "web/")
	}

	wg.Wait()
}
