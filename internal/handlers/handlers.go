package handlers

import (
	"encoding/hex"
	"fmt"
	"github.com/codebynumbers/go-shorty/internal/configuration"
	"github.com/codebynumbers/go-shorty/internal/connections"
	"github.com/julienschmidt/httprouter"
	"hash/fnv"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type ResultPageData struct {
	ShortenedUrl string
}

func IndexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tmpl, _ := template.New("").ParseFiles("web/templates/index.html", "web/templates/base.html")
	_ = tmpl.ExecuteTemplate(w, "base", nil)
}

func ExpandHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func ShortenHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
		stmt, err := connections.Db.Prepare("INSERT INTO urls (tag, url) values (?, ?)")

		if err != nil {
			log.Println(err)
		}
		_, err = stmt.Exec(tag, url)

		if err != nil {
			log.Println(err)
		}
	}

	return fmt.Sprintf("http://%s/%s", configuration.AppConfig.ExternalDomain, tag)
}

// cachedGetUrl will check redis for url by tag, then db. If found in db, update the cache.
func cachedGetUrl(tag string) string {

	url, _ := connections.Cache.Get(fmt.Sprintf("urls:%s", tag)).Result()

	// Check db
	if url == "" {
		rows, err := connections.Db.Query("SELECT url from urls where tag=?", tag)

		if err != nil {
			log.Println(err)
		}

		if rows.Next() {
			rows.Scan(&url)

			// update cache
			connections.Cache.Set(fmt.Sprintf("urls:%s", tag), url, 0)
		}
		rows.Close()
	}

	return url
}
