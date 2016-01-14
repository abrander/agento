package api

import (
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/abrander/agento/monitor"
	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/plugins/transports/ssh"
)

type (
	Message struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}

	Status struct {
		Uptime  time.Duration `json:"uptime"`
		Clock   time.Time     `json:"clock"`
		Started time.Time     `json:"start"`
	}
)

var (
	StartTime  = time.Now()
	wsupgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func wsHandler(c *gin.Context, emitter monitor.Emitter) {
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	ticker := time.Tick(time.Second)
	changes := emitter.Subscribe()

	status := Status{
		Started: StartTime,
	}

	for {
		select {
		case t := <-ticker:
			status.Clock = t
			status.Uptime = t.Sub(StartTime)
			err := conn.WriteJSON(Message{Type: "status", Payload: status})
			if err != nil {
				goto unsubscribe
			}
		case msg := <-changes:
			err := conn.WriteJSON(msg)
			if err != nil {
				goto unsubscribe
			}
		}
	}

unsubscribe:
	emitter.Unsubscribe(changes)
}

func Run(wg sync.WaitGroup, admin monitor.Admin, emitter monitor.Emitter) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Logger())

	router.Use(static.Serve("/", static.LocalFile("web/", false)))

	router.GET("/ws", func(c *gin.Context) {
		wsHandler(c, emitter)
	})

	{
		a := router.Group("/agent")

		a.GET("/", func(c *gin.Context) {
			c.JSON(200, plugins.AvailableAgents())
		})

	}

	{
		h := router.Group("/host")

		h.DELETE("/:id", func(c *gin.Context) {
			id := c.Param("id")

			err := admin.DeleteHost(id)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, nil)
			}
		})

		h.POST("/new", func(c *gin.Context) {
			var host monitor.Host
			c.Bind(&host)
			err := admin.AddHost(&host)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, host)
			}
		})

		h.GET("/", func(c *gin.Context) {
			c.JSON(200, admin.GetAllHosts())
		})
	}

	{
		m := router.Group("/monitor")

		m.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")

			mon, err := admin.GetMonitor(id)
			if err == monitor.ErrorInvalidId {
				c.AbortWithError(400, err)
			} else if err != nil {
				c.AbortWithError(404, err)
			} else {
				c.JSON(200, mon)
			}
		})

		m.PUT("/:id", func(c *gin.Context) {
			var mon monitor.Monitor
			c.Bind(&mon)
			err := admin.UpdateMonitor(&mon)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, mon)
			}
		})

		m.DELETE("/:id", func(c *gin.Context) {
			id := c.Param("id")

			err := admin.DeleteMonitor(id)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, nil)
			}
		})

		m.POST("/new", func(c *gin.Context) {
			var mon monitor.Monitor
			c.Bind(&mon)
			err := admin.AddMonitor(&mon)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, mon)
			}
		})

		m.GET("/", func(c *gin.Context) {
			c.JSON(200, admin.GetAllMonitors())
		})
	}

	{
		t := router.Group("/transport")

		t.GET("/", func(c *gin.Context) {
			c.JSON(200, plugins.AvailableTransports())
		})
	}

	templ := template.Must(template.New("web/index.html").Delims("[[", "]]").ParseFiles("web/index.html"))
	router.SetHTMLTemplate(templ)

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"sshPublicKey": ssh.PublicKey,
		})
	})

	router.Run(":9901")

	wg.Done()
}
