package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	c "github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	badgerstore "github.com/ostafen/clover/v2/store/badger"
)

var msgDB *c.DB

func init() {
	str, err := badgerstore.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		log.Fatal(err)
	}

	db, err := c.OpenWithStore(str)
	if err != nil {
		log.Fatal(err)
	}
	msgDB = db

	if err := msgDB.CreateCollection("messages"); err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer msgDB.Close()

	//send message endpoint
	// mux := http.NewServeMux()
	// mux.HandleFunc("Post /sent", HandleSend)
	mux := chi.NewMux()
	mux.Use(cors.Handler(cors.Options{}))
	mux.Post("/sent", HandleSend)

	srv := http.Server{
		Addr:                         ":8989",
		Handler:                      mux,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    &tls.Config{},
		ReadTimeout:                  2 * time.Second,
		ReadHeaderTimeout:            2 * time.Second,
		WriteTimeout:                 2 * time.Second,
		IdleTimeout:                  10 * time.Second,
		MaxHeaderBytes:               http.DefaultMaxHeaderBytes,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func HandleSend(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	msgMap := map[string]any{}
	if err := json.NewDecoder(r.Body).Decode(&msgMap); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//save the message
	now := time.Now()
	doc := document.NewDocument()
	doc.SetAll(msgMap)
	doc.Set("created", now)
	id, err := msgDB.InsertOne("messages", doc)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to send message", 500)
		return
	}

	w.Write([]byte(id))
}

func HandleReceive(w http.ResponseWriter, r *http.Request) {
	
}
