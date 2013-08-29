package main

import (
	"code.google.com/p/gorilla/mux"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	godis "github.com/simonz05/godis/redis"
	"github.com/fs111/simpleconfig"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
	"crypto/rand"
)

const (
	// special key in redis, that is our global counter
	COUNTER = "__counter__"
	HTTP    = "http"
	ROLL    = "http://localhost:9999/index.htm"
	alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var (
	redis  *godis.Client
	config *simpleconfig.Config
)

type KurzUrl struct {
	Key          string
	ShortUrl     string
	LongUrl      string
	EventId       string
	UserId       string
	Type         string
	CreationDate int64
	Clicks       int64
}

// Converts the KurzUrl to JSON.
func (k KurzUrl) Json() []byte {
	b, _ := json.Marshal(k)
	return b
}

// Creates a new KurzUrl instance. The Given key, shorturl and longurl will
// be used. Clicks will be set to 0 and CreationDate to time.Nanoseconds()
func NewKurzUrl(key, shorturl, longurl, eventid, user, etype string) *KurzUrl {
	kurl := new(KurzUrl)
	kurl.CreationDate = time.Now().UnixNano()
	kurl.Key = key
	kurl.LongUrl = longurl
	kurl.ShortUrl = shorturl
	kurl.EventId = eventid
	kurl.UserId = user
	kurl.Type = etype
	kurl.Clicks = 0
	return kurl
}

// stores a new KurzUrl for the given key, shorturl and longurl. Existing
// ones with the same url will be overwritten
func store(key, shorturl, longurl, eventid, user, etype string) *KurzUrl {
	kurl := NewKurzUrl(key, shorturl, longurl, eventid, user, etype)
	go redis.Hset(kurl.Key, "LongUrl", kurl.LongUrl)
	go redis.Hset(kurl.Key, "EventId", kurl.EventId)
	go redis.Hset(kurl.Key, "UserId", kurl.UserId)
	go redis.Hset(kurl.Key, "Type", kurl.Type)
	go redis.Hset(kurl.Key, "ShortUrl", kurl.ShortUrl)
	go redis.Hset(kurl.Key, "CreationDate", kurl.CreationDate)
	go redis.Hset(kurl.Key, "Clicks", kurl.Clicks)
	return kurl
}

// loads a KurzUrl instance for the given key. If the key is
// not found, os.Error is returned.
func load(key string) (*KurzUrl, error) {
	if ok, _ := redis.Hexists(key, "ShortUrl"); ok {
		kurl := new(KurzUrl)
		kurl.Key = key
		reply, _ := redis.Hmget(key, "LongUrl", "EventId", "UserId", "Type", "ShortUrl", "CreationDate", "Clicks")
		kurl.LongUrl, kurl.EventId, kurl.UserId, kurl.Type, kurl.ShortUrl, kurl.CreationDate, kurl.Clicks =
			reply.Elems[0].Elem.String(), reply.Elems[1].Elem.String(),
 			reply.Elems[2].Elem.String(), reply.Elems[3].Elem.String(),
 			reply.Elems[4].Elem.String(),
			reply.Elems[5].Elem.Int64(), reply.Elems[6].Elem.Int64()
		return kurl, nil
	}
	return nil, errors.New("unknown key: " + key)
}

func fileExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

// function to display the info about a KurzUrl given by it's Key
func info(w http.ResponseWriter, r *http.Request) {
	short := mux.Vars(r)["short"]
	if strings.HasSuffix(short, "+") {
		short = strings.Replace(short, "+", "", 1)
	}

	kurl, err := load(short)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(kurl.Json())
		io.WriteString(w, "\n")
	} else {
		http.Redirect(w, r, ROLL, http.StatusNotFound)
	}
}

// function to resolve a shorturl and redirect
func resolve(w http.ResponseWriter, r *http.Request) {
	short := mux.Vars(r)["short"]
	kurl, err := load(short)
	fmt.Println(kurl)
	fmt.Println(err)
	if err == nil {
		go redis.Hincrby(kurl.Key, "Clicks", 1)
		newClick(kurl.UserId, kurl.EventId, kurl.Type)
		//http.Redirect(w, r, kurl.LongUrl, http.StatusMovedPermanently)
		http.Redirect(w, r, kurl.LongUrl, http.StatusTemporaryRedirect)
	} else {
		http.Redirect(w, r, ROLL, http.StatusTemporaryRedirect)
	}
}

