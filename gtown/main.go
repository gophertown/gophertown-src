package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
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
	Slack     string `json:"slack"`
	Reddit    string `json:"reddit"`
	Twitter   string `json:"twitter"`
	Blog      string `json:"blog"`
	Website   string `json:"website"`
	Notes     string `json:"notes"`
	Avatar    string `json:"avatar"`
}

const contentTypeJSON = "application/json"

func (u user) keywords() []byte {

	terms := make(map[string]struct{})

	terms[u.Username] = struct{}{}
	terms[u.Name] = struct{}{}
	terms[u.IRC] = struct{}{}
	terms[u.Email] = struct{}{}
	terms[u.Discourse] = struct{}{}
	terms[u.Slack] = struct{}{}
	terms[u.Reddit] = struct{}{}
	terms[u.Twitter] = struct{}{}
	terms[u.Blog] = struct{}{}
	terms[u.Website] = struct{}{}
	terms[u.Notes] = struct{}{}

	b := bytes.Buffer{}
	for k := range terms {
		b.WriteString(k)
		b.WriteByte(' ')
	}

	return b.Bytes()
}

func userHandler(w http.ResponseWriter, r *http.Request, users map[string]user) {

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

func randomHandler(w http.ResponseWriter, r *http.Request, users map[string]user, usernames []string) {
	n := rand.Intn(len(usernames))
	u := users[usernames[n]]
	w.Header().Set("Content-Type", contentTypeJSON)
	jenc := json.NewEncoder(w)
	jenc.Encode([]user{u})
}

func searchHandler(w http.ResponseWriter, r *http.Request, users map[string]user, usernames []string, searchIndex *suffixarray.Index, offsets []int) {

	q := r.FormValue("for")

	var us []user
	seen := make(map[int]bool)

	idxs := searchIndex.Lookup([]byte(q), -1)
	for _, idx := range idxs {
		i := sort.Search(len(offsets), func(i int) bool { return offsets[i] > idx })
		if idx+len(q) < offsets[i] && !seen[i] {
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

	var users map[string]user = make(map[string]user)
	var usernames []string
	var offsets []int

	flag.Parse()

	var searchData []byte

	dataFiles, err := filepath.Glob(*gopherdir + "/*.json")
	if err != nil {
		log.Fatalf("unable to load any gophers: %v", err)
	}

	for _, f := range dataFiles {
		ujs, err := ioutil.ReadFile(f)
		if err != nil {
			log.Println(f, err)
			continue
		}

		var u user

		err = json.Unmarshal(ujs, &u)
		if err != nil {
			log.Println(f, err)
			continue
		}

		name := strings.TrimSuffix(path.Base(f), ".json")

		usernames = append(usernames, name)
		u.Username = name
		u.Avatar = "https://avatars.githubusercontent.com/" + name
		// TODO(dgryski): respect show_avatar
		// TODO(dgryski): u.Note markdown -> html
		// TODO(dgryski): handle IRC channels

		users[name] = u

		// update search index
		searchData = append(searchData, u.keywords()...)
		offsets = append(offsets, len(searchData))
	}

	log.Println("loaded", len(users), "gophers")
	searchIndex := suffixarray.New(searchData)

	if p := os.Getenv("PORT"); p != "" {
		*port, _ = strconv.Atoi(p)
	}

	http.HandleFunc("/gophers/user", func(w http.ResponseWriter, r *http.Request) { userHandler(w, r, users) })
	http.HandleFunc("/gophers/random", func(w http.ResponseWriter, r *http.Request) { randomHandler(w, r, users, usernames) })
	http.HandleFunc("/gophers/search", func(w http.ResponseWriter, r *http.Request) {
		searchHandler(w, r, users, usernames, searchIndex, offsets)
	})

	// The static files we know about
	http.Handle("/static/", http.FileServer(http.Dir(*sitedir)))

	// Everything else serves up the single-page app -- routing etc are handled on the client side.
	indexhtml := *sitedir + "/index.html"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, indexhtml) })

	log.Println("listening on port", *port)
	log.Fatalln(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
