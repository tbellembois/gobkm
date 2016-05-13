package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/tbellembois/gobkm/types"
	"honnef.co/go/js/dom"
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

func init() {
	w = dom.GetWindow()
	d = w.Document()

}

// Utils functions.
func isHidden(id string) bool {
	return d.GetElementByID(id).(dom.HTMLElement).Style().GetPropertyValue("display") == "block"
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

type arg struct {
	key string
	val string
}

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

func hasChildrenFolders(fldID string) bool {
	return hasClass(d.GetElementByID("folder-"+fldID).(dom.HTMLElement), ClassItemFolderOpen)
}

func undisplayChildrenFolders(fldID string) {
	// Removing folder content.
	d.GetElementByID("subfolders-" + fldID).SetInnerHTML("")
	// Changing folder icon.
	setClass(d.GetElementByID("folder-"+fldID).(dom.HTMLElement), ClassItemFolder+" "+ClassItemFolderClosed)
}

func isStarredBookmark(bkmID string) bool {
	return d.GetElementByID("bookmark-starred-link-"+bkmID) != nil
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
	d.GetElementByID("rename-input-box-form").(*dom.HTMLInputElement).Select()
}
func setRenameHiddenFormValue(val string) {
	setItemValue("rename-hidden-input-box-form", val)
}

func hideImport() {
	print("hideImport")
	hideItem("import-input-box")
}

func toogleDisplayImport(e dom.Event) {
	print("toogleDisplayImport")
	if isHidden("import-input-box") {
		showItem("import-input-box")
	} else {
		hideItem("import-input-box")
	}
}

func setWait() {
	print("setWait")
	d.GetElementsByTagName("body")[0].Class().SetString("wait")
}

func unsetWait() {
	print("unsetWait")
	d.GetElementsByTagName("body")[0].Class().Remove("wait")
}

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

func starBookmark(bkmID string) {
	go func() {

		star := true
		starredBookmark := isStarredBookmark(bkmID)
		starBookmarkDiv := d.GetElementByID("bookmark-star-" + bkmID).(dom.HTMLElement)
		fmt.Printf("starBookmarkDiv:%t\n", starredBookmark)

		if starredBookmark {
			star = false
		}

		req, err := http.NewRequest("GET", "/starBookmark/?star="+strconv.FormatBool(star)+"&bookmarkId="+bkmID, nil)
		if err != nil {
			return
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("starBookmark request error")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
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

			var data newBookmarkStruct
			err = json.NewDecoder(resp.Body).Decode(&data)
			if err != nil {
				fmt.Println("starBookmark JSON decoder error")
				return
			}
			newBkm := createBookmark(bkmID, data.BookmarkTitle, data.BookmarkURL, data.BookmarkFavicon, data.BookmarkStarred, true)
			fmt.Printf("data:%v\n", data)

			li := d.CreateElement("li").(*dom.HTMLLIElement)
			li.AppendChild(newBkm)
			d.GetElementByID("starred").AppendChild(li)
		}
	}()
}

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

func createFolder(fldID string, fldTitle string, nbChildrenFolders int) folderStruct {

	fmt.Println("createFolder:fldID=" + fldID)
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

func displaySubfolder(pFldID string, fldID string, fldTitle string, nbChildrenFolders int) {

	fmt.Printf("displaySubfolder:pFldID=%s,fldID=%s,", pFldID, fldID)
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

type newFolderStruct struct {
	FolderID    int64
	FolderTitle string
}

func addFolder(e dom.Event) {

	go func() {

		fmt.Println("addFolder")

		fldName := d.GetElementByID("add-folder").(*dom.HTMLInputElement).Value

		req, err := http.NewRequest("GET", "/addFolder/?folderName="+fldName, nil)

		if err != nil {
			return
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("addFolder request error")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("addFolder response code error")
			return
		}

		var data newFolderStruct
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
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

		draggedItem := d.GetElementByID(draggedItemID)
		draggedItemIDDigit := strings.Split(draggedItemID, "-")[1]

		if strings.HasPrefix(draggedItemID, "folder") {
			req, err := http.NewRequest("GET", "/deleteFolder/?folderId="+draggedItemIDDigit, nil)
			if err != nil {
				return
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("dropDelete request error")
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Println("dropDelete response code error")
				return
			}

			children := d.GetElementByID("subfolders-" + draggedItemIDDigit)
			children.ParentNode().RemoveChild(children)
			draggedItem.ParentNode().RemoveChild(draggedItem)
		} else {
			req, err := http.NewRequest("GET", "/deleteBookmark/?bookmarkId="+draggedItemIDDigit, nil)
			if err != nil {
				return
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("dropDelete request error")
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Println("dropDelete response code error")
				return
			}

			draggedItem.ParentNode().RemoveChild(draggedItem)
		}
	}()

}

func dropFolder(e dom.Event) {

	e.PreventDefault()

	go func() {

		draggedItem := d.GetElementByID(draggedItemID)
		var (
			draggedItemIDDigit  string
			draggedItemChildren dom.Element
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
			//TODO
			// Can not move a folder into one of its children.
			//TODO

			req, err := http.NewRequest("GET", "/moveFolder/?sourceFolderId="+draggedItemIDDigit+"&destinationFolderId="+droppedItemIDDigit, nil)
			if err != nil {
				return
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("dropFolder request error")
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Println("dropFolder response code error")
				return
			}

			droppedItemChildren.InsertBefore(draggedItemChildren, droppedItemChildren.FirstChild())
			droppedItemChildren.InsertBefore(draggedItem, droppedItemChildren.FirstChild())
			removeClass(droppedItem, ClassItemOver)
			removeClass(droppedItem, ClassItemFolderClosed)
			addClass(droppedItem, ClassItemFolderOpen)
		} else if draggedItem != nil && strings.HasPrefix(draggedItemID, "bookmark") {

			req, err := http.NewRequest("GET", "/moveBookmark/?bookmarkId="+draggedItemIDDigit+"&destinationFolderId="+droppedItemIDDigit, nil)
			if err != nil {
				return
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("dropFolder request error")
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
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
			url := e.(*dom.DragEvent).Get("dataTransfer").Get("items").Get("length")
			//url := js.Global.Get("dataTransfer").Call("getData", "URL")
			fmt.Printf("url:%v", url)
			//Url, err := url.Parse(url)

		}
	}()

}

func dropRename(e dom.Event) {

	//draggedItemID = e.(*dom.DragEvent).Get("dragItemId").String()
	draggedItemIDDigit := strings.Split(draggedItemID, "-")[1]

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

func renameFolder(e dom.Event) {

	go func() {

		var (
			resp *http.Response
		)

		fldID := d.GetElementByID("rename-hidden-input-box-form").(*dom.HTMLInputElement).Value
		fldName := d.GetElementByID("rename-input-box-form").(*dom.HTMLInputElement).Value
		fldIDDigit := strings.Split(fldID, "-")[1]

		if strings.HasPrefix(fldID, "folder") {

			if resp = sendRequest("/renameFolder/", []arg{{key: "folderId", val: fldIDDigit}, {key: "folderName", val: fldName}}); resp.StatusCode != http.StatusOK {
				fmt.Println("renameFolder response code error")
				return
			}

			if resp.StatusCode != http.StatusOK {
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

}
