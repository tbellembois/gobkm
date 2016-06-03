package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/tbellembois/gobkm/types"
	"honnef.co/go/js/dom"
	"honnef.co/go/js/xhr"
)

// CSS classes.
const (
	ClassItemOver           = "folder-over"
	ClassItemFolder         = "folder"
	ClassItemFolderOpen     = "fa-folder-open-o"
	ClassItemFolderClosed   = "fa-folder-o"
	ClassItemBookmark       = "bookmark"
	ClassBookmarkStarred    = "fa fa-star"
	ClassBookmarkNotStarred = "fa fa-star-o"
)

var (
	w dom.Window
	d dom.Document
	// The dragged element ID.
	draggedItemID string
)

type folderStruct struct {
	fld     dom.HTMLDivElement
	subFlds dom.HTMLUListElement
}

type newBookmarkStruct struct {
	BookmarkId      int64
	BookmarkTitle   string
	BookmarkURL     string
	BookmarkFavicon string
	BookmarkStarred bool
}

type newFolderStruct struct {
	FolderID    int64
	FolderTitle string
}

func init() {
	w = dom.GetWindow()
	d = w.Document()

}

//
// Utils functions.
//
func getClosest(elem dom.Node, selector string) dom.Node {

	firstChar := string(selector[0])

	for ; elem.NodeName() != d.NodeName(); elem = elem.ParentNode() {

		// class selector
		if firstChar == "." {
			if elem.(dom.HTMLElement).Class().Contains(selector[1:]) {
				return elem
			}
		}
		// id selector
		if firstChar == "#" {
			if elem.(dom.HTMLElement).ID() == selector[1:] {
				return elem
			}
		}
		// attribute selector
		if firstChar == "[" {
			if elem.(dom.HTMLElement).HasAttribute(selector[1 : len(selector)-2]) {
				return elem
			}
		}
		// tag selector
		if elem.(dom.HTMLElement).TagName() == selector {
			return elem
		}
	}
	return nil
}

func isHidden(id string) bool {
	return d.GetElementByID(id).(dom.HTMLElement).Style().GetPropertyValue("display") == "none"
}
func isDisabled(id string) bool {
	return d.GetElementByID(id).(dom.HTMLElement).HasAttribute("disabled")
}
func hideItem(id string) {
	d.GetElementByID(id).(dom.HTMLElement).Style().SetProperty("display", "none", "")
}
func showItem(id string) {
	d.GetElementByID(id).(dom.HTMLElement).Style().SetProperty("display", "block", "")
}
func enableItem(id string) {
	d.GetElementByID(id).RemoveAttribute("disabled")
}
func disableItem(id string) {
	d.GetElementByID(id).SetAttribute("disabled", "true")
}
func resetItemValue(id string) {
	d.GetElementByID(id).SetAttribute("value", "")
}
func setItemValue(id string, val string) {
	d.GetElementByID(id).SetAttribute("value", val)
}
func setClass(el dom.HTMLElement, class string) {
	el.Class().SetString(class)
}
func addClass(el dom.HTMLElement, class string) {
	el.Class().Add(class)
}
func removeClass(el dom.HTMLElement, class string) {
	el.Class().Remove(class)
}
func hasClass(el dom.HTMLElement, class string) bool {
	return el.Class().Contains(class)
}
func openInParent(url string) {
	w.Parent().Open(url, "", "")
}

func isStarredBookmark(bkmID string) bool {
	return d.GetElementByID("bookmark-starred-link-"+bkmID) != nil
}
func hasChildrenFolders(fldID string) bool {
	return hasClass(d.GetElementByID("folder-"+fldID).(dom.HTMLElement), ClassItemFolderOpen)
}

func setWait() {
	print("setWait")
	d.GetElementsByTagName("body")[0].Class().SetString("wait")
}
func unsetWait() {
	print("unsetWait")
	d.GetElementsByTagName("body")[0].Class().Remove("wait")
}