// Determines if the string rawurl is a valid URL to be stored.
func isValidUrl(rawurl string) (u *url.URL, err error) {
	if len(rawurl) == 0 {
		return nil, errors.New("empty url")
	}
	// XXX this needs some love...
	if !strings.HasPrefix(rawurl, HTTP) {
		rawurl = fmt.Sprintf("%s://%s", HTTP, rawurl)
	}
	return url.Parse(rawurl)
}

func updateUser(userId string, etype string) {
    key := "user_" + userId
    if ok, _ := redis.Hexists(key, "InviteCount"); ok {
	if etype == "invite"{
	    go redis.Hincrby(key, "InviteCount", 1)
	} else if etype == "share"{
	    go redis.Hincrby(key, "ShareCount", 1)
	} else if etype == "attend"{
	    go redis.Hincrby(key, "AttendCount", 1)
	}
    } else {
	go redis.Hset(key, "InviteCount", 0)
	go redis.Hset(key, "InviteClicks", 0)
	go redis.Hset(key, "ShareCount", 0)
	go redis.Hset(key, "ShareClicks", 0)
	go redis.Hset(key, "AttendCount", 0)
	go redis.Hset(key, "AttendClicks", 0)
    }
}

func newUrl(userId string, eventId string, etype string) {
    userKey := "user_" + userId
    newUrlUpdate(userKey, etype)
    eventKey := "event_" + eventId
    newUrlUpdate(eventKey, etype)
}

func newClick(userId string, eventId string, etype string) {
    userKey := "user_" + userId
    newClickUpdate(userKey, etype)
    eventKey := "event_" + eventId
    newClickUpdate(eventKey, etype)
}

func newClickUpdate(key, etype string) {
    if etype == "invite"{
        go redis.Hincrby(key, "InviteClicks", 1)
    } else if etype == "share"{
        go redis.Hincrby(key, "ShareClicks", 1)
    } else if etype == "attend"{
        go redis.Hincrby(key, "AttendClicks", 1)
    }
}

func newUrlUpdate(key, etype string) {
    if ok, _ := redis.Hexists(key, "InviteCount"); ok {
	if etype == "invite"{
	    go redis.Hincrby(key, "InviteCount", 1)
	} else if etype == "share"{
	    go redis.Hincrby(key, "ShareCount", 1)
	} else if etype == "attend"{
	    go redis.Hincrby(key, "AttendCount", 1)
	}
    } else {
	go redis.Hset(key, "InviteCount", 1)
	go redis.Hset(key, "InviteClicks", 0)
	go redis.Hset(key, "ShareCount", 1)
	go redis.Hset(key, "ShareClicks", 0)
	go redis.Hset(key, "AttendCount", 1)
	go redis.Hset(key, "AttendClicks", 0)
    }
}

// function to shorten and store a url
func shorten(w http.ResponseWriter, r *http.Request) {
	host := config.GetStringDefault("hostname", "localhost")
	leUrl := r.FormValue("url")
	fmt.Println(leUrl)
	eventId := r.FormValue("eventid")
	theUrl, err := isValidUrl(string(leUrl))
	userId := r.FormValue("user")
	fmt.Println(userId)
	etype := r.FormValue("type")
	if err == nil {
		//ctr, _ := redis.Incr(COUNTER)
		//encoded := Encode(ctr)
		encoded := getUrl()
		location := fmt.Sprintf("%s://%s/%s", HTTP, host, encoded)
		fmt.Println(location)
		kurl := store(encoded, location, theUrl.String(), eventId, userId, etype)
		newUrl(userId, eventId, etype)
		w.Header().Set("Content-Type", "application/json")
		w.Write(kurl.Json())
		io.WriteString(w, "\n")
	} else {
		http.Redirect(w, r, ROLL, http.StatusNotFound)
	}
}

func getUrl() string {
	bytes := make([]byte, 5)
	for {
		rand.Read(bytes)
		for i, b := range bytes {
			bytes[i] = alphanum[b%byte(len(alphanum))]
		}
		id := string(bytes)
		fmt.Println(id)
		if ok, _ := redis.Hexists(id, "ShortUrl"); !ok {
		    return id
		}
	}
}

