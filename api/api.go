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

func wsHandler(c *gin.Context) {
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	ticker := time.Tick(time.Second)
	changes := monitor.SubscribeChanges()

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
	monitor.UnsubscribeChanges(changes)
}

func Run(wg sync.WaitGroup) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Logger())

	router.Use(static.Serve("/", static.LocalFile("web/", false)))

	router.GET("/ws", wsHandler)

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

			err := monitor.DeleteHost(id)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, nil)
			}
		})

		h.POST("/new", func(c *gin.Context) {
			var host monitor.Host
			c.Bind(&host)
			err := monitor.AddHost(&host)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, host)
			}
		})

		h.GET("/", func(c *gin.Context) {
			c.JSON(200, monitor.GetAllHosts())
		})
	}

	{
		m := router.Group("/monitor")

		m.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")

			mon, err := monitor.GetMonitor(id)
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
			err := monitor.UpdateMonitor(&mon)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, mon)
			}
		})

		m.DELETE("/:id", func(c *gin.Context) {
			id := c.Param("id")

			err := monitor.DeleteMonitor(id)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, nil)
			}
		})

		m.POST("/new", func(c *gin.Context) {
			var mon monitor.Monitor
			c.Bind(&mon)
			err := monitor.AddMonitor(&mon)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, mon)
			}
		})

		m.GET("/", func(c *gin.Context) {
			c.JSON(200, monitor.GetAllMonitors())
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
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	router.Run(":9901")

	wg.Done()
}