func undisplayChildrenFolders(fldID string) {
	// Removing folder content.
	d.GetElementByID("subfolders-" + fldID).SetInnerHTML("")
	// Changing folder icon.
	setClass(d.GetElementByID("folder-"+fldID).(dom.HTMLElement), ClassItemFolder+" "+ClassItemFolderClosed)
}
func showRenameBox() {
	showItem("rename-input-box")
	disableItem("add-folder")
	disableItem("add-folder-button")
}
func hideRenameBox() {
	hideItem("rename-input-box")
	resetItemValue("rename-input-box-form")
	resetItemValue("rename-hidden-input-box-form")
	enableItem("add-folder")
	enableItem("add-folder-button")
}
func setRenameFormValue(val string) {
	setItemValue("rename-input-box-form", val)
	d.GetElementByID("rename-input-box-form").(*dom.HTMLInputElement).Call("select")
}
func setRenameHiddenFormValue(val string) {
	setItemValue("rename-hidden-input-box-form", val)
}
func hideImport() {
	hideItem("import-input-box")
}
func toogleDisplayImport() {
	if isHidden("import-input-box") {
		showItem("import-input-box")
	} else {
		hideItem("import-input-box")
	}
}

func displaySubfolder(pFldID string, fldID string, fldTitle string, nbChildrenFolders int) {

	if d.GetElementByID("folder-"+fldID) != nil {
		return
	}

	newFld := createFolder(fldID, fldTitle, nbChildrenFolders)

	d.GetElementByID("subfolders-" + pFldID).AppendChild(newFld.fld)
	d.GetElementByID("subfolders-" + pFldID).AppendChild(newFld.subFlds)

}

func displayBookmark(pFldID string, bkmID string, bkmTitle string, bkmURL string, bkmFavicon string, bkmStarred bool) {

	if d.GetElementByID("bookmark-"+bkmID) != nil {
		return
	}

	newBkm := createBookmark(bkmID, bkmTitle, bkmURL, bkmFavicon, bkmStarred, false)

	d.GetElementByID("subfolders-" + pFldID).AppendChild(newBkm)

}

//
// drag/over/leave folder/bookmark listeners
//
func overItem(e dom.Event) {
	e.PreventDefault()
	addClass(e.Target().(dom.HTMLElement), ClassItemOver)
}

func leaveItem(e dom.Event) {
	e.PreventDefault()
	removeClass(e.Target().(dom.HTMLElement), ClassItemOver)
}

func dragItem(e dom.Event) {
	draggedItemID = e.Target().ID()
}

func dropRename(e dom.Event) {

	//draggedItemID = e.(*dom.DragEvent).Get("dragItemId").String()
	draggedItemIDDigit := strings.Split(draggedItemID, "-")[1]
	defer func() { draggedItemID = "" }()

	if strings.HasPrefix(draggedItemID, "folder") {
		draggedFldName := d.GetElementByID(draggedItemID).TextContent()
		setRenameFormValue(draggedFldName)

	} else {
		draggedBkmName := d.GetElementByID("bookmark-link-" + draggedItemIDDigit).TextContent()
		setRenameFormValue(draggedBkmName)
	}
	showRenameBox()
	setRenameHiddenFormValue(draggedItemID)
	removeClass(d.GetElementByID("rename-box").(dom.HTMLElement), ClassItemOver)

}

//
// HTML elements creation helpers
//
func createBookmark(bkmID string, bkmTitle string, bkmURL string, bkmFavicon string, bkmStarred bool, starred bool) dom.HTMLElement {

	// Link (actually a clickable div).
	a := d.CreateElement("div").(*dom.HTMLDivElement)
	a.SetTitle(bkmTitle)
	a.AppendChild(d.CreateTextNode(bkmTitle))
	a.AddEventListener("click", false, func(e dom.Event) { openInParent(bkmURL) })
	// Main div.
	md := d.CreateElement("div").(*dom.HTMLDivElement)
	md.SetClass(ClassItemBookmark)
	// Favicon.
	fav := d.CreateElement("img").(*dom.HTMLImageElement)
	fav.Src = bkmFavicon
	fav.SetClass("favicon")
	// Star.
	str := d.CreateElement("div").(*dom.HTMLDivElement)
	str.AddEventListener("click", false, func(e dom.Event) { starBookmark(bkmID) })

	if starred {
		a.SetID("bookmark-starred-link-" + bkmID)
		md.SetID("bookmark-starred-" + bkmID)
		md.SetDraggable(false)
		str.SetID("bookmark-starred-star-" + bkmID)
		str.SetClass(ClassBookmarkStarred)
	} else {
		a.SetID("bookmark-link-" + bkmID)
		md.SetID("bookmark-" + bkmID)
		md.SetDraggable(true)
		str.SetID("bookmark-star-" + bkmID)
		if bkmStarred {
			str.SetClass(ClassBookmarkStarred)
		} else {
			str.SetClass(ClassBookmarkNotStarred)
		}
		md.AddEventListener("dragstart", false, dragItem)

	}

	md.AppendChild(str)
	md.AppendChild(fav)
	md.AppendChild(a)

	return md
}

