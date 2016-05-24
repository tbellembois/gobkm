package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"text/template"
	"time"

	"golang.org/x/net/html"

	"github.com/tbellembois/gobkm/models"
	"github.com/tbellembois/gobkm/types"

	log "github.com/Sirupsen/logrus"
)

// Env is a structure used to pass objects throughout the application
type Env struct {
	DB                  models.Datastore
	GoBkmProxyURL       string // the application URL
	TplMainData         string // main template data
	CssMainData         []byte // main css data
	CssAwesoneFontsData []byte // awesome fonts css data
	JsData              []byte // js data
}

// staticDataStruct is used  to pass data to the Main template
type staticDataStruct struct {
	Bkms                []*types.Bookmark
	CssMainData         string
	CssAwesoneFontsData string
	JsData              string
	GoBkmProxyURL       string
}

// newFolderStruct is returned by the NewFolderHandler to pass the new folder id to the view
type newFolderStruct struct {
	FolderId    int64
	FolderTitle string
}

// newBookmarkStruct is returned by the NewBookmarkHandler to pass the new bookmark id to the view
type newBookmarkStruct struct {
	BookmarkId      int64
	BookmarkTitle   string
	BookmarkURL     string
	BookmarkFavicon string
	BookmarkStarred bool
}

// exportBookmarksStruct is used to build the bookmarks and folders tree in the export operation
type exportBookmarksStruct struct {
	Fld  *types.Folder
	Bkms []*types.Bookmark
	Sub  []*exportBookmarksStruct
}

// failHTTP send an HTTP error (httpStatus) with the given errorMessage
func failHTTP(w http.ResponseWriter, functionName string, errorMessage string, httpStatus int) {

	log.Error("%s: %s", functionName, errorMessage)
	w.WriteHeader(httpStatus)
	fmt.Fprint(w, errorMessage)

}

// insertIndent the "depth" number of tabs to the given io.Writer
func insertIndent(wr io.Writer, depth int) {

	for i := 0; i < depth; i++ {
		wr.Write([]byte("\t"))
	}

}

// UpdateBookmarkFavicon retrieves and updates the favicon for the given bookmark.
func (env *Env) UpdateBookmarkFavicon(bkm *types.Bookmark) {

	if u, err := url.Parse(bkm.URL); err == nil {

		bkmDomain := u.Scheme + "://" + u.Host
		faviconRequestUrl := "http://www.google.com/s2/favicons?domain_url=" + bkmDomain

		log.WithFields(log.Fields{
			"bkmDomain":         bkmDomain,
			"faviconRequestUrl": faviconRequestUrl,
		}).Debug("UpdateBookmarkFavicon")

		// Requesting Google for favicon.
		if response, err := http.Get(faviconRequestUrl); err == nil {

			defer response.Body.Close()
			contentType := response.Header.Get("Content-Type")

			log.WithFields(log.Fields{
				"response.ContentLength": response.ContentLength,
				"contentType":            contentType,
			}).Debug("UpdateBookmarkFavicon")

			// Converting the image into a base64 string.
			image, _ := ioutil.ReadAll(response.Body)
			bkm.Favicon = "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(image)

			log.WithFields(log.Fields{
				"bkm": bkm,
			}).Debug("UpdateBookmarkFavicon")

			// Updating the bookmark into the DB.
			env.DB.UpdateBookmark(bkm)
			if err = env.DB.FlushErrors(); err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("UpdateBookmarkFavicon")
			}
		}

	}
}

