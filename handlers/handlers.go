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

	"github.com/gorilla/websocket"
	"github.com/tbellembois/gobkm/models"
	"github.com/tbellembois/gobkm/types"

	log "github.com/sirupsen/logrus"
)

const faviconRequestBaseURL = "http://www.google.com/s2/favicons?domain_url="

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	wsconn *websocket.Conn
	wserr  error
)

// Env is a structure used to pass objects throughout the application.
type Env struct {
	DB                  models.Datastore
	GoBkmProxyURL       string // the application URL
	GoBkmProxyHost      string // the application Host
	TplMainData         string // main template data
	TplAddBookmarkData  string // add bookmark template data
	TplTestData         string // test template data
	CSSMainData         []byte // main css data
	CSSAwesoneFontsData []byte // awesome fonts css data
	JsData              []byte // js data
}

// staticDataStruct is used to pass static data to the Main template.
type staticDataStruct struct {
	Bkms                []*types.Bookmark
	CSSMainData         string
	CSSAwesoneFontsData string
	JsData              string
	GoBkmProxyURL       string
	GoBkmProxyHost      string
	NewBookmarkURL      string
	NewBookmarkTitle    string
}

// exportBookmarksStruct is used to build the bookmarks and folders tree in the export operation.
type exportBookmarksStruct struct {
	Fld  *types.Folder
	Bkms []*types.Bookmark
	Sub  []*exportBookmarksStruct
}

// failHTTP send an HTTP error (httpStatus) with the given errorMessage.
func failHTTP(w http.ResponseWriter, functionName string, errorMessage string, httpStatus int) {
	log.WithFields(log.Fields{
		"functionName": functionName,
		"errorMessage": errorMessage,
	}).Error("failHTTP")
	w.WriteHeader(httpStatus)
	// JS console log
	fmt.Fprint(w, errorMessage)
}

// insertIndent the "depth" number of tabs to the given io.Writer.
func insertIndent(wr io.Writer, depth int) {
	for i := 0; i < depth; i++ {
		if _, err := wr.Write([]byte("\t")); err != nil {
			// Just logging the error.
			log.WithFields(log.Fields{
				"err": err,
			}).Error("insertIdent")
		}
	}
}