func createFolder(fldID string, fldTitle string, nbChildrenFolders int) folderStruct {

	// Main div.
	md := d.CreateElement("div").(*dom.HTMLDivElement)
	md.SetTitle(fldTitle)
	md.SetClass(ClassItemFolder + " " + ClassItemFolderClosed)
	md.SetID("folder-" + fldID)
	md.SetDraggable(true)
	// Subfolders.
	ul := d.CreateElement("ul").(*dom.HTMLUListElement)
	ul.SetID("subfolders-" + fldID)

	md.AppendChild(d.CreateTextNode(fldTitle))
	md.AppendChild(ul)

	md.AddEventListener("click", false, func(e dom.Event) { getChildrenItems(e, fldID) })
	md.AddEventListener("dragover", false, func(e dom.Event) { overItem(e) })
	md.AddEventListener("dragleave", false, func(e dom.Event) { leaveItem(e) })
	md.AddEventListener("dragstart", false, func(e dom.Event) { dragItem(e) })
	md.AddEventListener("drop", false, func(e dom.Event) { dropFolder(e) })

	return folderStruct{fld: *md, subFlds: *ul}
}

//
// AJAX requests
//
// structure to pass arguments to the sendRequest method
type arg struct {
	key string
	val string
}

// sendRequest performs a GET request to url with args
func sendRequest(url string, args []arg) *http.Response {

	var err error
	var req *http.Request
	var resp *http.Response

	url = strings.Join([]string{url, "?"}, "")
	for i, arg := range args {
		if i == 0 {
			url = strings.Join([]string{url, fmt.Sprintf("%s=%s", arg.key, arg.val)}, "")
		} else {
			url = strings.Join([]string{url, fmt.Sprintf("%s=%s", arg.key, arg.val)}, "&")
		}
	}

	if req, err = http.NewRequest("GET", url, nil); err != nil {
		fmt.Println("request build error:", url)
		return resp
	}

	client := &http.Client{}
	if resp, err = client.Do(req); err != nil {
		fmt.Println("request error:", url)
		return resp
	}
	defer resp.Body.Close()

	return resp

}

func importBookmarks(e dom.Event) {
	e.PreventDefault()
	go func() {

		setWait()
		setItemValue("import-button", "importing...")

		fileSelect := d.GetElementByID("import-file").(*dom.HTMLInputElement)
		file := fileSelect.Files()[0]

		req := xhr.NewRequest("POST", "/import/")
		if err := req.Send(file.Object); err != nil {
			fmt.Println("importBookmarks response code error")
			return
		}

		unsetWait()
		setItemValue("import-button", "import")
		hideImport()
		d.GetElementByID("folder-1").(*dom.HTMLDivElement).Click()
	}()
}

func starBookmark(bkmID string) {
	go func() {

		var (
			err  error
			resp *http.Response
			data newBookmarkStruct // returned struct from server
		)

		star := true
		starredBookmark := isStarredBookmark(bkmID)
		starBookmarkDiv := d.GetElementByID("bookmark-star-" + bkmID).(dom.HTMLElement)
		if starredBookmark {
			star = false
		}

		if resp = sendRequest("/starBookmark/", []arg{{key: "star", val: strconv.FormatBool(star)}, {key: "bookmarkId", val: bkmID}}); resp.StatusCode != http.StatusOK {
			fmt.Println("starBookmark response code error")
			return
		}

		if starredBookmark {
			if starBookmarkDiv != nil {
				setClass(starBookmarkDiv, ClassBookmarkNotStarred)
			}
			el := d.GetElementByID("bookmark-starred-" + bkmID)
			el.ParentNode().RemoveChild(el)
		} else {
			setClass(starBookmarkDiv, ClassBookmarkStarred)

			if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
				fmt.Println("starBookmark JSON decoder error")
				return
			}
			newBkm := createBookmark(bkmID, data.BookmarkTitle, data.BookmarkURL, data.BookmarkFavicon, data.BookmarkStarred, true)

			li := d.CreateElement("li").(*dom.HTMLLIElement)
			li.AppendChild(newBkm)
			d.GetElementByID("starred").AppendChild(li)
		}
	}()
}