// AddBookmarkHandler handles the bookmarks creation.
func (env *Env) AddBookmarkHandler(w http.ResponseWriter, r *http.Request) {

	var destinationFolderId int
	var err error
	var js []byte                 // the returned JSON
	var bookmarkUrlDecoded string // the URL encoded string

	// GET parameters retrieval.
	bookmarkUrl := r.URL.Query()["bookmarkUrl"]
	destinationFolderIdParam := r.URL.Query()["destinationFolderId"]

	log.WithFields(log.Fields{
		"bookmarkUrl":              bookmarkUrl,
		"destinationFolderIdParam": destinationFolderIdParam,
	}).Debug("AddBookmarkHandler:Query parameter")

	// Parameters check.
	if len(bookmarkUrl) == 0 || len(destinationFolderIdParam) == 0 {
		failHTTP(w, "AddBookmarkHandler", "bookmarkUrl empty", http.StatusBadRequest)
		return
	}

	// Decoding the URL
	if bookmarkUrlDecoded, err = url.QueryUnescape(bookmarkUrl[0]); err != nil {
		failHTTP(w, "AddBookmarkHandler", "URL decode error", http.StatusInternalServerError)
		return
	}

	// destinationFolderId int convertion.
	if destinationFolderId, err = strconv.Atoi(destinationFolderIdParam[0]); err != nil {
		failHTTP(w, "AddBookmarkHandler", "destinationFolderId Atoi conversion", http.StatusInternalServerError)
		return
	}

	// Getting the destination folder.
	dstFld := env.DB.GetFolder(destinationFolderId)
	// Creating a new Bookmark.
	newBookmark := types.Bookmark{Title: bookmarkUrlDecoded, URL: bookmarkUrlDecoded, Folder: dstFld}
	// Saving the bookmark into the DB, getting its id.
	bookmarkId := env.DB.SaveBookmark(&newBookmark)
	// Datastore error check
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "AddBookmarkHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	// Building the JSON result.
	if js, err = json.Marshal(newBookmarkStruct{BookmarkId: bookmarkId, BookmarkURL: bookmarkUrlDecoded}); err != nil {
		failHTTP(w, "AddBookmarkHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	// Updating the bookmark favicon.
	newBookmark.Id = int(bookmarkId)
	go env.UpdateBookmarkFavicon(&newBookmark)

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

// AddFolderHandler handles the folders creation.
func (env *Env) AddFolderHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	var js []byte // the returned JSON

	// GET parameters retrieval.
	folderName := r.URL.Query()["folderName"]
	if folderName == nil {
		return
	}

	log.WithFields(log.Fields{
		"folderName": folderName,
	}).Debug("AddFolderHandler:Query parameter")

	// Parameters check.
	if len(folderName[0]) == 0 {
		failHTTP(w, "AddFolderHandler", "bookmarkUrl empty", http.StatusBadRequest)
		return
	}

	// Getting the root folder.
	rootFolder := env.DB.GetFolder(1)
	// Creating a new Folder.
	newFolder := types.Folder{Title: folderName[0], Parent: rootFolder}
	// Saving the folder into the DB, getting its id.
	folderId := env.DB.SaveFolder(&newFolder)
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "AddFolderHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	// Building the JSON result.
	if js, err = json.Marshal(newFolderStruct{FolderId: folderId, FolderTitle: folderName[0]}); err != nil {
		failHTTP(w, "AddFolderHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

// DeleteFolderHandler handles the folders deletion.
func (env *Env) DeleteFolderHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	var folderId int

	// GET parameters retrieval.
	folderIdParam := r.URL.Query()["folderId"]

	log.WithFields(log.Fields{
		"folderIdParam": folderIdParam,
	}).Debug("DeleteFolderHandler:Query parameter")

	// Parameters check.
	if len(folderIdParam) == 0 {
		failHTTP(w, "DeleteFolderHandler", "folderIdParam empty", http.StatusBadRequest)
		return
	}

	// folderId int convertion.
	if folderId, err = strconv.Atoi(folderIdParam[0]); err != nil {
		failHTTP(w, "DeleteFolderHandler", "folderId Atoi conversion", http.StatusInternalServerError)
		return
	}

	// Getting the folder.
	fld := env.DB.GetFolder(folderId)
	// Deleting it.
	env.DB.DeleteFolder(fld)
	// Datastore error check
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "DeleteFolderHandler", err.Error(), http.StatusInternalServerError)
		return
	}

}

// DeleteBookmarkHandler handles the bookmarks deletion.
func (env *Env) DeleteBookmarkHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	var bookmarkId int

	// GET parameters retrieval.
	bookmarkIdParam := r.URL.Query()["bookmarkId"]

	log.WithFields(log.Fields{
		"bookmarkIdParam": bookmarkIdParam,
	}).Debug("DeleteBookmarkHandler:Query parameter")

	// Parameters check.
	if len(bookmarkIdParam) == 0 {
		failHTTP(w, "DeleteBookmarkHandler", "bookmarkIdParam empty", http.StatusBadRequest)
		return
	}

	// bookmarkId int convertion.
	if bookmarkId, err = strconv.Atoi(bookmarkIdParam[0]); err != nil {
		failHTTP(w, "DeleteBookmarkHandler", "bookmarkId Atoi conversion", http.StatusInternalServerError)
		return
	}

	// Getting the bookmark.
	bkm := env.DB.GetBookmark(bookmarkId)
	// Deleting it.
	env.DB.DeleteBookmark(bkm)
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "DeleteBookmarkHandler", err.Error(), http.StatusInternalServerError)
		return
	}

}

