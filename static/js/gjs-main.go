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
	ClassDraggedItem        = "dragged-item"
	ClassItemOver           = "folder-over"
	ClassRenameOver         = "rename-over"
	ClassDeleteOver         = "delete-over"
	ClassItemFolder         = "folder"
	ClassItemFolderOpen     = "fa-folder-open-o"
	ClassItemFolderClosed   = "fa-folder-o"
	ClassItemBookmark       = "bookmark"
	ClassBookmarkStarred    = "fa fa-star"
	ClassBookmarkNotStarred = "fa fa-star-o"
	DragImage               = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAB4AAAAeCAQAAACROWYpAAAAAmJLR0QA/4ePzL8AAAAJcEhZcwACKaQAAimkAScW8mcAAAAHdElNRQfgBhEMEhox+QS1AAAC70lEQVQ4y5WVTWicdRDGf+/uJmSlaRJ7qKI0ftWrTUQUtfQDlQakh1q0pPWj1QZWRVMhHqpCraRgDwEpNIhSixS82EsLGsTiIado9OBJqR7SJWHbmGS3bbLZ7O7787DvbrMxtduZ2zDPO/Of55l5A1YxgQAfJsU8Q8EsSEBDJgJuK6d/9Ge9GG6iUROA8FVnTtjkOk/pJbdX4w1Y+JbXj4qId/ipLtrTENy4b1jqj6AVP25B99/o6+bg13L21UERB7ymh25ZuThz5D9QxD79zfX1uYmV4PjEjjtzxFniNFeBFvpopcROyHK50nqwcsLGfbl4ORxzq78649y8nSJ2OBuaddoxd5d/Kl3xYN3bI14P5PzEMR0H8LGsGyJw5rpPAHjqe4+5qG+vkESYspgSD+kMgE8uA8+7BWz2wj7xFV2wt1I7AQHh3uD4QGIYiEPMLhIsV1WMR1wiydo4cIa1yRNfmw5GBXCL08eimb7jgjlzZr3kvSK2edFr5sy55J4o6zNL2gW4sTxxskZIuw9Efp8JEWN2RpEHXVPL+0rn7MKhCzatyuz/eczz+m2MxTW0crsW8hAUcJ1/f2Ey+mKL7ZG3GRMxsC2KdNhcqzyi6fB+wG7Tg1Gw3+lwcmEyP1n4w3uigf1upjCZn8xnyy+KmHBYM2F3levnLFTmfUTnfNxtvne1xvN03gGf8Xl/eV3Ew5bn3AxWhebORfvFj/QKgJvrRLIdTDpyQHxJ874QiSSoSP2cqY+Hc/TAX2Azybr5tJikialnKXO6yGBwFgKoLogE+D69ZDlJDxtoL3ZvZALoYKLcOs4/TJVH472sZyQ4XN2s+u26m0w4OvLUN8QpcJ55oJndtBKyg13fcRCCqVVuaXXJSnNHV5XFfh33rlucIlN590Yc3/A9uuRgA4fXd82/uQzY5Aea98MGDiCA+wyHauDPLc+6qyFoxPvWkl8aiD9oxqcbrFprvtvsOf/UdLjpNv4XNfijnvVM2Hlz6L/COmiuQg3JqwAAAABJRU5ErkJggg=="
)

var (
	w dom.Window
	d dom.Document
	// The dragged element ID.
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
	hideImport()
	hideRenameBox()
	w.Parent().Open(url, "", "")
}

func isStarredBookmark(bkmID string) bool {
	return d.GetElementByID("bookmark-starred-link-"+bkmID) != nil
}
func hasChildrenFolders(fldID string) bool {
	return hasClass(d.GetElementByID("folder-"+fldID).(dom.HTMLElement), ClassItemFolderOpen)
}

