package main

import (
	"encoding/json"
	"flag"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type user struct {
	Username  string `json:"username"`
	Name      string `json:"name"`
	IRC       string `json:"irc"`
	Email     string `json:"email"`
	Discourse string `json:"discourse"`
	Reddit    string `json:"reddit"`
	Twitter   string `json:"twitter"`
	Blog      string `json:"blog"`
	Website   string `json:"website"`
	Notes     string `json:"notes"`
	Avatar    string `json:"avatar"`
}

const contentTypeJSON = "application/json"

var users map[string]user = make(map[string]user)
var usernames []string
var searchIndex *suffixarray.Index
var offsets []int

func (u user) keywords() string {

	var k []string

	k = append(k, u.Username)

	if u.Name != "" {
		k = append(k, u.Name)
	}
	if u.IRC != "" {
		k = append(k, u.IRC)
	}
	if u.Email != "" {
		k = append(k, u.Email)
	}
	if u.Discourse != "" {
		k = append(k, u.Discourse)
	}
	if u.Reddit != "" {
		k = append(k, u.Reddit)
	}
	if u.Twitter != "" {
		k = append(k, u.Twitter)
	}
	if u.Blog != "" {
		k = append(k, u.Blog)
	}
	if u.Website != "" {
		k = append(k, u.Website)
	}
	if u.Notes != "" {
		k = append(k, u.Notes)
	}

	return strings.Join(k, " ")
}

func userHandler(w http.ResponseWriter, r *http.Request) {

	name := r.FormValue("username")

	u, ok := users[name]

	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)

	jenc := json.NewEncoder(w)
	jenc.Encode([]user{u})
}

func randomHandler(w http.ResponseWriter, r *http.Request) {
	n := rand.Intn(len(usernames))
	u := users[usernames[n]]
	w.Header().Set("Content-Type", contentTypeJSON)
	jenc := json.NewEncoder(w)
	jenc.Encode([]user{u})
}

func searchHandler(w http.ResponseWriter, r *http.Request) {

	q := r.FormValue("for")

	var us []user
	seen := make(map[int]bool)

	idxs := searchIndex.Lookup([]byte(q), -1)
	for _, idx := range idxs {
		i := sort.Search(len(offsets), func(i int) bool { return offsets[i] > idx })
		if !seen[i] {
			us = append(us, users[usernames[i]])
			seen[i] = true
		}
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	jenc := json.NewEncoder(w)
	jenc.Encode(us)
}

func main() {

	gopherdir := flag.String("gopherdir", "", "gopher json files")
	sitedir := flag.String("site", ".", "site to serve")
	port := flag.Int("port", 8080, "port")

	flag.Parse()

	var searchData []byte

	filepath.Walk(*gopherdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(info.Name(), ".json") || info.Name() == "template.json" {
			return nil
		}

		ujs, err := ioutil.ReadFile(path)
		if err != nil {
			log.Println(path, err)
			return nil
		}

		var u user

		err = json.Unmarshal(ujs, &u)
		if err != nil {
			log.Println(path, err)
			return nil
		}

		name := strings.TrimSuffix(info.Name(), ".json")

		usernames = append(usernames, name)
		u.Username = name
		u.Avatar = "https://avatars.githubusercontent.com/" + name
		// TODO(dgryski): respect show_avatar
		// TODO(dgryski): u.Note markdown -> html
		// TODO(dgryski): handle IRC channels

		users[name] = u

		// update search index
		searchData = append(searchData, []byte(u.keywords())...)
		offsets = append(offsets, len(searchData))

		return nil
	})

	log.Println("loaded", len(users), "gophers")
	searchIndex = suffixarray.New(searchData)

	if p := os.Getenv("PORT"); p != "" {
		*port, _ = strconv.Atoi(p)
	}

	http.HandleFunc("/gophers/user", userHandler)
	http.HandleFunc("/gophers/random", randomHandler)
	http.HandleFunc("/gophers/search", searchHandler)

	// everything else comes from the site
	http.Handle("/", http.FileServer(http.Dir(*sitedir)))

	log.Println("listening on port", *port)
	log.Fatalln(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
