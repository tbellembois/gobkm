package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/GeertJohan/go.rice"
	log "github.com/Sirupsen/logrus"
	"github.com/tbellembois/gobkm/handlers"
	"github.com/tbellembois/gobkm/models"
)

const (
	dbURL = "./bkm.db"
)

var (
	datastore   *models.SQLiteDataStore
	templateBox *rice.Box
	err         error
	logf        *os.File
)

// A decorator to set custom HTTP headers.
func decoratedHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(rw, req)
	})
}

func main() {
	// Getting the program parameters.
	listenPort := flag.String("port", "8080", "the port to listen")
	goBkmProxyURL := flag.String("proxy", "http://localhost:"+*listenPort, "the proxy full URL if used")
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
		"goBkmURL":   *goBkmProxyURL,
		"logfile":    *logfile,
		"debug":      *debug,
	}).Debug("main:flags")

	// Database initialization.
	if datastore, err = models.NewDBstore(dbURL); err != nil {
		log.Panic(err)
	}
	// Database creation.
	datastore.CreateDatabase()
	datastore.PopulateDatabase()
	// Error check.
	if datastore.FlushErrors() != nil {
		log.Panic(err)
	}

	// Environment creation.
	env := handlers.Env{DB: datastore, GoBkmProxyURL: *goBkmProxyURL}
	// Building a rice box with the static directory.
	if templateBox, err = rice.FindBox("static"); err != nil {
		log.Fatal(err)
	}
	// Getting the HTML template files content as a string.
	if env.TplMainData, err = templateBox.String("main.html"); err != nil {
		log.Fatal(err)
	}
	if env.TplAddBookmarkData, err = templateBox.String("addBookmark.html"); err != nil {
		log.Fatal(err)
	}

	// Handlers initialization.
	http.HandleFunc("/getChildrenFolders/", env.GetChildrenFoldersHandler)
	http.HandleFunc("/getFolderBookmarks/", env.GetFolderBookmarksHandler)
	http.HandleFunc("/moveFolder/", env.MoveFolderHandler)
	http.HandleFunc("/moveBookmark/", env.MoveBookmarkHandler)
	http.HandleFunc("/renameFolder/", env.RenameFolderHandler)
	http.HandleFunc("/renameBookmark/", env.RenameBookmarkHandler)
	http.HandleFunc("/addFolder/", env.AddFolderHandler)
	http.HandleFunc("/addBookmark/", env.AddBookmarkHandler)
	http.HandleFunc("/deleteFolder/", env.DeleteFolderHandler)
	http.HandleFunc("/deleteBookmark/", env.DeleteBookmarkHandler)
	http.HandleFunc("/starBookmark/", env.StarBookmarkHandler)
	http.HandleFunc("/export/", env.ExportHandler)
	http.HandleFunc("/import/", env.ImportHandler)
	http.HandleFunc("/searchBookmarks/", env.SearchBookmarkHandler)
	// websocket handler
	http.HandleFunc("/socket/", env.SocketHandler)
	// bookmarklet handler
	http.HandleFunc("/bookmarkThis/", env.BookmarkThisHandler)
	//http.HandleFunc("/bookmarkThis2/", env.BookmarkThis2Handler)
	http.HandleFunc("/addBookmarkBookmarklet/", env.AddBookmarkBookmarkletHandler)
	http.HandleFunc("/", env.MainHandler)

	// Rice boxes initialization.
	// Awesome fonts may need to send the Access-Control-Allow-Origin header to "*"
	fontsBox := rice.MustFindBox("static/fonts")
	fontsFileServer := http.StripPrefix("/fonts/", decoratedHandler(http.FileServer(fontsBox.HTTPBox())))
	http.Handle("/fonts/", fontsFileServer)

	cssBox := rice.MustFindBox("static/css")
	cssFileServer := http.StripPrefix("/css/", http.FileServer(cssBox.HTTPBox()))
	http.Handle("/css/", cssFileServer)

	jsBox := rice.MustFindBox("static/js")
	jsFileServer := http.StripPrefix("/js/", http.FileServer(jsBox.HTTPBox()))
	http.Handle("/js/", jsFileServer)

	imgBox := rice.MustFindBox("static/img")
	imgFileServer := http.StripPrefix("/img/", http.FileServer(imgBox.HTTPBox()))
	http.Handle("/img/", imgFileServer)

	manifestBox := rice.MustFindBox("static/manifest")
	manifestFileServer := http.StripPrefix("/manifest/", http.FileServer(manifestBox.HTTPBox()))
	http.Handle("/manifest/", manifestFileServer)

	if err = http.ListenAndServe(":"+*listenPort, nil); err != nil {
		log.Fatal(err)
	}
}