// RenameFolderHandler handles the folder rename.
func (env *Env) RenameFolderHandler(w http.ResponseWriter, r *http.Request) {

	var folderId int
	var err error

	// GET parameters retrieval.
	folderIdParam := r.URL.Query()["folderId"]
	folderName := r.URL.Query()["folderName"]

	log.WithFields(log.Fields{
		"folderId":   folderId,
		"folderName": folderName,
	}).Debug("RenameFolderHandler:Query parameter")

	// Parameters check.
	if len(folderIdParam) == 0 || len(folderName) == 0 {
		failHTTP(w, "RenameFolderHandler", "folderId or folderName empty", http.StatusBadRequest)
		return
	}

	// folderId int convertion.
	if folderId, err = strconv.Atoi(folderIdParam[0]); err != nil {
		failHTTP(w, "RenameFolderHandler", "folderId Atoi conversion", http.StatusInternalServerError)
		return
	}

	// Getting the folder.
	fld := env.DB.GetFolder(folderId)
	// Renaming it.
	fld.Title = folderName[0]
	// Updating the folder into the DB.
	env.DB.UpdateFolder(fld)
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "AddBookmarkHandler", err.Error(), http.StatusInternalServerError)
		return
	}

}

// RenameBookmarkHandler handles the bookmarks rename.
func (env *Env) RenameBookmarkHandler(w http.ResponseWriter, r *http.Request) {

	var bookmarkId int
	var err error

	// GET parameters retrieval.
	bookmarkIdParam := r.URL.Query()["bookmarkId"]
	bookmarkName := r.URL.Query()["bookmarkName"]

	log.WithFields(log.Fields{
		"bookmarkId":   bookmarkId,
		"bookmarkName": bookmarkName,
	}).Debug("RenameBookmarkHandler:Query parameter")

	// Parameters check.
	if len(bookmarkIdParam) == 0 || len(bookmarkName) == 0 {
		failHTTP(w, "RenameBookmarkHandler", "bookmarkId or bookmarkName empty", http.StatusBadRequest)
		return
	}

	// bookmarkId int convertion.
	if bookmarkId, err = strconv.Atoi(bookmarkIdParam[0]); err != nil {
		failHTTP(w, "RenameBookmarkHandler", "bookmarkId Atoi conversion", http.StatusInternalServerError)
		return
	}

	// Getting the bookmark.
	bkm := env.DB.GetBookmark(bookmarkId)
	// Renaming it.
	bkm.Title = bookmarkName[0]
	// Updating the folder into the DB.
	env.DB.UpdateBookmark(bkm)
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "RenameBookmarkHandler", err.Error(), http.StatusInternalServerError)
		return
	}

}

