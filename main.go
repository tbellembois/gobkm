package main

//go:generate rice embed-go

import (
	"flag"
	"net/http"
	"net/url"
	"os"

	rice "github.com/GeertJohan/go.rice"
	log "github.com/sirupsen/logrus"
	"github.com/tbellembois/gobkm/handlers"
	"github.com/tbellembois/gobkm/models"

	"github.com/justinas/alice"
	"github.com/rs/cors"
)

var (
	datastore   *models.SQLiteDataStore
	templateBox *rice.Box
	err         error
	logf        *os.File
)

func main() {
	// Getting the program parameters.
	listenPort := flag.String("port", "8081", "the port to listen")
	proxyURL := flag.String("proxy", "http://localhost:"+*listenPort, "the proxy full URL if used")
	historySize := flag.Int("history", 3, "the folder history size")
	username := flag.String("username", "", "the default login username")
	dbPath := flag.String("db", "bkm.db", "the full sqlite db path")
	logfile := flag.String("logfile", "", "log to the given file")
	debug := flag.Bool("debug", false, "debug (verbose log), default is error")
	flag.Parse()

	// Logging to file if logfile parameter specified.
	if *logfile != "" {
		if logf, err = os.OpenFile(*logfile, os.O_WRONLY|os.O_CREATE, 0755); err != nil {
			log.Panic(err)
		} else {
			log.SetOutput(logf)
		}
	}
	// Setting the log level.
	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}
	log.WithFields(log.Fields{
		"listenPort":  *listenPort,
		"proxyURL":    *proxyURL,
		"historySize": *historySize,
		"username":    *username,
		"logfile":     *logfile,
		"debug":       *debug,
	}).Debug("main:flags")

	// Database initialization.
	if datastore, err = models.NewDBstore(*dbPath); err != nil {
		log.Panic(err)
	}
	// Database creation.
	datastore.CreateDatabase()
	datastore.PopulateDatabase()
	// Error check.
	if datastore.FlushErrors() != nil {
		log.Panic(err)
	}

	// host from URL
	u, err := url.Parse(*proxyURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Debug(u)

	// Environment creation.
	env := handlers.Env{
		DB:               datastore,
		GoBkmProxyURL:    *proxyURL,
		GoBkmProxyHost:   u.Host,
		GoBkmHistorySize: *historySize,
		GoBkmUsername:    *username,
	}
	// Building a rice box with the static directory.
	if templateBox, err = rice.FindBox("static"); err != nil {
		log.Fatal(err)
	}
	// Getting the HTML template files content as a string.
	if env.TplMainData, err = templateBox.String("index.html"); err != nil {
		log.Fatal(err)
	}

	// CORS handler
	c := cors.New(cors.Options{
		Debug:            true,
		AllowedOrigins:   []string{"http://localhost:8081", *proxyURL},
		AllowCredentials: true,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"Authorization", "DNT", "User-Agent", "X-Requested-With", "If-Modified-Since", "Cache-Control", "Content-Type", "Range"},
	})

	mux := http.NewServeMux()

	// Handlers initialization.
	mux.HandleFunc("/addBookmark/", env.AddBookmarkHandler)
	mux.HandleFunc("/addFolder/", env.AddFolderHandler)
	mux.HandleFunc("/deleteBookmark/", env.DeleteBookmarkHandler)
	mux.HandleFunc("/deleteFolder/", env.DeleteFolderHandler)
	mux.HandleFunc("/getTags/", env.GetTagsHandler)
	mux.HandleFunc("/getStars/", env.GetStarsHandler)
	mux.HandleFunc("/getFolderChildren/", env.GetFolderChildrenHandler)
	mux.HandleFunc("/getTree/", env.GetTreeHandler)
	mux.HandleFunc("/import/", env.ImportHandler)
	mux.HandleFunc("/export/", env.ExportHandler)
	mux.HandleFunc("/updateFolder/", env.UpdateFolderHandler)
	mux.HandleFunc("/updateBookmark/", env.UpdateBookmarkHandler)
	mux.HandleFunc("/searchBookmarks/", env.SearchBookmarkHandler)
	mux.HandleFunc("/starBookmark/", env.StarBookmarkHandler)
	mux.HandleFunc("/", env.MainHandler)

	waBox := rice.MustFindBox("static/wasm")
	waFileServer := http.StripPrefix("/wasm/", http.FileServer(waBox.HTTPBox()))
	mux.Handle("/wasm/", waFileServer)

	chain := alice.New(c.Handler).Then(mux)

	if err = http.ListenAndServe(":"+*listenPort, chain); err != nil {
		log.Fatal(err)
	}
}
