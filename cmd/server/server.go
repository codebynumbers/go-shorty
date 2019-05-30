package main

import (
  "fmt"
  "log"
  "math"
  "net/http"
  "strings"
)

// Config
var port = "3000"

var numerals = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
var next_id = 1000
var urls = make(map[string]string)
var tags = make(map[string]string)

func shortenHandler(w http.ResponseWriter, r *http.Request) {
  url, ok := r.URL.Query()["url"]
  if (ok) {
    w.Write([]byte(shorten(url[0])))
  }
}

func lengthenHandler(w http.ResponseWriter, r *http.Request) {
  tag, t_ok := r.URL.Query()["tag"]
  if (t_ok) {
    url, u_ok := urls[tag[0]]
    if (u_ok) {
      w.Write([]byte(url))
    }
  }
}

func main() {
  mux := http.NewServeMux()
  mux.Handle("/shorten", http.HandlerFunc(shortenHandler))
  mux.Handle("/lengthen", http.HandlerFunc(lengthenHandler))
  log.Println(fmt.Sprintf("Listening on %s...", port))
  http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
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
    tag, ok := tags[url]

    if !ok {
        next_id += 1
        tag = baseN(next_id)

        // setup 2-way lookup
        urls[tag] = url
        tags[url] = tag
    }
    return fmt.Sprintf("http://short.ly/%s", tag)
}