// UpdateBookmarkFavicon retrieves and updates the favicon for the given bookmark.
func (env *Env) UpdateBookmarkFavicon(bkm *types.Bookmark) {
	if u, err := url.Parse(bkm.URL); err == nil {
		// Building the favicon request URL.
		bkmDomain := u.Scheme + "://" + u.Host
		faviconRequestURL := faviconRequestBaseURL + bkmDomain
		log.WithFields(log.Fields{
			"bkmDomain":         bkmDomain,
			"faviconRequestUrl": faviconRequestURL,
		}).Debug("UpdateBookmarkFavicon")

		// Getting the favicon.
		if response, err := http.Get(faviconRequestURL); err == nil {
			defer func() {
				if err := response.Body.Close(); err != nil {
					log.WithFields(log.Fields{
						"err": err,
					}).Error("UpdateBookmarkFavicon:error closing response Body")
				}
			}()

			// Getting the favicon image type.
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

// SearchBookmarkHandler returns the bookmarks matching the search.
func (env *Env) SearchBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)
	// GET parameters retrieval.
	search := r.URL.Query()["search"]
	log.WithFields(log.Fields{
		"search": search,
	}).Debug("SearchBookmarkHandler:Query parameter")

	// Parameters check.
	if len(search) == 0 {
		failHTTP(w, "SearchBookmarkHandler", "search empty", http.StatusBadRequest)
		return
	}

	// Searching the bookmarks.
	bkms := env.DB.SearchBookmarks(search[0])

	// Adding them into a map.
	var bookmarksMap []*types.Bookmark
	for _, bkm := range bkms {
		bookmarksMap = append(bookmarksMap, bkm)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(bookmarksMap); err != nil {
		failHTTP(w, "SearchBookmarkHandler", err.Error(), http.StatusInternalServerError)
	}
}

// AddBookmarkHandler handles the bookmarks creation with drag and drop.
func (env *Env) AddBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		b   types.Bookmark
	)

	if err := r.ParseForm(); err != nil {
		failHTTP(w, "AddBookmarkHandler", "form parsing error", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&b); err != nil {
		failHTTP(w, "AddBookmarkHandler", "form decoding error", http.StatusBadRequest)
	}
	log.WithFields(log.Fields{
		"b": b,
	}).Debug("AddBookmarkHandler:Query parameter")

	// Getting the destination folder.
	dstFld := env.DB.GetFolder(b.Folder.Id)
	// Creating a new Bookmark.
	newBookmark := types.Bookmark{Title: b.Title, URL: b.URL, Folder: dstFld, Tags: b.Tags}
	// Saving the bookmark into the DB, getting its id.
	bookmarkID := env.DB.SaveBookmark(&newBookmark)
	// Datastore error check
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "AddBookmarkHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	// Updating the bookmark favicon.
	newBookmark.Id = int(bookmarkID)
	go env.UpdateBookmarkFavicon(&newBookmark)

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(newBookmark); err != nil {
		failHTTP(w, "AddBookmarkHandler", err.Error(), http.StatusInternalServerError)
	}
}

// AddFolderHandler handles the folders creation.
func (env *Env) AddFolderHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		f   types.Folder
	)

	if err := r.ParseForm(); err != nil {
		failHTTP(w, "AddFolderHandler", "form parsing error", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&f); err != nil {
		failHTTP(w, "AddFolderHandler", "form decoding error", http.StatusBadRequest)
	}
	log.WithFields(log.Fields{
		"f": f,
	}).Debug("AddFolderHandler:Query parameter")

	// Leaving on empty folder name.
	if f.Title == "" {
		return
	}

	// Getting the root folder.
	parentFolder := env.DB.GetFolder(f.Parent.Id)
	// Creating a new Folder.
	newFolder := types.Folder{Title: f.Title, Parent: parentFolder}
	// Saving the folder into the DB, getting its id.
	newFolder.Id = int(env.DB.SaveFolder(&newFolder))
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "AddFolderHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	//if err = json.NewEncoder(w).Encode(types.Folder{Id: int(folderID), Title: folderName[0], Parent: parentFolder}); err != nil {
	if err = json.NewEncoder(w).Encode(newFolder); err != nil {
		failHTTP(w, "AddFolderHandler", err.Error(), http.StatusInternalServerError)
	}
}

// DeleteFolderHandler handles the folders deletion.
func (env *Env) DeleteFolderHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		folderID int
	)
	// GET parameters retrieval.
	folderIDParam := r.URL.Query()["itemId"]
	log.WithFields(log.Fields{
		"folderIdParam": folderIDParam,
	}).Debug("DeleteFolderHandler:Query parameter")

	// Parameters check.
	if len(folderIDParam) == 0 {
		failHTTP(w, "DeleteFolderHandler", "folderIdParam empty", http.StatusBadRequest)
		return
	}
	// folderId int convertion.
	if folderID, err = strconv.Atoi(folderIDParam[0]); err != nil {
		failHTTP(w, "DeleteFolderHandler", "folderId Atoi conversion", http.StatusInternalServerError)
		return
	}

	// Getting the folder.
	fld := env.DB.GetFolder(folderID)
	// Deleting it.
	env.DB.DeleteFolder(fld)
	// Datastore error check
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "DeleteFolderHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	// Returning an empty JSON to trigger de done() ajax function.
	if err = json.NewEncoder(w).Encode(types.Folder{}); err != nil {
		failHTTP(w, "DeleteFolderHandler", err.Error(), http.StatusInternalServerError)
	}
}

