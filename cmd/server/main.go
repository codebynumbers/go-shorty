// Package provides a simple url shortener using redis fronted sqlite
package main

import (
	"fmt"
	"github.com/codebynumbers/go-shorty/internal/configuration"
	"github.com/codebynumbers/go-shorty/internal/connections"
	"github.com/codebynumbers/go-shorty/internal/handlers"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func main() {

	config := configuration.Configure()
	connections.InitDb(config)
	connections.InitRedis(config)

	handlerEnv := handlers.HandlerEnv{
		AppConfig: config,
		Db:        connections.Db,
		Cache:     connections.Cache,
	}

	router := httprouter.New()
	router.GET("/", handlerEnv.IndexHandler)
	router.GET("/:tag", handlerEnv.ExpandHandler)
	router.POST("/data/shorten/", handlerEnv.ShortenHandler)

	servingDomain := fmt.Sprintf("%s:%s", config.HostDomain, config.HostPort)
	log.Println(fmt.Sprintf("Listening on %s...", servingDomain))
	http.ListenAndServe(servingDomain, router)
}
