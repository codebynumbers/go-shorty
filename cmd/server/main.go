// Package provides a simple url shortener using redis fronted sqlite
package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/codebynumbers/go-shorty/internal/configuration"
	"github.com/go-redis/redis"
	"github.com/julienschmidt/httprouter"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/mattn/go-sqlite3"
	"hash/fnv"
	"html/template"
	"log"
	"net/http"
	"strings"
)

var servingDomain string
var client *redis.Client
var db *sql.DB
var config configuration.Config

type ResultPageData struct {
	ShortenedUrl string
}

func main() {
	var err error

	if err = envconfig.Process("GOSHORTY", &config); err != nil {
		log.Fatal(err)
	}

	initDbs()
	router := httprouter.New()
	router.GET("/", indexHandler)
	router.GET("/:tag", expandHandler)
	router.POST("/data/shorten/", shortenHandler)

	servingDomain = fmt.Sprintf("%s:%s", config.Domain, config.HostPort)
	log.Println(fmt.Sprintf("Listening on %s...", servingDomain))
	http.ListenAndServe(servingDomain, router)
}

func initDbs() {
	var err error
	db, err = sql.Open(config.DbDriver, config.DbPath)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	client = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tmpl, _ := template.New("").ParseFiles("web/templates/index.html", "web/templates/base.html")
	_ = tmpl.ExecuteTemplate(w, "base", nil)
}

func expandHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	tag := ps.ByName("tag")

	url := cachedGetUrl(tag)

	// give up
	if url == "" {
		w.WriteHeader(404)
		w.Write([]byte("404 page not found"))
		return
	}

	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	// respond
	http.Redirect(w, r, url, 301)
}

func shortenHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()
	url := r.Form.Get("url")
	if url != "" {

		context := ResultPageData{
			ShortenedUrl: shorten(url),
		}

		tmpl, _ := template.New("").ParseFiles("web/templates/result.html", "web/templates/base.html")
		_ = tmpl.ExecuteTemplate(w, "base", context)
	}
}

// shorten encodes the url and returns and new url to reach it at
func shorten(url string) string {
	hash := fnv.New32a()
	hash.Write([]byte(url))
	tag := hex.EncodeToString(hash.Sum(nil))

	cachedUrl := cachedGetUrl(tag)

	if cachedUrl == "" {
		stmt, err := db.Prepare("INSERT INTO urls (tag, url) values (?, ?)")

		if err != nil {
			log.Println(err)
		}
		_, err = stmt.Exec(tag, url)

		if err != nil {
			log.Println(err)
		}
	}

	return fmt.Sprintf("http://%s/%s", servingDomain, tag)
}

// cachedGetUrl will check redis for url by tag, then db. If found in db, update the cache.
func cachedGetUrl(tag string) string {

	url, _ := client.Get(fmt.Sprintf("urls:%s", tag)).Result()

	// Check db
	if url == "" {
		rows, err := db.Query("SELECT url from urls where tag=?", tag)

		if err != nil {
			log.Println(err)
		}

		if rows.Next() {
			rows.Scan(&url)

			// update cache
			client.Set(fmt.Sprintf("urls:%s", tag), url, 0)
		}
		rows.Close()
	}

	return url
}