// DeleteBookmarkHandler handles the bookmarks deletion.
func (env *Env) DeleteBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		bookmarkID int
	)
	// GET parameters retrieval.
	bookmarkIDParam := r.URL.Query()["itemId"]
	log.WithFields(log.Fields{
		"bookmarkIdParam": bookmarkIDParam,
	}).Debug("DeleteBookmarkHandler:Query parameter")

	// Parameters check.
	if len(bookmarkIDParam) == 0 {
		failHTTP(w, "DeleteBookmarkHandler", "bookmarkIdParam empty", http.StatusBadRequest)
		return
	}
	// bookmarkId int convertion.
	if bookmarkID, err = strconv.Atoi(bookmarkIDParam[0]); err != nil {
		failHTTP(w, "DeleteBookmarkHandler", "bookmarkId Atoi conversion", http.StatusInternalServerError)
		return
	}
	// the id in the view in negative, reverting
	bookmarkID = -bookmarkID

	// Getting the bookmark.
	bkm := env.DB.GetBookmark(bookmarkID)
	// Deleting it.
	env.DB.DeleteBookmark(bkm)
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "DeleteBookmarkHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	// Returning an empty JSON to trigger de done() ajax function.
	if err = json.NewEncoder(w).Encode(types.Bookmark{}); err != nil {
		failHTTP(w, "DeleteBookmarkHandler", err.Error(), http.StatusInternalServerError)
	}
}

// RenameFolderHandler handles the folder rename.
func (env *Env) RenameFolderHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		f   types.Folder
	)

	if err := r.ParseForm(); err != nil {
		failHTTP(w, "AddFolderHandler", "form parsing error", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&f); err != nil {
		failHTTP(w, "AddFolderHandler", "form decoding error", http.StatusBadRequest)
	}
	log.WithFields(log.Fields{
		"f": f,
	}).Debug("AddFolderHandler:Query parameter")

	// Leaving on empty folder name.
	if f.Title == "" {
		return
	}

	// Getting the folder.
	fld := env.DB.GetFolder(f.Id)
	// Renaming it.
	fld.Title = f.Title
	// Updating the folder into the DB.
	env.DB.UpdateFolder(fld)
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "RenameFolderHandler", err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(types.Folder{Id: int(fld.Id), Title: fld.Title}); err != nil {
		failHTTP(w, "RenameFolderHandler", err.Error(), http.StatusInternalServerError)
	}
}

// RenameBookmarkHandler handles the bookmarks rename.
func (env *Env) RenameBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err          error
		b            types.Bookmark
		bookmarkID   int
		bookmarkTags []*types.Tag
	)

	if err := r.ParseForm(); err != nil {
		failHTTP(w, "AddBookmarkHandler", "form parsing error", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&b); err != nil {
		failHTTP(w, "AddBookmarkHandler", "form decoding error", http.StatusBadRequest)
	}
	log.WithFields(log.Fields{
		"b": b,
	}).Debug("AddBookmarkHandler:Query parameter")

	// the id in the view in negative, reverting
	bookmarkID = -b.Id

	// Getting the bookmark.
	bkm := env.DB.GetBookmark(bookmarkID)
	// Getting the tags.
	for _, t := range b.Tags {
		if t.Id == -1 {
			// the tag is a new one with name t
			// adding it into the db
			t.Id = int(env.DB.SaveTag(&types.Tag{Name: string(t.Name)}))
		}
		bookmarkTags = append(bookmarkTags, env.DB.GetTag(t.Id))
	}
	log.WithFields(log.Fields{
		"bookmarkTags": bookmarkTags,
	}).Debug("RenameBookmarkHandler")

	// Renaming it.
	bkm.Title = b.Title
	bkm.URL = b.URL
	bkm.Tags = bookmarkTags
	// Updating the folder into the DB.
	env.DB.UpdateBookmark(bkm)
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "RenameBookmarkHandler", err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(types.Bookmark{Id: bookmarkID, Title: b.Title, URL: b.URL, Tags: b.Tags}); err != nil {
		failHTTP(w, "RenameBookmarkHandler", err.Error(), http.StatusInternalServerError)
	}
}