func addFolder(e dom.Event) {
	e.PreventDefault()
	go func() {

		var (
			err  error
			resp *http.Response
			data newFolderStruct // returned struct from server
		)

		fldName := d.GetElementByID("add-folder").(*dom.HTMLInputElement).Value

		if resp = sendRequest("/addFolder/", []arg{{key: "folderName", val: fldName}}); resp.StatusCode != http.StatusOK {
			fmt.Println("starBookmark response code error")
			return
		}

		if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
			fmt.Println("addFolder JSON decoder error")
			return
		}

		newFld := createFolder(strconv.Itoa(int(data.FolderID)), data.FolderTitle, 0)

		rootFld := d.GetElementByID("subfolders-1")
		rootFld.InsertBefore(newFld.fld, rootFld.FirstChild())
		rootFld.InsertBefore(newFld.subFlds, rootFld.FirstChild())
	}()
}

func dropDelete(e dom.Event) {
	e.PreventDefault()
	go func() {

		var (
			resp *http.Response
		)

		draggedItem := d.GetElementByID(draggedItemID)
		draggedItemIDDigit := strings.Split(draggedItemID, "-")[1]
		defer func() { draggedItemID = "" }()

		if strings.HasPrefix(draggedItemID, "folder") {
			if resp = sendRequest("/deleteFolder/", []arg{{key: "folderId", val: draggedItemIDDigit}}); resp.StatusCode != http.StatusOK {
				fmt.Println("dropDelete response code error")
				return
			}

			children := d.GetElementByID("subfolders-" + draggedItemIDDigit)
			children.ParentNode().RemoveChild(children)
			draggedItem.ParentNode().RemoveChild(draggedItem)
		} else {
			if resp = sendRequest("/deleteBookmark/", []arg{{key: "bookmarkId", val: draggedItemIDDigit}}); resp.StatusCode != http.StatusOK {
				fmt.Println("dropDelete response code error")
				return
			}

			draggedItem.ParentNode().RemoveChild(draggedItem)
		}

		removeClass(d.GetElementByID("delete-box").(dom.HTMLElement), ClassItemOver)
	}()

}