// StarBookmarkHandler handles the bookmark starring/unstarring.
func (env *Env) StarBookmarkHandler(w http.ResponseWriter, r *http.Request) {

	var bookmarkId int
	var err error
	var js []byte // the returned JSON
	star := true

	// GET parameters retrieval.
	bookmarkIdParam := r.URL.Query()["bookmarkId"]
	starParam := r.URL.Query()["star"]

	log.WithFields(log.Fields{
		"bookmarkId": bookmarkId,
		"starParam":  starParam,
	}).Debug("StarBookmarkHandler:Query parameter")

	// Parameters check.
	if len(bookmarkIdParam) == 0 {
		failHTTP(w, "StarBookmarkHandler", "bookmarkId empty", http.StatusBadRequest)
		return
	}

	// star parameter retrieval.
	if len(starParam) == 0 || starParam[0] != "true" {
		star = false
	}

	log.WithFields(log.Fields{
		"star": star,
	}).Debug("StarBookmarkHandler")

	// bookmarkId int convertion.
	if bookmarkId, err = strconv.Atoi(bookmarkIdParam[0]); err != nil {
		failHTTP(w, "StarBookmarkHandler", "bookmarkId Atoi conversion", http.StatusInternalServerError)
		return
	}

	// Getting the bookmark.
	bkm := env.DB.GetBookmark(bookmarkId)
	// Renaming it.
	bkm.Starred = star
	// Updating the folder into the DB.
	env.DB.UpdateBookmark(bkm)
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "StarBookmarkHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	// Building the result struct.
	resultBookmarkStruct := newBookmarkStruct{BookmarkId: int64(bookmarkId), BookmarkTitle: bkm.Title, BookmarkURL: bkm.URL, BookmarkFavicon: bkm.Favicon, BookmarkStarred: bkm.Starred}

	log.WithFields(log.Fields{
		"resultBookmarkStruct": resultBookmarkStruct,
	}).Debug("StarBookmarkHandler")

	// Building the JSON result.
	if js, err = json.Marshal(resultBookmarkStruct); err != nil {
		failHTTP(w, "StarBookmarkHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

// MoveBookmarkHandler handles the bookmarks move.
func (env *Env) MoveBookmarkHandler(w http.ResponseWriter, r *http.Request) {

	var bookmarkId int
	var destinationFolderId int
	var err error

	// GET parameters retrieval.
	bookmarkIdParam := r.URL.Query()["bookmarkId"]
	destinationFolderIdParam := r.URL.Query()["destinationFolderId"]

	log.WithFields(log.Fields{
		"bookmarkIdParam":          bookmarkIdParam,
		"destinationFolderIdParam": destinationFolderIdParam,
	}).Debug("MoveBookmarkHandler:Query parameter")

	// Parameters check.
	if len(bookmarkIdParam) == 0 || len(destinationFolderIdParam) == 0 {
		failHTTP(w, "MoveBookmarkHandler", "bookmarkIdParam or destinationFolderIdParam empty", http.StatusBadRequest)
		return
	}

	// bookmarkId and destinationFolderId int convertion.
	if bookmarkId, err = strconv.Atoi(bookmarkIdParam[0]); err != nil {
		failHTTP(w, "MoveBookmarkHandler", "bookmarkId Atoi conversion", http.StatusInternalServerError)
		return
	}
	if destinationFolderId, err = strconv.Atoi(destinationFolderIdParam[0]); err != nil {
		failHTTP(w, "MoveBookmarkHandler", "destinationFolderId Atoi conversion", http.StatusInternalServerError)
		return
	}

	// Getting the bookmark
	bkm := env.DB.GetBookmark(bookmarkId)
	// and the destination folder if it exists.
	if destinationFolderId != 0 {
		dstFld := env.DB.GetFolder(destinationFolderId)

		log.WithFields(log.Fields{
			"srcBkm": bkm,
			"dstFld": dstFld,
		}).Debug("MoveBookmarkHandler: retrieved Folder instances")

		// Updating the source folder parent.
		bkm.Folder = dstFld
	} else {
		bkm.Folder = nil
	}

	// Updating the folder into the DB.
	env.DB.UpdateBookmark(bkm)
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "MoveBookmarkHandler", err.Error(), http.StatusInternalServerError)
		return
	}

}

// MoveFolderHandler handles the folders move.
func (env *Env) MoveFolderHandler(w http.ResponseWriter, r *http.Request) {

	var sourceFolderId int
	var destinationFolderId int
	var err error

	// GET parameters retrieval.
	sourceFolderIdParam := r.URL.Query()["sourceFolderId"]
	destinationFolderIdParam := r.URL.Query()["destinationFolderId"]

	log.WithFields(log.Fields{
		"sourceFolderIdParam":      sourceFolderIdParam,
		"destinationFolderIdParam": destinationFolderIdParam,
	}).Debug("GetChildrenFoldersHandler:Query parameter")

	// Parameters check.
	if len(sourceFolderIdParam) == 0 || len(destinationFolderIdParam) == 0 {
		failHTTP(w, "MoveFolderHandler", "sourceFolderIdParam or destinationFolderIdParam empty", http.StatusBadRequest)
		return
	}

	// sourceFolderId and destinationFolderId convertion.
	if sourceFolderId, err = strconv.Atoi(sourceFolderIdParam[0]); err != nil {
		failHTTP(w, "MoveFolderHandler", "sourceFolderId Atoi conversion", http.StatusInternalServerError)
		return
	}
	if destinationFolderId, err = strconv.Atoi(destinationFolderIdParam[0]); err != nil {
		failHTTP(w, "MoveFolderHandler", "destinationFolderId Atoi conversion", http.StatusInternalServerError)
		return
	}

	// Getting the source folder
	srcFld := env.DB.GetFolder(sourceFolderId)
	// and the destination folder if it exists.
	if destinationFolderId != 0 {
		dstFld := env.DB.GetFolder(destinationFolderId)

		log.WithFields(log.Fields{
			"srcFld": srcFld,
			"dstFld": dstFld,
		}).Debug("MoveFolderHandler: retrieved Folder instances")

		// Updating the source folder parent.
		srcFld.Parent = dstFld
	} else {
		srcFld.Parent = nil
	}

	// Updating the source folder into the DB.
	env.DB.UpdateFolder(srcFld)
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "MoveBookmarkHandler", err.Error(), http.StatusInternalServerError)
		return
	}

}