// StarBookmarkHandler handles the bookmark starring/unstarring.
func (env *Env) StarBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	var (
		bookmarkID int
		err        error
		star       = true
	)
	// GET parameters retrieval.
	bookmarkIDParam := r.URL.Query()["bookmarkId"]
	starParam := r.URL.Query()["star"]
	log.WithFields(log.Fields{
		"bookmarkId": bookmarkID,
		"starParam":  starParam,
	}).Debug("StarBookmarkHandler:Query parameter")

	// Parameters check.
	if len(bookmarkIDParam) == 0 {
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
	if bookmarkID, err = strconv.Atoi(bookmarkIDParam[0]); err != nil {
		failHTTP(w, "StarBookmarkHandler", "bookmarkId Atoi conversion", http.StatusInternalServerError)
		return
	}
	// the id in the view in negative, reverting
	bookmarkID = -bookmarkID

	// Getting the bookmark.
	bkm := env.DB.GetBookmark(bookmarkID)
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
	resultBookmarkStruct := types.Bookmark{Id: bookmarkID, Title: bkm.Title, URL: bkm.URL, Favicon: bkm.Favicon, Starred: bkm.Starred}
	log.WithFields(log.Fields{
		"resultBookmarkStruct": resultBookmarkStruct,
	}).Debug("StarBookmarkHandler")

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(resultBookmarkStruct); err != nil {
		failHTTP(w, "StarBookmarkHandler", err.Error(), http.StatusInternalServerError)
	}
}

