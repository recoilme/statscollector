package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"

	sp "github.com/recoilme/slowpoke"
	//"github.com/thinkerou/favicon"
)

type Hit struct {
	Referer string   `form:"referer" json:"referer" binding:"exists,alphanum,min=1,max=250"`
	Urls    []string `form:"urls" json:"urls" binding:"exists"`
}

type StatResp struct {
	Url   string
	View  uint64
	Click uint64
	CTR   float64
}

func main() {
	srv := &http.Server{
		Addr:    ":8088",
		Handler: InitRouter(),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// Close db
	if err := sp.CloseAll(); err != nil {
		log.Fatal("Database Shutdown:", err)
	}
	log.Println("Server exiting")

}

func InitRouter() *gin.Engine {
	r := gin.Default()

	r.Use(CORSMiddleware())
	r.POST("/api/view", View)
	r.POST("/api/click", Click)
	r.GET("/api/stat/:referer", Stat)

	return r
}

// View register urls view
// Example:
// curl -d '{"referer":"hotpop","urls":["url1","url2"]}' -H "Content-Type: application/json" -X POST http://localhost:8088/api/view
func View(c *gin.Context) {
	var err error
	switch c.Request.Method {
	case "POST":
		var h Hit
		err = c.ShouldBind(&h)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, err)
			return
		}
		if h.Urls != nil {
			for _, u := range h.Urls {
				sp.Counter("counters/view"+h.Referer, []byte(u))
				sp.CloseAll()
			}
			c.JSON(http.StatusOK, h)
			return
		}
		c.JSON(http.StatusUnprocessableEntity, errors.New("empty urls"))
	}
}

// Click register urls clicks
// Example:
// curl -d '{"referer":"hotpop","urls":["url1","url2"]}' -H "Content-Type: application/json" -X POST http://localhost:8088/api/click
func Click(c *gin.Context) {
	var err error
	switch c.Request.Method {
	case "POST":
		var h Hit
		err = c.ShouldBind(&h)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, err)
			return
		}
		if h.Urls != nil {
			for _, u := range h.Urls {
				sp.Counter("counters/click"+h.Referer, []byte(u))
				sp.CloseAll()
			}
			c.JSON(http.StatusOK, h)
			return
		}
		c.JSON(http.StatusUnprocessableEntity, errors.New("empty urls"))
	}
}

// Stat show stat
// Example:
// curl  -H "Content-Type: application/json" -X GET http://localhost:8088/api/stat/hotpop
// [{"Url":"url1","View":3,"Click":2,"CTR":1.5},{"Url":"url2","View":3,"Click":2,"CTR":1.5}]
func Stat(c *gin.Context) {
	//var err error
	referer := c.Param("referer")
	//fmt.Println("referer", referer)
	var resp []StatResp
	resp = make([]StatResp, 0, 0)
	switch c.Request.Method {
	case "GET":
		keys, err := sp.Keys("counters/view"+referer, nil, uint32(0), uint32(0), true)
		fmt.Println(keys)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, err)
			return
		}
		for _, key := range keys {
			var v, c uint64
			vbin, err := sp.Get("counters/view"+referer, key)
			if err == nil {
				v = binary.BigEndian.Uint64(vbin)
				cbin, err := sp.Get("counters/click"+referer, key)
				if err == nil {
					c = binary.BigEndian.Uint64(cbin)
				}
				var s StatResp
				s.Url = string(key)
				s.View = v
				s.Click = c
				if c > 0 {
					s.CTR = float64(v) / float64(c)
				} else {
					s.CTR = 0
				}
				resp = append(resp, s)
			}

		}
		c.JSON(http.StatusOK, resp)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			//fmt.Println("OPTIONS")
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}
