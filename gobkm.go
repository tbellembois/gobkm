package main

//go:generate gopherjs build gopherjs/gjs-common.go -o static/js/gjs-common.js -m
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
		"listenPort": *listenPort,
		"proxyURL":   *proxyURL,
		"logfile":    *logfile,
		"debug":      *debug,
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
		DB:             datastore,
		GoBkmProxyURL:  *proxyURL,
		GoBkmProxyHost: u.Host,
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
		AllowedOrigins:   []string{"http://localhost:8080", *proxyURL},
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
	mux.HandleFunc("/getBookmarkTags/", env.GetBookmarkTagsHandler)
	mux.HandleFunc("/getTags/", env.GetTagsHandler)
	mux.HandleFunc("/getStars/", env.GetStarsHandler)
	mux.HandleFunc("/getBranchNodes/", env.GetBranchNodesHandler)
	mux.HandleFunc("/getTree/", env.GetTreeHandler)
	mux.HandleFunc("/import/", env.ImportHandler)
	mux.HandleFunc("/export/", env.ExportHandler)
	mux.HandleFunc("/moveFolder/", env.MoveFolderHandler)
	mux.HandleFunc("/moveBookmark/", env.MoveBookmarkHandler)
	mux.HandleFunc("/renameFolder/", env.RenameFolderHandler)
	mux.HandleFunc("/renameBookmark/", env.RenameBookmarkHandler)
	mux.HandleFunc("/searchBookmarks/", env.SearchBookmarkHandler)
	mux.HandleFunc("/starBookmark/", env.StarBookmarkHandler)
	mux.HandleFunc("/", env.MainHandler)

	// Rice boxes initialization.
	// Awesome fonts may need to send the Access-Control-Allow-Origin header to "*"
	cssBox := rice.MustFindBox("static/css")
	cssFileServer := http.StripPrefix("/css/", http.FileServer(cssBox.HTTPBox()))
	mux.Handle("/css/", cssFileServer)

	jsBox := rice.MustFindBox("static/js")
	jsFileServer := http.StripPrefix("/js/", http.FileServer(jsBox.HTTPBox()))
	mux.Handle("/js/", jsFileServer)

	imgBox := rice.MustFindBox("static/img")
	imgFileServer := http.StripPrefix("/img/", http.FileServer(imgBox.HTTPBox()))
	mux.Handle("/img/", imgFileServer)

	fontsBox := rice.MustFindBox("static/fonts")
	fontsFileServer := http.StripPrefix("/fonts/", http.FileServer(fontsBox.HTTPBox()))
	mux.Handle("/fonts/", fontsFileServer)

	manifestBox := rice.MustFindBox("static/manifest")
	manifestFileServer := http.StripPrefix("/manifest/", http.FileServer(manifestBox.HTTPBox()))
	mux.Handle("/manifest/", manifestFileServer)

	chain := alice.New(c.Handler).Then(mux)

	if err = http.ListenAndServe(":"+*listenPort, chain); err != nil {
		log.Fatal(err)
	}
}