// MoveBookmarkHandler handles the bookmarks move.
func (env *Env) MoveBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		b   types.Bookmark
	)

	if err := r.ParseForm(); err != nil {
		failHTTP(w, "MoveBookmarkHandler", "form parsing error", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&b); err != nil {
		failHTTP(w, "MoveBookmarkHandler", "form decoding error", http.StatusBadRequest)
	}
	log.WithFields(log.Fields{
		"b": b,
	}).Debug("MoveBookmarkHandler:Query parameter")

	// the id in the view in negative, reverting
	bookmarkID := -b.Id

	// Getting the bookmark
	bkm := env.DB.GetBookmark(bookmarkID)
	// and the destination folder if it exists.
	if b.Folder.Id != 0 {
		dstFld := env.DB.GetFolder(b.Folder.Id)
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
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	// Returning an empty JSON to trigger de done() ajax function.
	if err = json.NewEncoder(w).Encode(types.Bookmark{}); err != nil {
		failHTTP(w, "MoveBookmarkHandler", err.Error(), http.StatusInternalServerError)
	}
}

// MoveFolderHandler handles the folders move.
func (env *Env) MoveFolderHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		f   types.Folder
	)

	if err := r.ParseForm(); err != nil {
		failHTTP(w, "MoveFolderHandler", "form parsing error", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&f); err != nil {
		failHTTP(w, "MoveFolderHandler", "form decoding error", http.StatusBadRequest)
	}
	log.WithFields(log.Fields{
		"f": f,
	}).Debug("MoveFolderHandler:Query parameter")

	// Getting the source folder.
	srcFld := env.DB.GetFolder(f.Id)
	// and the destination folder if it exists.
	if f.Parent.Id != 0 {
		dstFld := env.DB.GetFolder(f.Parent.Id)
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
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	// Returning an empty JSON to trigger de done() ajax function.
	if err = json.NewEncoder(w).Encode(types.Folder{}); err != nil {
		failHTTP(w, "MoveBookmarkHandler", err.Error(), http.StatusInternalServerError)
	}
}

// getChildren recursively get subfolders and bookmarks of the folder f
func (env *Env) getChildren(f *types.Folder) types.Folder {
	log.WithFields(log.Fields{"f.Id": f.Id}).Debug("getChildren")

	f.Folders = env.DB.GetFolderSubfolders(f.Id)
	if f.Folders != nil && len(f.Folders) > 0 {
		for _, fld := range f.Folders {
			log.WithFields(log.Fields{"fld": fld}).Debug("getChildren")
			env.getChildren(fld)
		}
	}

	f.Bookmarks = env.DB.GetFolderBookmarks(f.Id)

	return *f
}

// GetTreeHandler return the entire folder and bookmark tree
func (env *Env) GetTreeHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)

	// Adding root folder.
	rootNode := &types.Folder{Id: 0, Title: "/"}

	// Getting the root folder children folders and bookmarks.
	rootNode.Folders = env.DB.GetFolderSubfolders(1)
	rootNode.Bookmarks = env.DB.GetFolderBookmarks(1)

	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "GetBranchNodesHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	// Recursively getting the subfolders and bookmarks.
	for _, fld := range rootNode.Folders {
		env.getChildren(fld)
	}

	// Adding them into a map.
	// for _, bkm := range bkms {
	// 	// Returning a default favicon if needed
	// 	if bkm.Favicon == "" {
	// 		bkm.Favicon = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAACXBIWXMAAAsSAAALEgHS3X78AAACiElEQVQ4EaVTzU8TURCf2tJuS7tQtlRb6UKBIkQwkRRSEzkQgyEc6lkOKgcOph78Y+CgjXjDs2i44FXY9AMTlQRUELZapVlouy3d7kKtb0Zr0MSLTvL2zb75eL838xtTvV6H/xELBptMJojeXLCXyobnyog4YhzXYvmCFi6qVSfaeRdXdrfaU1areV5KykmX06rcvzumjY/1ggkR3Jh+bNf1mr8v1D5bLuvR3qDgFbvbBJYIrE1mCIoCrKxsHuzK+Rzvsi29+6DEbTZz9unijEYI8ObBgXOzlcrx9OAlXyDYKUCzwwrDQx1wVDGg089Dt+gR3mxmhcUnaWeoxwMbm/vzDFzmDEKMMNhquRqduT1KwXiGt0vre6iSeAUHNDE0d26NBtAXY9BACQyjFusKuL2Ry+IPb/Y9ZglwuVscdHaknUChqLF/O4jn3V5dP4mhgRJgwSYm+gV0Oi3XrvYB30yvhGa7BS70eGFHPoTJyQHhMK+F0ZesRVVznvXw5Ixv7/C10moEo6OZXbWvlFAF9FVZDOqEABUMRIkMd8GnLwVWg9/RkJF9sA4oDfYQAuzzjqzwvnaRUFxn/X2ZlmGLXAE7AL52B4xHgqAUqrC1nSNuoJkQtLkdqReszz/9aRvq90NOKdOS1nch8TpL555WDp49f3uAMXhACRjD5j4ykuCtf5PP7Fm1b0DIsl/VHGezzP1KwOiZQobFF9YyjSRYQETRENSlVzI8iK9mWlzckpSSCQHVALmN9Az1euDho9Xo8vKGd2rqooA8yBcrwHgCqYR0kMkWci08t/R+W4ljDCanWTg9TJGwGNaNk3vYZ7VUdeKsYJGFNkfSzjXNrSX20s4/h6kB81/271ghG17l+rPTAAAAAElFTkSuQmCC"
	// 	}
	// 	// Escaping HTML characters
	// 	bkm.Title = html.EscapeString(bkm.Title)

	// 	// negating the node id to have unique ids in the view between folders and bookmarks
	// 	newNode := types.Node{Key: -bkm.Id, Title: bkm.Title, Starred: bkm.Starred, Folder: false, Lazy: false, Icon: bkm.Favicon, URL: bkm.URL, Children: nil}
	// 	rootNode.Children = append(rootNode.Children, &newNode)
	// }

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(rootNode); err != nil {
		failHTTP(w, "GetBranchNodesHandler", err.Error(), http.StatusInternalServerError)
	}
}

// GetTagsHandler retrieves the tags.
func (env *Env) GetTagsHandler(w http.ResponseWriter, r *http.Request) {

	var (
		err error
	)

	// Getting the tags.
	tags := env.DB.GetTags()
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "GetTagsHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(tags); err != nil {
		failHTTP(w, "GetTagsHandler", err.Error(), http.StatusInternalServerError)
	}
}

