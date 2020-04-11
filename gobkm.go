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

	// Handlers initialization.
	http.HandleFunc("/addBookmark/", env.AddBookmarkHandler)
	http.HandleFunc("/addFolder/", env.AddFolderHandler)
	http.HandleFunc("/deleteBookmark/", env.DeleteBookmarkHandler)
	http.HandleFunc("/deleteFolder/", env.DeleteFolderHandler)
	http.HandleFunc("/getBookmarkTags/", env.GetBookmarkTagsHandler)
	http.HandleFunc("/getTags/", env.GetTagsHandler)
	http.HandleFunc("/getStars/", env.GetStarsHandler)
	http.HandleFunc("/getBranchNodes/", env.GetBranchNodesHandler)
	http.HandleFunc("/getTree/", env.GetTreeHandler)
	http.HandleFunc("/import/", env.ImportHandler)
	http.HandleFunc("/export/", env.ExportHandler)
	http.HandleFunc("/moveFolder/", env.MoveFolderHandler)
	http.HandleFunc("/moveBookmark/", env.MoveBookmarkHandler)
	http.HandleFunc("/renameFolder/", env.RenameFolderHandler)
	http.HandleFunc("/renameBookmark/", env.RenameBookmarkHandler)
	http.HandleFunc("/searchBookmarks/", env.SearchBookmarkHandler)
	http.HandleFunc("/starBookmark/", env.StarBookmarkHandler)
	http.HandleFunc("/", env.MainHandler)

	// Rice boxes initialization.
	// Awesome fonts may need to send the Access-Control-Allow-Origin header to "*"
	cssBox := rice.MustFindBox("static/css")
	cssFileServer := http.StripPrefix("/css/", http.FileServer(cssBox.HTTPBox()))
	http.Handle("/css/", cssFileServer)

	jsBox := rice.MustFindBox("static/js")
	jsFileServer := http.StripPrefix("/js/", http.FileServer(jsBox.HTTPBox()))
	http.Handle("/js/", jsFileServer)

	imgBox := rice.MustFindBox("static/img")
	imgFileServer := http.StripPrefix("/img/", http.FileServer(imgBox.HTTPBox()))
	http.Handle("/img/", imgFileServer)

	fontsBox := rice.MustFindBox("static/fonts")
	fontsFileServer := http.StripPrefix("/fonts/", http.FileServer(fontsBox.HTTPBox()))
	http.Handle("/fonts/", fontsFileServer)

	manifestBox := rice.MustFindBox("static/manifest")
	manifestFileServer := http.StripPrefix("/manifest/", http.FileServer(manifestBox.HTTPBox()))
	http.Handle("/manifest/", manifestFileServer)

	if err = http.ListenAndServe(":"+*listenPort, nil); err != nil {
		log.Fatal(err)
	}
}