func dropFolder(e dom.Event) {

	e.PreventDefault()
	// Putting the following instruction inside the go routine does not work. I don't know why.
	url := e.(*dom.DragEvent).Get("dataTransfer").Call("getData", "URL").String()

	go func() {

		draggedItem := d.GetElementByID(draggedItemID)
		defer func() { draggedItemID = "" }()

		var (
			draggedItemIDDigit  string
			draggedItemChildren dom.Element
			err                 error
			resp                *http.Response
		)
		if draggedItem != nil {
			draggedItemIDDigit = strings.Split(draggedItemID, "-")[1]
			draggedItemChildren = d.GetElementByID("subfolders-" + draggedItemIDDigit)
		}
		droppedItem := e.Target().(dom.HTMLElement)
		droppedItemID := droppedItem.GetAttribute("id")
		droppedItemIDDigit := strings.Split(droppedItemID, "-")[1]
		droppedItemChildren := d.GetElementByID("subfolders-" + droppedItemIDDigit)

		if draggedItem != nil && strings.HasPrefix(draggedItemID, "folder") {
			// Can not move a folder into itself.
			if draggedItemIDDigit == droppedItemIDDigit {
				fmt.Println("can not move a folder into itself")
				return
			}
			// Can not move a folder into its first parent.
			draggedParentChildrenUlId := getClosest(draggedItem, "UL").(dom.HTMLElement).ID()
			if draggedParentChildrenUlId == "subfolders-"+droppedItemIDDigit {
				fmt.Println("can not move a folder into its first parent")
				return
			}

			// Can not move a folder into one of its children.
			//TODO
			if getClosest(droppedItem, "#subfolders-"+draggedItemIDDigit) != nil {
				fmt.Println("can not move a folder into one of its children")
				return
			}

			if resp = sendRequest("/moveFolder/", []arg{{key: "sourceFolderId", val: draggedItemIDDigit}, {key: "destinationFolderId", val: droppedItemIDDigit}}); resp.StatusCode != http.StatusOK {
				fmt.Println("dropFolder response code error")
				return
			}

			droppedItemChildren.InsertBefore(draggedItemChildren, droppedItemChildren.FirstChild())
			droppedItemChildren.InsertBefore(draggedItem, droppedItemChildren.FirstChild())
			removeClass(droppedItem, ClassItemOver)
			removeClass(droppedItem, ClassItemFolderClosed)
			addClass(droppedItem, ClassItemFolderOpen)
		} else if draggedItem != nil && strings.HasPrefix(draggedItemID, "bookmark") {

			if resp = sendRequest("/moveBookmark/", []arg{{key: "bookmarkId", val: draggedItemIDDigit}, {key: "destinationFolderId", val: droppedItemIDDigit}}); resp.StatusCode != http.StatusOK {
				fmt.Println("dropFolder response code error")
				return
			}

			if hasChildrenFolders(droppedItemIDDigit) {
				droppedItemChildren.InsertBefore(draggedItem, droppedItemChildren.FirstChild())
				removeClass(droppedItem, ClassItemFolderClosed)
				addClass(droppedItem, ClassItemFolderOpen)
			} else {
				draggedItem.ParentNode().RemoveChild(draggedItem)
			}
		} else {
			var dataBkm newBookmarkStruct
			if resp = sendRequest("/addBookmark/", []arg{{key: "bookmarkUrl", val: url}, {key: "destinationFolderId", val: droppedItemIDDigit}}); resp.StatusCode != http.StatusOK {
				fmt.Println("dropFolder response code error")
				return
			}
			if err = json.NewDecoder(resp.Body).Decode(&dataBkm); err != nil {
				fmt.Println("getChildrenFolders JSON decoder error")
				return
			}

			newBkm := createBookmark(strconv.Itoa(int(dataBkm.BookmarkId)), dataBkm.BookmarkURL, dataBkm.BookmarkURL, "", dataBkm.BookmarkStarred, false)

			droppedItemChildren.InsertBefore(newBkm, droppedItemChildren.FirstChild())
			removeClass(droppedItem, ClassItemOver)
			removeClass(droppedItem, ClassItemFolderClosed)
			addClass(droppedItem, ClassItemFolderOpen)
		}
	}()

}

func renameFolder(e dom.Event) {
	e.PreventDefault()
	go func() {

		var (
			resp *http.Response
		)

		defer func() { draggedItemID = "" }()

		fldID := d.GetElementByID("rename-hidden-input-box-form").(*dom.HTMLInputElement).Value
		fldName := d.GetElementByID("rename-input-box-form").(*dom.HTMLInputElement).Value
		fldIDDigit := strings.Split(fldID, "-")[1]

		if strings.HasPrefix(fldID, "folder") {

			if resp = sendRequest("/renameFolder/", []arg{{key: "folderId", val: fldIDDigit}, {key: "folderName", val: fldName}}); resp.StatusCode != http.StatusOK {
				fmt.Println("renameFolder response code error")
				return
			}

			d.GetElementByID(draggedItemID).SetInnerHTML(fldName)

		} else {

			if resp = sendRequest("/renameBookmark/", []arg{{key: "bookmarkId", val: fldIDDigit}, {key: "bookmarkName", val: fldName}}); resp.StatusCode != http.StatusOK {
				fmt.Println("renameBookmark response code error")
				return
			}

			d.GetElementByID("bookmark-link-" + fldIDDigit).SetInnerHTML(fldName)

		}

		hideRenameBox()
	}()

}

