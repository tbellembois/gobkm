package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"
	"github.com/tbellembois/gobkm/types"
	"honnef.co/go/js/dom"
	"honnef.co/go/js/xhr"
)

// CSS classes.
const (
	ClassDraggedItem             = "dragged-item"
	ClassItemOver                = "folder-over"
	ClassRenameOver              = "rename-over"
	ClassDeleteOver              = "delete-over"
	ClassItemFolder              = "folder"
	ClassItemFolderAwesome       = "fa"
	ClassItemFolderAwesomeOpen   = "fa-folder-open-o"
	ClassItemFolderAwesomeClosed = "fa-folder-o"
	ClassItemFolderOpen          = ClassItemFolderAwesome + " " + ClassItemFolderAwesomeOpen
	ClassItemFolderClosed        = ClassItemFolderAwesome + " " + ClassItemFolderAwesomeClosed
	ClassItemBookmark            = "bookmark"
	ClassItemBookmarkLink        = "bookmark-link"
	ClassItemBookmarkLinkEdited  = "bookmark-link-edited"
	ClassBookmarkStarred         = "fa fa-star"
	ClassBookmarkNotStarred      = "fa fa-star-o"
)

var (
	w           dom.Window
	d           dom.Document
	changeTimer int
)

type folderStruct struct {
	fld     dom.HTMLDivElement
	subFlds dom.HTMLUListElement
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
	if e := d.GetElementByID(id); e != nil {
		e.(dom.HTMLElement).Style().SetProperty("display", "none", "")
	}
}
func showItem(id string) {
	if e := d.GetElementByID(id); e != nil {
		e.(dom.HTMLElement).Style().SetProperty("display", "block", "")
	}
}
func enableItem(id string) {
	if e := d.GetElementByID(id); e != nil {
		e.RemoveAttribute("disabled")
	}
}
func disableItem(id string) {
	if e := d.GetElementByID(id); e != nil {
		e.SetAttribute("disabled", "true")
	}
}
func resetItemValue(id string) {
	if e := d.GetElementByID(id); e != nil {
		e.SetAttribute("value", "")
	}
}
func setItemValue(id string, val string) {
	if e := d.GetElementByID(id); e != nil {
		e.SetAttribute("value", val)
	}
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
	hideImport()
	hideRenameBox()
	w.Parent().Open(url, "", "")
}

func isStarredBookmark(bkmID string) bool {
	return d.GetElementByID("bookmark-starred-link-"+bkmID) != nil
}
func hasChildrenFolders(fldID string) bool {
	return hasClass(d.GetElementByID("folder-"+fldID).(dom.HTMLElement), ClassItemFolderAwesomeOpen)
}

func setWait() {
	d.GetElementsByTagName("body")[0].Class().SetString("wait")
}
func unsetWait() {
	d.GetElementsByTagName("body")[0].Class().Remove("wait")
}

