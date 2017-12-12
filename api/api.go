package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/abrander/agento/core"
	"github.com/abrander/agento/logger"
	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/userdb"
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
	startTime  = time.Now()
	wsupgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func wsHandler(c *gin.Context, emitter core.Emitter, subject userdb.Subject) {
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	ticker := time.Tick(time.Second)
	changes := emitter.Subscribe(subject)

	status := Status{
		Started: startTime,
	}

	for {
		select {
		case t := <-ticker:
			status.Clock = t
			status.Uptime = t.Sub(startTime)
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

func getSubject(c *gin.Context) userdb.Subject {
	return c.MustGet("subject").(userdb.Subject)
}

func getAccountId(c *gin.Context) string {
	// First try to read header
	accountId := c.Request.Header.Get("X-Agento-Account")

	if accountId != "" {
		return accountId
	}

	subject := c.MustGet("subject")
	switch subject.(type) {
	case userdb.Account:
		return subject.(userdb.Account).GetId()
	case userdb.User:
		accounts, err := subject.(userdb.User).GetAccounts()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return ""
		}

		if len(accounts) == 1 {
			return accounts[0].GetId()
		}
	}

	c.AbortWithError(http.StatusInternalServerError, userdb.ErrorInvalidAccountId)
	return ""
}

func Init(router gin.IRouter, store core.Store, emitter core.Emitter, db userdb.Database) {
	router.GET("/ws/:key", func(c *gin.Context) {
		key := c.Param("key")
		subject, error := db.ResolveKey(key)
		if error != nil {
			logger.Yellow("api", "[%s %s] Could not resolve API key '%s' from %s, aborting", c.Request.Method, c.Request.URL, key, c.Request.RemoteAddr)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		logger.Green("api", "[%s %s] API key '%s' authorized for %s", c.Request.Method, c.Request.URL, key, subject.GetId())

		wsHandler(c, emitter, subject)
	})

	router.Use(func(c *gin.Context) {
		key := c.Request.Header.Get("X-Agento-Secret")
		if key == "" {
			logger.Yellow("api", "[%s %s] No API key found, aborting request from %s", c.Request.Method, c.Request.URL, c.Request.RemoteAddr)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		subject, error := db.ResolveKey(key)
		if error != nil {
			logger.Yellow("api", "[%s %s] Could not resolve API key '%s' from %s, aborting", c.Request.Method, c.Request.URL, key, c.Request.RemoteAddr)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		logger.Green("api", "[%s %s] API key '%s' authorized for %s", c.Request.Method, c.Request.URL, key, subject.GetId())

		c.Set("subject", subject)
	})

	{
		a := router.Group("/agent")

		a.GET("/", func(c *gin.Context) {
			c.JSON(200, plugins.GetDocAgents())
		})

	}

	{
		h := router.Group("/host")

		h.DELETE("/:id", func(c *gin.Context) {
			id := c.Param("id")
			subject := getSubject(c)

			err := store.DeleteHost(subject, id)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, nil)
			}
		})

		h.POST("/new", func(c *gin.Context) {
			var host core.Host
			subject := getSubject(c)

			c.Bind(&host)
			err := store.AddHost(subject, &host)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, host)
			}
		})

		h.GET("/", func(c *gin.Context) {
			subject := getSubject(c)
			accountId := getAccountId(c)

			hosts, err := store.GetAllHosts(subject, accountId)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, hosts)
			}
		})
	}

	{
		m := router.Group("/probe")

		m.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			subject := getSubject(c)

			probe, err := store.GetProbe(subject, id)
			if err != nil {
				c.AbortWithError(404, err)
			} else {
				c.JSON(200, probe)
			}
		})

		m.PUT("/:id", func(c *gin.Context) {
			var probe core.Probe
			subject := getSubject(c)

			c.Bind(&probe)
			err := store.UpdateProbe(subject, &probe)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, probe)
			}
		})

		m.DELETE("/:id", func(c *gin.Context) {
			id := c.Param("id")
			subject := getSubject(c)

			err := store.DeleteProbe(subject, id)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, nil)
			}
		})

		m.POST("/new", func(c *gin.Context) {
			var probe core.Probe
			subject := getSubject(c)

			c.Bind(&probe)
			err := store.AddProbe(subject, &probe)
			if err != nil {
				logger.Yellow("api", "Error: %s", err.Error())
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, probe)
			}
		})

		m.GET("/", func(c *gin.Context) {
			subject := getSubject(c)
			accountId := getAccountId(c)

			probes, err := store.GetAllProbes(subject, accountId)
			if err != nil {
				c.AbortWithError(500, err)
			} else {
				c.JSON(200, probes)
			}
		})
	}

	{
		t := router.Group("/transport")

		t.GET("/", func(c *gin.Context) {
			c.JSON(200, plugins.GetDocTransports())
		})
	}
}