func userStats(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	fmt.Println(id)
	c, _ := redis.Get(COUNTER)
	last := c.Int64()
	w.Header().Set("Content-Type", "application/json")
	stats := make(map[string]map[string]int64)
        invites := make(map[string]int64)
        shares := make(map[string]int64)
        attends := make(map[string]int64)
	for i := last; i > 0; i -= 1 {
	    kurl, err := load(Encode(i))
    	    if err == nil {
		if kurl.UserId == id {
		    if kurl.Type == "invite"{
			invites[kurl.LongUrl] = kurl.Clicks
		    }
		    if kurl.Type == "share"{
			shares[kurl.LongUrl] = kurl.Clicks
		    }
		    if kurl.Type == "attend"{
			attends[kurl.LongUrl] = kurl.Clicks
		    }
		}
	    }
	}
	stats["invites"] = invites
	stats["shares"] = shares
	stats["attends"] = attends
	fmt.Println(stats)
	s, _ := json.Marshal(stats)
	w.Write(s)
}

func eventStats(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	fmt.Println(url)
	c, _ := redis.Get(COUNTER)
	last := c.Int64()
	w.Header().Set("Content-Type", "application/json")
	stats := make(map[string]int64)
        var invitesCount int64 = 0
        var invitesClicks int64 = 0
        var sharesCount int64 = 0
        var sharesClicks int64 = 0
        var attendsCount int64 = 0
        var attendsClicks int64 = 0
	for i := last; i > 0; i -= 1 {
	    kurl, err := load(Encode(i))
    	    if err == nil {
		if kurl.LongUrl == url {
		    if kurl.Type == "invite"{
			invitesCount += 1
			invitesClicks += kurl.Clicks
		    }
		    if kurl.Type == "share"{
			sharesCount += 1
			sharesClicks += kurl.Clicks
		    }
		    if kurl.Type == "attend"{
			attendsCount += 1
			attendsClicks += kurl.Clicks
		    }
		}
	    }
	}
	stats["invitesCount"] = invitesCount
	stats["sharesCount"] = sharesCount
	stats["attendsCount"] = attendsCount
	stats["invitesClicks"] = invitesClicks
	stats["sharesClicks"] = sharesClicks
	stats["attendsClicks"] = attendsClicks
	fmt.Println(stats)
	s, _ := json.Marshal(stats)
	w.Write(s)
}


//Returns a json array with information about the last shortened urls. If data 
// is a valid integer, that's the amount of data it will return, otherwise
// a maximum of 10 entries will be returned.
func latest(w http.ResponseWriter, r *http.Request) {
	// TODO: currently it just returns all keys
	w.Header().Set("Content-Type", "application/json")
	var kurls = []*KurzUrl{}
	keys, _ := redis.Keys("*")
	for i := 0; i < len(keys); i += 1 {
		key := keys[i]
		kurl, err := load(key)
		if err == nil {
			kurls = append(kurls, kurl)
		}
	}
	s, _ := json.Marshal(kurls)
	w.Write(s)
}

func static(w http.ResponseWriter, r *http.Request) {
	fname := mux.Vars(r)["fileName"]
	// empty means, we want to serve the index file. Due to a bug in http.serveFile
	// the file cannot be called index.html, anything else is fine.
	if fname == "" {
		fname = "index.htm"
	}
	staticDir := config.GetStringDefault("static-directory", "")
	staticFile := path.Join(staticDir, fname)
	if fileExists(staticFile) {
		http.ServeFile(w, r, staticFile)
	}
}

func main() {
	flag.Parse()
	path := flag.Arg(0)

	config, _ = simpleconfig.NewConfig(path)

	host := config.GetStringDefault("redis.address", "tcp:localhost:6379")
	db := config.GetIntDefault("redis.database", 0)
	passwd := config.GetStringDefault("redis.password", "")

	redis = godis.New(host, db, passwd)

	router := mux.NewRouter()
	router.HandleFunc("/shorten/{url:(.*$)}", shorten)

	router.HandleFunc("/{short:([a-zA-Z0-9]+$)}", resolve)
	router.HandleFunc("/{short:([a-zA-Z0-9]+)\\+$}", info)
	router.HandleFunc("/info/{short:[a-zA-Z0-9]+}", info)
	router.HandleFunc("/latest/{data:[0-9]+}", latest)
	router.HandleFunc("/user/{id:(.*$)}", userStats)
	router.HandleFunc("/event/{url:(.*$)}", eventStats)

	router.HandleFunc("/{fileName:(.*$)}", static)

	listen := config.GetStringDefault("listen", "0.0.0.0")
	//listen := "192.168.1.13"
	port := config.GetStringDefault("port", "9999")
	s := &http.Server{
		Addr:    listen + ":" + port,
		Handler: router,
	}
	s.ListenAndServe()
}