// GetBookmarkTagsHandler retrieves the tags for the given bookmark.
func (env *Env) GetBookmarkTagsHandler(w http.ResponseWriter, r *http.Request) {

	var (
		err        error
		bookmarkID int
	)
	// GET parameters retrieval.
	bookmarkIDParam := r.URL.Query()["bookmarkId"]
	log.WithFields(log.Fields{
		"bookmarkIDParam": bookmarkIDParam,
	}).Debug("GetBookmarkTagsHandler:Query parameter")

	// Parameters check.
	if len(bookmarkIDParam) == 0 {
		failHTTP(w, "GetBookmarkTagsHandler", "folderIdParam empty", http.StatusBadRequest)
		return
	}
	// folderId int convertion.
	if bookmarkID, err = strconv.Atoi(bookmarkIDParam[0]); err != nil {
		failHTTP(w, "GetBookmarkTagsHandler", "folderId Atoi conversion", http.StatusInternalServerError)
		return
	}
	// the id in the view in negative, reverting
	bookmarkID = -bookmarkID

	// Getting the tags.
	tags := env.DB.GetBookmarkTags(bookmarkID)
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "GetBookmarkTagsHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(tags); err != nil {
		failHTTP(w, "GetBookmarkTagsHandler", err.Error(), http.StatusInternalServerError)
	}
}

// GetBranchNodesHandler retrieves the subfolders and bookmarks of the given folder.
func (env *Env) GetBranchNodesHandler(w http.ResponseWriter, r *http.Request) {
	var (
		key int
		err error
		f   types.Folder
	)

	// GET parameters retrieval.
	parentIdParam := r.URL.Query().Get("parentId")
	log.WithFields(log.Fields{
		"keyParam": parentIdParam,
	}).Debug("GetBranchNodesHandler:Query parameter")

	// Returning the root folder if not parameters are passed.
	if len(parentIdParam) == 0 {
		parentIdParam = "1"
	}
	// key int convertion.
	if key, err = strconv.Atoi(parentIdParam); err != nil {
		failHTTP(w, "GetBranchNodesHandler", "key Atoi conversion", http.StatusInternalServerError)
		return
	}

	// Getting the folder children folders.
	f.Folders = env.DB.GetFolderSubfolders(key)
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "GetBranchNodesHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	// Getting the folder bookmarks.
	f.Bookmarks = env.DB.GetFolderBookmarks(key)
	// Datastore error check.
	if err = env.DB.FlushErrors(); err != nil {
		failHTTP(w, "GetBranchNodesHandler", err.Error(), http.StatusInternalServerError)
		return
	}

	// Adding them into a map.
	// for _, bkm := range bkms {
	// 	// Returning a default favicon if needed
	// 	if bkm.Favicon == "" {
	// 		bkm.Favicon = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAACXBIWXMAAAsSAAALEgHS3X78AAACiElEQVQ4EaVTzU8TURCf2tJuS7tQtlRb6UKBIkQwkRRSEzkQgyEc6lkOKgcOph78Y+CgjXjDs2i44FXY9AMTlQRUELZapVlouy3d7kKtb0Zr0MSLTvL2zb75eL838xtTvV6H/xELBptMJojeXLCXyobnyog4YhzXYvmCFi6qVSfaeRdXdrfaU1areV5KykmX06rcvzumjY/1ggkR3Jh+bNf1mr8v1D5bLuvR3qDgFbvbBJYIrE1mCIoCrKxsHuzK+Rzvsi29+6DEbTZz9unijEYI8ObBgXOzlcrx9OAlXyDYKUCzwwrDQx1wVDGg089Dt+gR3mxmhcUnaWeoxwMbm/vzDFzmDEKMMNhquRqduT1KwXiGt0vre6iSeAUHNDE0d26NBtAXY9BACQyjFusKuL2Ry+IPb/Y9ZglwuVscdHaknUChqLF/O4jn3V5dP4mhgRJgwSYm+gV0Oi3XrvYB30yvhGa7BS70eGFHPoTJyQHhMK+F0ZesRVVznvXw5Ixv7/C10moEo6OZXbWvlFAF9FVZDOqEABUMRIkMd8GnLwVWg9/RkJF9sA4oDfYQAuzzjqzwvnaRUFxn/X2ZlmGLXAE7AL52B4xHgqAUqrC1nSNuoJkQtLkdqReszz/9aRvq90NOKdOS1nch8TpL555WDp49f3uAMXhACRjD5j4ykuCtf5PP7Fm1b0DIsl/VHGezzP1KwOiZQobFF9YyjSRYQETRENSlVzI8iK9mWlzckpSSCQHVALmN9Az1euDho9Xo8vKGd2rqooA8yBcrwHgCqYR0kMkWci08t/R+W4ljDCanWTg9TJGwGNaNk3vYZ7VUdeKsYJGFNkfSzjXNrSX20s4/h6kB81/271ghG17l+rPTAAAAAElFTkSuQmCC"
	// 	}
	// 	// Escaping HTML characters
	// 	bkm.Title = html.EscapeString(bkm.Title)

	// 	// negating the node id to have unique ids in the view between folders and bookmarks
	// 	newNode := types.Node{Key: -bkm.Id, Title: bkm.Title, Starred: bkm.Starred, Folder: false, Lazy: false, Icon: bkm.Favicon, URL: bkm.URL, Children: nil}
	// 	nodesMap = append(nodesMap, &newNode)
	// }

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(f); err != nil {
		failHTTP(w, "GetBranchNodesHandler", err.Error(), http.StatusInternalServerError)
	}

}

