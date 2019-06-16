package handlers

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/codebynumbers/go-shorty/internal/configuration"
	"github.com/go-redis/redis"
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

type Cacheable interface {
	Sleep()
}

type HandlerEnv struct {
	AppConfig configuration.Config
	Db        *sql.DB
	Cache     *redis.Client
}

func (env *HandlerEnv) IndexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tmpl, _ := template.New("").ParseFiles("web/templates/index.html", "web/templates/base.html")
	_ = tmpl.ExecuteTemplate(w, "base", nil)
}

func (env *HandlerEnv) ExpandHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	tag := ps.ByName("tag")

	url, err := env.cachedGetUrl(tag)

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("500 Internal Server Error"))
		return
	}

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

func (env *HandlerEnv) ShortenHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()
	url := r.Form.Get("url")
	if url != "" {

		shortened, err := env.shorten(url)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("500 Internal Server Error"))
			return
		}

		context := ResultPageData{
			ShortenedUrl: shortened,
		}

		tmpl, _ := template.New("").ParseFiles("web/templates/result.html", "web/templates/base.html")
		_ = tmpl.ExecuteTemplate(w, "base", context)
	}
}

// shorten encodes the url and returns a new url to reach it at
func (env *HandlerEnv) shorten(url string) (string, error) {
	tag := generateHash(url)

	cachedUrl, err := env.cachedGetUrl(tag)
	if err != nil {
		return "", err
	}

	if cachedUrl == "" {
		stmt, err := env.Db.Prepare("INSERT INTO urls (tag, url) values (?, ?)")

		if err != nil {
			log.Println(err)
			return "", err
		}
		_, err = stmt.Exec(tag, url)

		if err != nil {
			log.Println(err)
			return "", err
		}
	}

	return fmt.Sprintf("http://%s/%s", env.AppConfig.ExternalDomain, tag), nil
}

func generateHash(url string) (string) {
	hash := fnv.New32a()
	hash.Write([]byte(url))
	return hex.EncodeToString(hash.Sum(nil))
}

// cachedGetUrl will check redis for url by tag, then db. If found in db, update the cache.
func (env *HandlerEnv) cachedGetUrl(tag string) (string, error) {

	url, _ := env.Cache.Get(fmt.Sprintf("urls:%s", tag)).Result()

	// Check db
	if url == "" {
		rows, err := env.Db.Query("SELECT url from urls where tag=?", tag)

		if err != nil {
			log.Println(err)
			return "", err
		}

		if rows.Next() {
			rows.Scan(&url)

			// update cache
			env.Cache.Set(fmt.Sprintf("urls:%s", tag), url, 0)
		}
		rows.Close()
	}

	return url, nil
}