// GetFolderBookmarksHandler retrieves the bookmarks for the given folder.
func (env *Env) GetFolderBookmarksHandler(w http.ResponseWriter, r *http.Request) {

	var folderId int
	var err error
	var js []byte // the returned JSON

	// GET parameters retrieval.
	folderIdParam := r.URL.Query()["folderId"]

	log.WithFields(log.Fields{
		"folderIdParam": folderIdParam,
	}).Debug("GetFolderBookmarksHandler:Query parameter")

	// Parameters check.
	if len(folderIdParam) == 0 {
		failHTTP(w, "GetFolderBookmarksHandler", "folderIdParam empty", http.StatusBadRequest)
		return
	}

	// folderId convertion.
	if folderId, err = strconv.Atoi(folderIdParam[0]); err != nil {
		failHTTP(w, "GetFolderBookmarksHandler", "folderId Atoi conversion", http.StatusInternalServerError)
		return
	}

	// Getting the folder bookmarks.
	bkms := env.DB.GetFolderBookmarks(folderId)
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "GetFolderBookmarksHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	// Adding them into a map.
	bookmarksMap := make([]*types.Bookmark, 0)
	for _, bkm := range bkms {
		bookmarksMap = append(bookmarksMap, bkm)
	}

	// Building the JSON result.
	if js, err = json.Marshal(bookmarksMap); err != nil {
		failHTTP(w, "GetFolderBookmarksHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

// GetChildrenFoldersHandler retrieves the subfolders for the given folder.
func (env *Env) GetChildrenFoldersHandler(w http.ResponseWriter, r *http.Request) {

	var folderId int
	var err error
	var js []byte // the returned JSON

	// GET parameters retrieval.
	folderIdParam := r.URL.Query()["folderId"]

	log.WithFields(log.Fields{
		"folderIdParam": folderIdParam,
	}).Debug("GetChildrenFoldersHandler:Query parameter")

	// Parameters check.
	if len(folderIdParam) == 0 {
		failHTTP(w, "GetChildrenFoldersHandler", "folderIdParam empty", http.StatusBadRequest)
		return
	}

	// folderId int convertion.
	if folderId, err = strconv.Atoi(folderIdParam[0]); err != nil {
		failHTTP(w, "GetChildrenFoldersHandler", "folderId Atoi conversion", http.StatusInternalServerError)
		return
	}

	// Getting the folder children folders.
	flds := env.DB.GetChildrenFolders(folderId)
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "GetChildrenFoldersHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	// Adding them into a map.
	foldersMap := make([]*types.Folder, 0)
	for _, fld := range flds {
		foldersMap = append(foldersMap, fld)
	}

	// Building the JSON result.
	if js, err = json.Marshal(foldersMap); err != nil {
		failHTTP(w, "GetChildrenFoldersHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// MainHandler handles the main application page.
func (env *Env) MainHandler(w http.ResponseWriter, r *http.Request) {

	var folderAndBookmark = new(staticDataStruct)

	// Getting the starred bookmarks.
	starredBookmarks := env.DB.GetStarredBookmarks()

	// Getting the static data.
	folderAndBookmark.JsData = string(env.JsData)
	folderAndBookmark.GoBkmProxyURL = env.GoBkmProxyURL
	folderAndBookmark.Bkms = starredBookmarks

	// Building the main template.
	htmlTpl := template.New("main")
	htmlTpl.Parse(env.TplMainData)

	htmlTpl.Execute(w, folderAndBookmark)

}

// TestHandler
func (env *Env) TestHandler(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"r": r,
	}).Debug("TestHandler")
}

// ImportHandler handles the import requests.
func (env *Env) ImportHandler(w http.ResponseWriter, r *http.Request) {

	// Getting the uploaded import file.
	//	file, _, err := r.FormFile("importFile")
	//	if err != nil {
	//		failHTTP(w, "ImportHandler", err.Error(), http.StatusInternalServerError)
	//	}

	file, err := ioutil.ReadAll(r.Body)
	if err != nil {
		failHTTP(w, "ImportHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	// Parsing the HTML.
	//doc, err := html.Parse(file)
	doc, err := html.Parse(bytes.NewReader(file))
	if err != nil {
		failHTTP(w, "ImportHandler", err.Error(), http.StatusBadRequest)
		return
	}

	// Building a new import folder name.
	currentDate := time.Now().Local()
	importFolderName := "import-" + currentDate.Format("2006-01-02")

	// Creating a new folder.
	importFolder := types.Folder{Title: importFolderName}
	id := env.DB.SaveFolder(&importFolder)

	importFolder.Id = int(id)

	// Function to recursively parse the n node.
	var f func(n *html.Node, parentFolder *types.Folder)

	f = func(n *html.Node, parentFolder *types.Folder) {

		// Keeping the parent folder before calling f recursively.
		var parentFolderBackup types.Folder
		parentFolderBackup = *parentFolder

		// Parsing the n children.
		for c := n.FirstChild; c != nil; c = c.NextSibling {

			// Got a dt tag.
			if c.Type == html.ElementNode && c.Data == "dt" {
				dtTag := c.FirstChild

				switch dtTag.Data {
				case "h3":
					// Got a <dt><h3> tag.
					// Building the new folder.
					h3Value := dtTag.FirstChild.Data
					newFolder := types.Folder{Title: h3Value, Parent: parentFolder}
					// Saving it into the DB.
					id := env.DB.SaveFolder(&newFolder)
					newFolder.Id = int(id)
					// Updating the parent folder for next recursion.
					parentFolder = &newFolder
				case "a":
					// Got a <dt><a> tag.
					var h3Value string
					var h3Href string
					var h3Icon string

					// Parsing the link attributes for href and icon.
					for _, attr := range dtTag.Attr {
						key := attr.Key
						val := attr.Val
						if key == "href" {
							h3Href = val
						}
						if key == "icon" {
							h3Icon = val
						}
					}

					// Looking for a link title.
					if dtTag.FirstChild != nil {
						h3Value = dtTag.FirstChild.Data
					} else {
						h3Value = h3Href
					}

					// Creating the new Bookmark.
					newBookmark := types.Bookmark{Title: h3Value, URL: h3Href, Favicon: h3Icon, Folder: parentFolder}
					log.WithFields(log.Fields{
						"newBookmark": newBookmark,
					}).Debug("ImportHandler:Saving bookmark")
					// And saving it.
					env.DB.SaveBookmark(&newBookmark)
				}
			}

			// Calling recursively f for each child of n.
			f(c, parentFolder)

			// Restoring the parent folder.
			parentFolder = &parentFolderBackup
		}

	}

	// Importing the folders and bookmarks.
	f(doc, &importFolder)

	// Database errors check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "ImportHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	// Returning "ok" to inform the AJAX caller that everyting was fine.
	w.Write([]byte("ok"))

}

// ExportHandler handles the export requests.
func (env *Env) ExportHandler(w http.ResponseWriter, r *http.Request) {

	// Getting the root folder.
	rootFolder := env.DB.GetFolder(1)

	// HTML header and footer definition.
	header := `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<!-- This is an automatically generated file.
     It will be read and overwritten.
     DO NOT EDIT! -->
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>GoBkm</TITLE>
<H1>GoBkm</H1>
<DL><p>` + "\n"
	footer := "</DL><p>\n"

	// Writing the header meta informations.
	w.Header().Set("Content-Disposition", "attachment; filename=gobkm.html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Writing the HTML header
	w.Write([]byte(header))

	// Exporting the bookmarks.
	env.ExportTree(w, &exportBookmarksStruct{Fld: rootFolder}, 0)

	// Writing the HTML footer.
	w.Write([]byte(footer))

}

// ExportTree recursively exports in HTML the given bookmark struct.
func (env *Env) ExportTree(wr io.Writer, eb *exportBookmarksStruct, depth int) *exportBookmarksStruct {

	// Depth is just for cosmetics indent purposes.
	depth++

	log.WithFields(log.Fields{
		"*eb": *eb,
	}).Debug("ExportTree")

	// Writing the folder title.
	insertIndent(wr, depth)
	wr.Write([]byte("<DT><H3>" + eb.Fld.Title + "</H3>\n"))
	insertIndent(wr, depth)
	wr.Write([]byte("<DL><p>\n"))

	// For each children folder recursively building the bookmars tree.
	for _, child := range env.DB.GetChildrenFolders(eb.Fld.Id) {
		eb.Sub = append(eb.Sub, env.ExportTree(wr, &exportBookmarksStruct{Fld: child}, depth))
	}

	// Getting the folder bookmarks.
	eb.Bkms = env.DB.GetFolderBookmarks(eb.Fld.Id)
	// Writing them.
	for _, bkm := range eb.Bkms {
		insertIndent(wr, depth)
		wr.Write([]byte("<DT><A HREF=\"" + bkm.URL + "\" ICON=\"" + bkm.Favicon + "\">" + bkm.Title + "</A>\n"))
	}
	insertIndent(wr, depth)
	wr.Write([]byte("</DL><p>\n"))

	return eb

}
