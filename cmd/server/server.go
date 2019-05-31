package main

import (
  "fmt"
  "github.com/go-redis/redis"
  "github.com/julienschmidt/httprouter"
  "log"
  "math"
  "net/http"
  "strings"
)

// Config
var port = "3000"
var domain = fmt.Sprintf("localhost:%s", port)

var numerals = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
var next_id = 1000

var client = redis.NewClient(&redis.Options{
    Addr:     "localhost:6379", // use default Addr
    Password: "",               // no password set
    DB:       0,                // use default DB
})

func shortenHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
  url, ok := r.URL.Query()["url"]
  if (ok) {
    w.Write([]byte(shorten(url[0])))
  }
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
     log.Println(url)
     if !strings.HasPrefix(url, "http") {
         url = "http://" + url
         log.Println(url)
     }
     http.Redirect(w, r, url, 302)
  }
}

func main() {
  router := httprouter.New()
  router.GET("/s/", shortenHandler)
  router.GET("/e/:tag", expandHandler)

  log.Println(fmt.Sprintf("Listening on %s...", port))
  http.ListenAndServe(fmt.Sprintf(":%s", port), router)
}


func baseN(num int) string {
    // Convert base 10 to base 62
    if num == 0 {
        return "0"
    }

    // must cast byte to string to concat
    return strings.TrimLeft(baseN(num / 62), "0") + string(numerals[num % 62])
}


func rebase(word string) int {
    // Convert from base62 back to base10
    power := len(word)-1
    sum := 0
    for _, char := range word {
        pos := strings.Index(numerals, string(char))
        sum += int(float64(pos) * math.Pow(float64(62), float64(power)))
        power -= 1
    }
    return sum
}

func shorten(url string) string {
    /* Check list, if new url, insert and bump id
       return "shortened" url
    */
    tag, err := client.Get(fmt.Sprintf("tags:%s", url)).Result()
    if err == redis.Nil {
        next_id += 1
        tag = baseN(next_id)

        client.Set(fmt.Sprintf("urls:%s", tag), url, 0).Err()
        client.Set(fmt.Sprintf("tags:%s", url), tag, 0).Err()

    } else if err != nil {
        panic(err)
    }

    return fmt.Sprintf("http://%s/e/%s", domain, tag)
}