func getChildrenItems(e dom.Event, fldIDDigit string) {

	go func() {

		var (
			err     error
			resp    *http.Response
			dataFld []types.Folder
			dataBkm []types.Bookmark
		)

		// Toggle display.
		if hasChildrenFolders(fldIDDigit) {
			undisplayChildrenFolders(fldIDDigit)
			return
		}

		// Getting the folder subfolders.
		if resp = sendRequest("/getChildrenFolders/", []arg{{key: "folderId", val: fldIDDigit}}); resp.StatusCode != http.StatusOK {
			fmt.Println("getChildrenFolders response code error")
			return
		}
		if err = json.NewDecoder(resp.Body).Decode(&dataFld); err != nil {
			fmt.Println("getChildrenFolders JSON decoder error")
			return
		}
		for _, fld := range dataFld {
			displaySubfolder(fldIDDigit, strconv.Itoa(fld.Id), fld.Title, fld.NbChildrenFolders)
		}

		// Getting the folder bookmarks.
		if resp = sendRequest("/getFolderBookmarks/", []arg{{key: "folderId", val: fldIDDigit}}); resp.StatusCode != http.StatusOK {
			fmt.Println("getChildrenFolders response code error")
			return
		}
		if err = json.NewDecoder(resp.Body).Decode(&dataBkm); err != nil {
			fmt.Println("getChildrenFolders JSON decoder error")
			return
		}
		for _, bkm := range dataBkm {
			displayBookmark(fldIDDigit, strconv.Itoa(bkm.Id), bkm.Title, bkm.URL, bkm.Favicon, bkm.Starred)
		}

		// Changing the folder icon.
		setClass(d.GetElementByID("folder-"+fldIDDigit).(dom.HTMLElement), ClassItemFolder+" "+ClassItemFolderOpen)

	}()

}

func main() {

	// Add/Rename folder button listener.
	d.GetElementByID("add-folder-button").AddEventListener("click", false, addFolder)
	d.GetElementByID("rename-folder-button").AddEventListener("click", false, renameFolder)

	// Bind enter key to add or rename a folder.
	d.AddEventListener("keydown", false, func(e dom.Event) {
		ke := e.(*dom.KeyboardEvent)
		if ke.KeyCode == 13 {
			e.PreventDefault()
		}
	})

	// Root folder listeners.
	fld := d.GetElementByID("folder-1").(*dom.HTMLDivElement)
	fld.AddEventListener("click", false, func(e dom.Event) {
		getChildrenItems(e, "1")
	})
	fld.AddEventListener("dragover", false, func(e dom.Event) {
		overItem(e)
	})
	fld.AddEventListener("dragleave", false, func(e dom.Event) {
		leaveItem(e)
	})

	// Rename and delete boxes listeners.
	rb := d.GetElementByID("rename-box").(*dom.HTMLDivElement)
	rb.AddEventListener("dragover", false, func(e dom.Event) {
		overItem(e)
	})
	rb.AddEventListener("dragleave", false, func(e dom.Event) {
		leaveItem(e)
	})
	rb.AddEventListener("drop", false, func(e dom.Event) {
		dropRename(e)
	})
	db := d.GetElementByID("delete-box").(*dom.HTMLDivElement)
	db.AddEventListener("dragover", false, func(e dom.Event) {
		overItem(e)
	})
	db.AddEventListener("dragleave", false, func(e dom.Event) {
		leaveItem(e)
	})
	db.AddEventListener("drop", false, func(e dom.Event) {
		dropDelete(e)
	})

	// Import export links listeners.
	importLink := d.GetElementByID("import-link").(*dom.HTMLSpanElement)
	importLink.AddEventListener("click", false, func(e dom.Event) {
		toogleDisplayImport()
	})
	exportLink := d.GetElementByID("export-link").(*dom.HTMLSpanElement)
	exportLink.AddEventListener("click", false, func(e dom.Event) {
		openInParent("/export/")
	})

	// Import form listener.
	formImport := d.GetElementByID("import-file-form")
	formImport.AddEventListener("submit", false, func(e dom.Event) {
		importBookmarks(e)
	})

	// Enter key listener
	d.AddEventListener("keydown", false, func(e dom.Event) {
		fmt.Println(e.(*dom.KeyboardEvent).KeyCode)
		if e.(*dom.KeyboardEvent).KeyCode == 13 {
			if isDisabled("add-folder") {
				renameFolder(e)
			} else {
				addFolder(e)
			}
		}
	})

}
