package main

import (
	"database/sql"
	"fmt"
	"github.com/mongmx/booster/application/infrastructure/postgres"
	"github.com/mongmx/booster/application/member"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"time"

	"github.com/chenjiandongx/ginprom"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	//env     = "development"
	sslMode = "disable"
	dbPort  = "5432"
	dbHost  = "localhost"
	dbUser  = "root"
	dbPass  = "root"
	dbName  = "booster"

	g errgroup.Group
)

func main() {
	conn := fmt.Sprintf(
		"dbname=%s user=%s password=%s host=%s port=%s sslmode=%s",
		dbName, dbUser, dbPass, dbHost, dbPort, sslMode,
	)
	db, err := sql.Open("postgres", conn)
	must(err)
	defer func() {
		err := db.Close()
		must(err)
	}()
	serverAPI := &http.Server{
		Addr:         ":8080",
		Handler:      routerAPI(db),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	serverMetrics := &http.Server{
		Addr:         ":8081",
		Handler:      routerMetrics(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	g.Go(func() error {
		log.Println("API server listen on :8080")
		return serverAPI.ListenAndServe()
	})
	g.Go(func() error {
		log.Println("Metrics server listen on :8081")
		return serverMetrics.ListenAndServe()
	})
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}

func routerAPI(db *sql.DB) http.Handler {
	e := gin.New()
	e.Use(gin.Logger())
	e.Use(gin.Recovery())
	e.Use(ginprom.PromMiddleware(nil))

	memberRepo, err := postgres.NewMemberRepository(db)
	must(err)
	memberService, err := member.NewService(memberRepo)
	must(err)
	member.Routes(e, memberService)

	return e
}

func routerMetrics() http.Handler {
	e := gin.New()
	e.Use(gin.Logger())
	e.Use(gin.Recovery())
	e.GET("/metrics", ginprom.PromHandler(promhttp.Handler()))
	e.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	e.GET("/pong", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ping",
		})
	})
	return e
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