// MainHandler handles the main application page.
func (env *Env) MainHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("MainHandler called")
	var (
		folderAndBookmark = new(staticDataStruct)
		err               error
	)

	// Getting the starred bookmarks.
	starredBookmarks := env.DB.GetStarredBookmarks()

	// Getting the static data.
	folderAndBookmark.JsData = string(env.JsData)
	folderAndBookmark.GoBkmProxyURL = env.GoBkmProxyURL
	folderAndBookmark.GoBkmProxyHost = env.GoBkmProxyHost
	folderAndBookmark.Bkms = starredBookmarks

	// Building the HTML template.
	htmlTpl := template.New("main")
	if htmlTpl, err = htmlTpl.Parse(env.TplMainData); err != nil {
		failHTTP(w, "MainHandler", err.Error(), http.StatusInternalServerError)
		// TODO: should we exit the program ?
	}

	if err = htmlTpl.Execute(w, folderAndBookmark); err != nil {
		failHTTP(w, "MainHandler", err.Error(), http.StatusInternalServerError)
	}
}

// ImportHandler handles the import requests.
func (env *Env) ImportHandler(w http.ResponseWriter, r *http.Request) {
	// Getting the import file.
	file, err := ioutil.ReadAll(r.Body)
	if err != nil {
		failHTTP(w, "ImportHandler", err.Error(), http.StatusInternalServerError)
		return
	}
	// Parsing the HTML.
	doc, err := html.Parse(bytes.NewReader(file))
	if err != nil {
		failHTTP(w, "ImportHandler", err.Error(), http.StatusBadRequest)
		return
	}

	// Building a new import folder name.
	currentDate := time.Now().Local()
	importFolderName := "import-" + currentDate.Format("2006-01-02")
	// Creating and saving a new folder.
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
	if _, err = w.Write([]byte("ok")); err != nil {
		// Just logging the error.
		log.WithFields(log.Fields{
			"err": err,
		}).Error("ImportHandler")
	}
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
	if _, err := w.Write([]byte(header)); err != nil {
		// Just logging the error.
		log.WithFields(log.Fields{
			"err": err,
		}).Error("ExportHandler")
	}
	// Exporting the bookmarks.
	env.ExportTree(w, &exportBookmarksStruct{Fld: rootFolder}, 0)
	// Writing the HTML footer.
	if _, err := w.Write([]byte(footer)); err != nil {
		// Just logging the error.
		log.WithFields(log.Fields{
			"err": err,
		}).Error("ExportHandler")
	}
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
	for _, child := range env.DB.GetFolderSubfolders(eb.Fld.Id) {
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
