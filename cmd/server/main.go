// Package provides a simple url shortener using redis fronted sqlite
package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
	"hash/fnv"
	"html/template"
	"log"
	"net/http"
	"strings"
)

// Config
const port = "3000"
const domain = "localhost"
const dbDriver = "sqlite3"
const dbPath = "./shorty.db"

var servingDomain = fmt.Sprintf("%s:%s", domain, port)

var client = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379", // use default Addr
	Password: "",               // no password set
	DB:       0,                // use default DB
})

var db *sql.DB

func main() {
	initDb()
	router := httprouter.New()
	router.GET("/", indexHandler)
	router.GET("/:tag", expandHandler)
	router.POST("/data/shorten/", shortenHandler)
	log.Println(fmt.Sprintf("Listening on %s...", servingDomain))
	http.ListenAndServe(servingDomain, router)
}

func initDb() {
	var err error
	db, err = sql.Open(dbDriver, dbPath)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
	tmpl.Execute(w, nil)
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
		w.Write([]byte(shorten(url) + "\n"))
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
		}

		// update cache
		client.Set(fmt.Sprintf("urls:%s", tag), url, 0)
	}

	return url
}