func resetAddFolder() {
	d.GetElementByID("add-folder").(*dom.HTMLInputElement).Set("value", "")
}
func resetAll() {
	clearSearchResults()
	hideImport()
	hideRenameBox()
	resetAddFolder()
	enableItem("add-folder")
	enableItem("add-folder-button")

	for _, e := range d.GetElementsByClassName(ClassItemBookmarkLinkEdited) {
		removeClass(e.(dom.HTMLElement), ClassItemBookmarkLinkEdited)
		addClass(e.(dom.HTMLElement), ClassItemBookmarkLink)
	}
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
	d.GetElementByID("rename-input-box-form").(*dom.HTMLInputElement).Call("focus")
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

func clearSearchResults() {
	d.GetElementByID("search-result").SetInnerHTML("")
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
func keyDownItem(e dom.Event) {
	ke := e.(*dom.KeyboardEvent)
	if ke.KeyCode == 82 {
		e.PreventDefault()
		id := e.Target().(dom.HTMLElement).ID()
		dropRename(string(id))
	}
}

func mouseOverItem(e dom.Event) {
	e.PreventDefault()
	e.Target().(dom.HTMLElement).Focus()
}

func dragOverItem(e dom.Event) {
	e.PreventDefault()
	addClass(e.Target().(dom.HTMLElement), ClassItemOver)
}

func leaveItem(e dom.Event) {
	e.PreventDefault()
	removeClass(e.Target().(dom.HTMLElement), ClassItemOver)
}

func overRename(e dom.Event) {
	e.PreventDefault()
	addClass(e.Target().(dom.HTMLElement), ClassRenameOver)
}

func leaveRename(e dom.Event) {
	e.PreventDefault()
	removeClass(e.Target().(dom.HTMLElement), ClassRenameOver)
}

func overDelete(e dom.Event) {
	e.PreventDefault()
	addClass(e.Target().(dom.HTMLElement), ClassDeleteOver)
}

func leaveDelete(e dom.Event) {
	e.PreventDefault()
	removeClass(e.Target().(dom.HTMLElement), ClassDeleteOver)
}

func dragStartItem(e dom.Event) {
	draggedItemID := e.Target().ID()
	img := d.CreateElement("img").(*dom.HTMLImageElement)
	//img.Src = DragImage
	img.Src = "/img/ghost.png"
	e.(*dom.DragEvent).Get("dataTransfer").Call("setDragImage", img, -5, -5)
	e.(*dom.DragEvent).Get("dataTransfer").Call("setData", "draggedItemID", draggedItemID)
}

func dragItem(e dom.Event) {
	class := e.Target().Class()
	if class != nil {
		class.Add(ClassDraggedItem)
	}
}

func dropRename(elementId string) {
	resetAll()

	sl := strings.Split(elementId, "-")
	draggedItemIDDigit := sl[len(sl)-1]

	fmt.Println(elementId)
	fmt.Println(draggedItemIDDigit)

	el := d.GetElementByID(elementId).(dom.HTMLElement)

	if strings.HasPrefix(elementId, "folder") {
	} else {
		removeClass(el, ClassItemBookmarkLink)
		addClass(el, ClassItemBookmarkLinkEdited)
	}

	el.ParentNode().InsertBefore(d.GetElementByID("rename-input-box"), el.NextElementSibling())

	showRenameBox()
	if strings.HasPrefix(elementId, "folder") {
		draggedFldName := el.TextContent()
		setRenameFormValue(strings.Trim(draggedFldName, " "))
	} else {
		draggedBkmName := d.GetElementByID("bookmark-link-" + draggedItemIDDigit).TextContent()
		setRenameFormValue(draggedBkmName)
	}
	setRenameHiddenFormValue(elementId)
}

//
// HTML elements creation helpers
//
func createCloseDivButton(divID string) dom.HTMLElement {
	b := d.CreateElement("div").(*dom.HTMLDivElement)
	b.SetID("close-" + divID)
	b.SetClass("fa fa-times")
	b.AddEventListener("click", false, func(e dom.Event) { d.GetElementByID(divID).SetInnerHTML("") })
	return b
}

func createBookmark(bkmID string, bkmTitle string, bkmURL string, bkmFavicon string, bkmStarred bool, starred bool) dom.HTMLElement {
	// Link (actually a clickable div).
	//a := d.CreateElement("div").(*dom.HTMLDivElement)
	a := d.CreateElement("span").(*dom.HTMLSpanElement)
	a.SetTitle(bkmURL)
	a.SetAttribute("tabindex", "0")
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
	str.AddEventListener("click", false, func(e dom.Event) { starBookmark(bkmID, false) })

	if starred {
		a.SetID("bookmark-starred-link-" + bkmID)
		a.SetClass("bookmark-starred-link")
		md.SetID("bookmark-starred-" + bkmID)
		md.SetDraggable(false)
		str.SetID("bookmark-starred-star-" + bkmID)
		str.SetClass(ClassBookmarkStarred)
	} else {
		a.SetID("bookmark-link-" + bkmID)
		a.SetClass("bookmark-link")
		md.SetID("bookmark-" + bkmID)
		md.SetDraggable(true)
		str.SetID("bookmark-star-" + bkmID)
		if bkmStarred {
			str.SetClass(ClassBookmarkStarred)
		} else {
			str.SetClass(ClassBookmarkNotStarred)
		}
		md.AddEventListener("dragstart", false, dragStartItem)
		md.AddEventListener("drag", false, dragItem)
		a.AddEventListener("mouseover", false, func(e dom.Event) { mouseOverItem(e) })
		a.AddEventListener("keydown", false, func(e dom.Event) { keyDownItem(e) })
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
	md.SetAttribute("tabindex", "0")
	md.SetID("folder-" + fldID)
	md.SetDraggable(true)
	// Subfolders.
	ul := d.CreateElement("ul").(*dom.HTMLUListElement)
	ul.SetID("subfolders-" + fldID)

	md.AppendChild(d.CreateTextNode(" " + fldTitle))
	md.AppendChild(ul)

	md.AddEventListener("click", false, func(e dom.Event) { getChildrenItems(e, fldID) })
	md.AddEventListener("mouseover", false, func(e dom.Event) { mouseOverItem(e) })
	md.AddEventListener("keydown", false, func(e dom.Event) { keyDownItem(e) })
	md.AddEventListener("dragover", false, func(e dom.Event) { dragOverItem(e) })
	md.AddEventListener("dragleave", false, func(e dom.Event) { leaveItem(e) })
	md.AddEventListener("dragstart", false, func(e dom.Event) { dragStartItem(e) })
	md.AddEventListener("drag", false, func(e dom.Event) { dragItem(e) })
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
	var (
		err  error
		req  *http.Request
		resp *http.Response
	)

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

func searchBookmark() {

	go func() {

		setWait()
		hideRenameBox()
		hideImport()
		defer unsetWait()

		var (
			err     error
			resp    *http.Response
			dataBkm []types.Bookmark
		)

		s := d.GetElementByID("search-form-input").(*dom.HTMLInputElement).Value

		// Searching the bookmarks.
		if resp = sendRequest("/searchBookmarks/", []arg{{key: "search", val: s}}); resp.StatusCode != http.StatusOK {
			fmt.Println("searchBookmarks response code error")
			return
		}
		defer resp.Body.Close()

		if err = json.NewDecoder(resp.Body).Decode(&dataBkm); err != nil {
			fmt.Println("searchBookmarks JSON decoder error", err.Error())
			return
		}

		b := createCloseDivButton("search-result")
		d.GetElementByID("search-result").AppendChild(b)
		for _, bkm := range dataBkm {
			newBkm := createBookmark(strconv.Itoa(bkm.Id), bkm.Title, bkm.URL, bkm.Favicon, bkm.Starred, false)
			d.GetElementByID("search-result").AppendChild(newBkm)
		}
		d.GetElementByID("search-form-input").(*dom.HTMLInputElement).Set("value", "")
	}()
}

func starBookmark(bkmID string, forceUnstar bool) {
	go func() {
		var (
			star, starredBookmark bool
			starBookmarkDiv       dom.HTMLElement
			err                   error
			resp                  *http.Response
			data                  types.Bookmark // returned struct from server
		)

		if forceUnstar {
			star = false
			starredBookmark = true
		} else {
			star = true
			starredBookmark = isStarredBookmark(bkmID)
			starBookmarkDiv = d.GetElementByID("bookmark-star-" + bkmID).(dom.HTMLElement)
			if starredBookmark {
				star = false
			}
		}

		if resp = sendRequest("/starBookmark/", []arg{{key: "star", val: strconv.FormatBool(star)}, {key: "bookmarkId", val: bkmID}}); resp.StatusCode != http.StatusOK {
			fmt.Println("starBookmark response code error")
			return
		}
		defer resp.Body.Close()

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
			newBkm := createBookmark(bkmID, data.Title, data.URL, data.Favicon, data.Starred, true)

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
			//data newFolderStruct // returned struct from server
			data types.Folder // returned struct from server
		)

		fldName := d.GetElementByID("add-folder").(*dom.HTMLInputElement).Value

		if resp = sendRequest("/addFolder/", []arg{{key: "folderName", val: fldName}}); resp.StatusCode != http.StatusOK {
			fmt.Println("starBookmark response code error")
			return
		}
		defer resp.Body.Close()

		if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
			fmt.Println("addFolder JSON decoder error")
			return
		}

		newFld := createFolder(strconv.Itoa(int(data.Id)), data.Title, 0)

		rootFld := d.GetElementByID("subfolders-1")
		rootFld.InsertBefore(newFld.subFlds, rootFld.FirstChild())
		rootFld.InsertBefore(newFld.fld, rootFld.FirstChild())

		resetAddFolder()
	}()
}

func dropDelete(e dom.Event) {

	hideImport()

	e.PreventDefault()
	draggedItemID := e.(*dom.DragEvent).Get("dataTransfer").Call("getData", "draggedItemID").String()

	go func() {

		var (
			resp *http.Response
		)

		draggedItem := d.GetElementByID(draggedItemID)
		draggedItemIDDigit := strings.Split(draggedItemID, "-")[1]

		defer func() {
			removeClass(d.GetElementByID("delete-box").(*dom.HTMLDivElement), ClassDeleteOver)
		}()

		if strings.HasPrefix(draggedItemID, "folder") {
			if resp = sendRequest("/deleteFolder/", []arg{{key: "folderId", val: draggedItemIDDigit}}); resp.StatusCode != http.StatusOK {
				fmt.Println("dropDelete response code error")
				return
			}
			defer resp.Body.Close()

			children := d.GetElementByID("subfolders-" + draggedItemIDDigit)
			children.ParentNode().RemoveChild(children)
			draggedItem.ParentNode().RemoveChild(draggedItem)
		} else {
			if resp = sendRequest("/deleteBookmark/", []arg{{key: "bookmarkId", val: draggedItemIDDigit}}); resp.StatusCode != http.StatusOK {
				fmt.Println("dropDelete response code error")
				return
			}
			defer resp.Body.Close()

			draggedItem.ParentNode().RemoveChild(draggedItem)
		}

		removeClass(d.GetElementByID("delete-box").(dom.HTMLElement), ClassItemOver)
	}()

}

func dropFolder(e dom.Event) {

	e.PreventDefault()
	// Putting the following instruction inside the go routine does not work. I don't know why.
	u := e.(*dom.DragEvent).Get("dataTransfer").Call("getData", "URL").String()
	draggedItemID := e.(*dom.DragEvent).Get("dataTransfer").Call("getData", "draggedItemID").String()

	u = url.QueryEscape(u)

	go func() {

		draggedItem := d.GetElementByID(draggedItemID)

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

		defer func() {
			removeClass(droppedItem, ClassItemOver)
			if draggedItem != nil {
				removeClass(draggedItem.(*dom.HTMLDivElement), ClassDraggedItem)
			}
		}()

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
			if getClosest(droppedItem, "#subfolders-"+draggedItemIDDigit) != nil {
				fmt.Println("can not move a folder into one of its children")
				return
			}

			if resp = sendRequest("/moveFolder/", []arg{{key: "sourceFolderId", val: draggedItemIDDigit}, {key: "destinationFolderId", val: droppedItemIDDigit}}); resp.StatusCode != http.StatusOK {
				fmt.Println("dropFolder response code error")
				return
			}
			defer resp.Body.Close()

			droppedItemChildren.InsertBefore(draggedItemChildren, droppedItemChildren.FirstChild())
			droppedItemChildren.InsertBefore(draggedItem, droppedItemChildren.FirstChild())
			addClass(droppedItem, ClassItemFolderOpen)
			removeClass(droppedItem, ClassItemFolderClosed)

		} else if draggedItem != nil && strings.HasPrefix(draggedItemID, "bookmark") {

			if resp = sendRequest("/moveBookmark/", []arg{{key: "bookmarkId", val: draggedItemIDDigit}, {key: "destinationFolderId", val: droppedItemIDDigit}}); resp.StatusCode != http.StatusOK {
				fmt.Println("dropFolder response code error")
				return
			}
			defer resp.Body.Close()

			if hasChildrenFolders(droppedItemIDDigit) {
				droppedItemChildren.InsertBefore(draggedItem, droppedItemChildren.FirstChild())
				removeClass(droppedItem, ClassItemFolderClosed)
				removeClass(draggedItem.(*dom.HTMLDivElement), ClassDraggedItem)
				addClass(droppedItem, ClassItemFolderOpen)
			} else {
				draggedItem.ParentNode().RemoveChild(draggedItem)
			}

		} else {

			var dataBkm types.Bookmark
			if resp = sendRequest("/addBookmark/", []arg{{key: "bookmarkUrl", val: u}, {key: "destinationFolderId", val: droppedItemIDDigit}}); resp.StatusCode != http.StatusOK {
				fmt.Println("dropFolder response code error")
				return
			}
			defer resp.Body.Close()

			if err = json.NewDecoder(resp.Body).Decode(&dataBkm); err != nil {
				fmt.Println("getChildrenFolders JSON decoder error")
				return
			}

			newBkm := createBookmark(strconv.Itoa(dataBkm.Id), dataBkm.URL, dataBkm.URL, "", dataBkm.Starred, false)

			droppedItemChildren.InsertBefore(newBkm, droppedItemChildren.FirstChild())
			addClass(droppedItem, ClassItemFolderOpen)
			removeClass(droppedItem, ClassItemFolderClosed)
		}
	}()

}

func renameFolder(e dom.Event) {

	e.PreventDefault()

	go func() {

		var (
			resp *http.Response
		)

		fldID := d.GetElementByID("rename-hidden-input-box-form").(*dom.HTMLInputElement).Value
		fldName := d.GetElementByID("rename-input-box-form").(*dom.HTMLInputElement).Value

		sl := strings.Split(fldID, "-")
		fldIDDigit := sl[len(sl)-1]

		if strings.HasPrefix(fldID, "folder") {

			if resp = sendRequest("/renameFolder/", []arg{{key: "folderId", val: fldIDDigit}, {key: "folderName", val: fldName}}); resp.StatusCode != http.StatusOK {
				fmt.Println("renameFolder response code error")
				return
			}
			defer resp.Body.Close()

			d.GetElementByID(fldID).SetInnerHTML(" " + fldName)

			removeClass(d.GetElementByID(fldID).(*dom.HTMLDivElement), ClassDraggedItem)

		} else {

			if resp = sendRequest("/renameBookmark/", []arg{{key: "bookmarkId", val: fldIDDigit}, {key: "bookmarkName", val: fldName}}); resp.StatusCode != http.StatusOK {
				fmt.Println("renameBookmark response code error")
				return
			}
			defer resp.Body.Close()

			el := d.GetElementByID("bookmark-link-" + fldIDDigit).(dom.HTMLElement)
			el.SetInnerHTML(fldName)
			removeClass(el, ClassItemBookmarkLinkEdited)
			addClass(el, ClassItemBookmarkLink)

			removeClass(d.GetElementByID(fldID).(*dom.HTMLSpanElement), ClassDraggedItem)

		}

		hideRenameBox()
	}()

}

func getChildrenItems(e dom.Event, fldIDDigit string) {

	go func() {

		setWait()
		hideRenameBox()
		hideImport()
		defer unsetWait()

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
		defer resp.Body.Close()

		if err = json.NewDecoder(resp.Body).Decode(&dataFld); err != nil {
			fmt.Println("getChildrenFolders JSON decoder error", err.Error())
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
		defer resp.Body.Close()

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

func getWSBaseURL() string {
	document := js.Global.Get("window").Get("document")
	location := document.Get("location")

	wsProtocol := "ws"
	if location.Get("protocol").String() == "https:" {
		wsProtocol = "wss"
	}

	return fmt.Sprintf("%s://%s:%s/", wsProtocol, location.Get("hostname"), location.Get("port"))
}

func main() {

	// Websocket connection
	wsURL := getWSBaseURL() + "socket/"
	c, err := websocket.Dial(wsURL) // Blocks until connection is established
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := c.Read(buf) // Blocks until a WebSocket frame is received
			if err != nil {
				fmt.Println(err)
			}
			//fmt.Println(string(buf[:n]))

			var bkm types.Bookmark
			if err := json.Unmarshal(buf[:n], &bkm); err != nil {
				fmt.Println(err)
				return
			}
			newBkm := createBookmark(strconv.Itoa(bkm.Id), bkm.Title, bkm.URL, bkm.Favicon, false, false)

			rootChildrens := d.GetElementByID("subfolders-1")
			rootChildrens.InsertBefore(newBkm, rootChildrens.FirstChild())
		}
	}()

	// test
	//go func() {
	//	for {
	//		fmt.Println("still alive...")
	//		time.Sleep(3000 * time.Millisecond)
	//	}
	//}()

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
		dragOverItem(e)
	})
	fld.AddEventListener("dragleave", false, func(e dom.Event) {
		leaveItem(e)
	})
	fld.AddEventListener("drop", false, func(e dom.Event) {
		dropFolder(e)
	})

	db := d.GetElementByID("delete-box").(*dom.HTMLDivElement)
	db.AddEventListener("dragover", false, func(e dom.Event) {
		overDelete(e)
	})
	db.AddEventListener("dragleave", false, func(e dom.Event) {
		leaveDelete(e)
	})
	db.AddEventListener("drop", false, func(e dom.Event) {
		dropDelete(e)
	})

	// Import export links listeners.
	importLink := d.GetElementByID("import-box").(*dom.HTMLDivElement)
	importLink.AddEventListener("click", false, func(e dom.Event) {
		toogleDisplayImport()
	})
	exportLink := d.GetElementByID("export-box").(*dom.HTMLDivElement)
	exportLink.AddEventListener("click", false, func(e dom.Event) {
		openInParent("/export/")
	})

	// Import form listener.
	formImport := d.GetElementByID("import-file-form")
	formImport.AddEventListener("submit", false, func(e dom.Event) {
		importBookmarks(e)
	})

	// Search input listener.
	searchInput := d.GetElementByID("search-form-input")
	searchInput.AddEventListener("keyup", false, func(e dom.Event) {
		if changeTimer >= 0 {
			w.ClearTimeout(changeTimer)
		}
		changeTimer = w.SetTimeout(func() {
			clearSearchResults()
			searchBookmark()
			changeTimer = 0
		}, 400)
	})

	// Enter and Esc key listeners
	d.AddEventListener("keydown", false, func(e dom.Event) {
		if e.(*dom.KeyboardEvent).KeyCode == 13 {
			if isDisabled("add-folder") {
				renameFolder(e)
			} else {
				addFolder(e)
			}
		} else if e.(*dom.KeyboardEvent).KeyCode == 27 {
			resetAll()
		}
	})

	// Starred bookmarks listeners
	for _, e := range d.GetElementsByClassName("fa-star") {
		id := e.ID()
		idSplt := strings.Split(id, "-")
		idDigit := idSplt[len(idSplt)-1]
		e.AddEventListener("click", false, func(e dom.Event) {
			starBookmark(idDigit, true)
		})
	}
	for _, e := range d.GetElementsByClassName("bookmark-starred-link") {
		u := e.GetAttribute("Title")
		e.AddEventListener("click", false, func(e dom.Event) {
			openInParent(u)
		})
	}

}
