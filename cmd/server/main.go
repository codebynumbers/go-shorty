// Package provides a simple url shortener using redis fronted sqlite
package main

import (
	"fmt"
	"github.com/codebynumbers/go-shorty/internal/configuration"
	"github.com/codebynumbers/go-shorty/internal/connections"
	"github.com/codebynumbers/go-shorty/internal/handlers"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
)

func main() {

	config := configuration.Configure()
	connections.InitDb(config)
	connections.InitRedis(config)

	router := httprouter.New()
	router.GET("/", handlers.IndexHandler)
	router.GET("/:tag", handlers.ExpandHandler)
	router.POST("/data/shorten/", handlers.ShortenHandler)

	servingDomain := fmt.Sprintf("%s:%s", config.HostDomain, config.HostPort)
	log.Println(fmt.Sprintf("Listening on %s...", servingDomain))
	http.ListenAndServe(servingDomain, router)
}
