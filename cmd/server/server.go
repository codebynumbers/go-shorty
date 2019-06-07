package main

import (
	"encoding/hex"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/julienschmidt/httprouter"
	"hash/fnv"
	"html/template"
	"log"
	"net/http"
	"strings"
)

// Config
const port = "3000"
const domain = "localhost"

var serving_domain = fmt.Sprintf("%s:%s", domain, port)

var client = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379", // use default Addr
	Password: "",               // no password set
	DB:       0,                // use default DB
})

func main() {
	router := httprouter.New()
	router.GET("/", indexHandler)
	router.GET("/:tag", expandHandler)
	router.POST("/data/shorten/", shortenHandler)
	log.Println(fmt.Sprintf("Listening on %s...", serving_domain))
	http.ListenAndServe(serving_domain, router)
}

func indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func expandHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	tag := ps.ByName("tag")

	url, err := client.Get(fmt.Sprintf("urls:%s", tag)).Result()
	if err == redis.Nil {
		w.WriteHeader(404)
		w.Write([]byte("404 page not found"))
	} else if err != nil {
		panic(err)
	} else {
		if !strings.HasPrefix(url, "http") {
			url = "http://" + url
		}
		http.Redirect(w, r, url, 301)
	}
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
	tag, err := client.Get(fmt.Sprintf("tags:%s", url)).Result()
	if err == redis.Nil {
		hash := fnv.New32a()
		hash.Write([]byte(url))
		tag = hex.EncodeToString(hash.Sum(nil))
		client.Set(fmt.Sprintf("urls:%s", tag), url, 0).Err()
		client.Set(fmt.Sprintf("tags:%s", url), tag, 0).Err()

	} else if err != nil {
		panic(err)
	}

	return fmt.Sprintf("http://%s/%s", serving_domain, tag)
}