func setWait() {
	d.GetElementsByTagName("body")[0].Class().SetString("wait")
}
func unsetWait() {
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
	e.PreventDefault()
	id := e.Target().(dom.HTMLElement).ID()
	dropRename(string(id))
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
	img.Src = DragImage
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

	hideImport()

	sl := strings.Split(elementId, "-")
	draggedItemIDDigit := sl[len(sl)-1]

	fmt.Println(elementId)
	fmt.Println(draggedItemIDDigit)

	showRenameBox()
	if strings.HasPrefix(elementId, "folder") {
		draggedFldName := d.GetElementByID(elementId).TextContent()
		setRenameFormValue(draggedFldName)
	} else {
		draggedBkmName := d.GetElementByID("bookmark-link-" + draggedItemIDDigit).TextContent()
		setRenameFormValue(draggedBkmName)
	}
	setRenameHiddenFormValue(elementId)
}

//
// HTML elements creation helpers
//
func createBookmark(bkmID string, bkmTitle string, bkmURL string, bkmFavicon string, bkmStarred bool, starred bool) dom.HTMLElement {

	// Link (actually a clickable div).
	a := d.CreateElement("div").(*dom.HTMLDivElement)
	a.SetTitle(bkmURL + " - type \"r\" to rename")
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
	md.SetTitle(fldTitle + " - type \"r\" to rename")
	md.SetClass(ClassItemFolder + " " + ClassItemFolderClosed)
	md.SetAttribute("tabindex", "0")
	md.SetID("folder-" + fldID)
	md.SetDraggable(true)
	// Subfolders.
	ul := d.CreateElement("ul").(*dom.HTMLUListElement)
	ul.SetID("subfolders-" + fldID)

	md.AppendChild(d.CreateTextNode(fldTitle))
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

func starBookmark(bkmID string, forceUnstar bool) {

	go func() {

		var (
			star, starredBookmark bool
			starBookmarkDiv       dom.HTMLElement
			err                   error
			resp                  *http.Response
			data                  newBookmarkStruct // returned struct from server
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
	draggedItemID := e.(*dom.DragEvent).Get("dataTransfer").Call("getData", "draggedItemID").String()

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
			removeClass(draggedItem.(*dom.HTMLDivElement), ClassDraggedItem)
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

			droppedItemChildren.InsertBefore(draggedItemChildren, droppedItemChildren.FirstChild())
			droppedItemChildren.InsertBefore(draggedItem, droppedItemChildren.FirstChild())
			addClass(droppedItem, ClassItemFolderOpen)
			removeClass(droppedItem, ClassItemFolderClosed)

		} else if draggedItem != nil && strings.HasPrefix(draggedItemID, "bookmark") {

			if resp = sendRequest("/moveBookmark/", []arg{{key: "bookmarkId", val: draggedItemIDDigit}, {key: "destinationFolderId", val: droppedItemIDDigit}}); resp.StatusCode != http.StatusOK {
				fmt.Println("dropFolder response code error")
				return
			}

			if hasChildrenFolders(droppedItemIDDigit) {
				droppedItemChildren.InsertBefore(draggedItem, droppedItemChildren.FirstChild())
				removeClass(droppedItem, ClassItemFolderClosed)
				removeClass(draggedItem.(*dom.HTMLDivElement), ClassDraggedItem)
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

			d.GetElementByID(fldID).SetInnerHTML(fldName)

		} else {

			if resp = sendRequest("/renameBookmark/", []arg{{key: "bookmarkId", val: fldIDDigit}, {key: "bookmarkName", val: fldName}}); resp.StatusCode != http.StatusOK {
				fmt.Println("renameBookmark response code error")
				return
			}

			d.GetElementByID("bookmark-link-" + fldIDDigit).SetInnerHTML(fldName)

		}

		removeClass(d.GetElementByID(fldID).(*dom.HTMLDivElement), ClassDraggedItem)
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

	// Enter key listener
	d.AddEventListener("keydown", false, func(e dom.Event) {
		if e.(*dom.KeyboardEvent).KeyCode == 13 {
			if isDisabled("add-folder") {
				renameFolder(e)
			} else {
				addFolder(e)
			}
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
