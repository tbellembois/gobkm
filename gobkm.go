package main

import (
	"flag"
	"net/http"

	"github.com/GeertJohan/go.rice"
	log "github.com/Sirupsen/logrus"
	"github.com/tbellembois/gobkm/handlers"
	"github.com/tbellembois/gobkm/models"
)

const (
	dbURL = "./bkm.db"
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
	debug := flag.Bool("debug", false, "debug (verbose log)")
	flag.Parse()

	log.WithFields(log.Fields{
		"listenPort": *listenPort,
		"goBkmURL":   *goBkmProxyURL,
	}).Debug("main:flags")

	// Setting the log level.
	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}

	// Database initialization.
	datastore, err := models.NewDBstore(dbURL)
	if err != nil {
		log.Panic(err)
	}

	// Database creation.
	datastore.CreateDatabase()
	datastore.PopulateDatabase()

	// Environment creation.
	env := handlers.Env{DB: datastore, GoBkmProxyURL: *goBkmProxyURL}

	// Building a rice box with the static directory.
	templateBox, err := rice.FindBox("static")
	if err != nil {
		log.Fatal(err)
	}

	// Getting the main template file content as a string.
	env.TplMainData, err = templateBox.String("main.html")
	if err != nil {
		log.Fatal(err)
	}

	// Getting the bookmarks with no favicon.
	noIconBookmarks := env.DB.GetNoIconBookmarks()
	if err := env.DB.FlushErrors(); err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"len(noIconBookmarks)": len(noIconBookmarks),
	}).Debug("main")

	// Updating them.
	//for _, bkm := range noIconBookmarks {
	//go env.UpdateBookmarkFavicon(bkm)
	//env.UpdateBookmarkFavicon(bkm)
	//}

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

	http.ListenAndServe(":"+*listenPort, nil)

}
