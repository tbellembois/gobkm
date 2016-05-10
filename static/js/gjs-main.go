package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"honnef.co/go/js/dom"
)

const (
	ClassItemOver           = "folder-over"
	ClassItemFolder         = "folder"
	ClassItemFolderOpen     = "fa-folder-open-o"
	ClassItemFolderClosed   = "fa-folder-o"
	ClassBookmarkStarred    = "fa fa-star"
	ClassBookmarkNotStarred = "fa fa-star-o"
	ClassItemBookmark       = "bookmark"
)

var (
	w dom.Window
	d dom.Document
)

func init() {
	w = dom.GetWindow()
	d = w.Document()

	d.GetElementByID("add-folder-button").AddEventListener("click", false, addFolder)
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
	d.GetElementByID(id).SetNodeValue("")
}
func setItemValue(id string, val string) {
	d.GetElementByID(id).SetNodeValue(val)
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

func hasChildrenFolders(fldID string) bool {
	return hasClass(d.GetElementByID("folder-"+fldID).(dom.HTMLElement), ClassItemFolderOpen)
}

func undisplayChildrenFolders(fldID string) {
	// Removing folder content.
	d.GetElementByID("subfolders-" + fldID).SetInnerHTML("")
	// Changing folder icon.
	setClass(d.GetElementByID("folder-"+fldID).(dom.HTMLElement), ClassItemFolder+ClassItemFolderClosed)
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

func dragBookmark(e dom.Event) {
	e.(dom.DragEvent).Set("dragItemId", e.Target().ID())
}

func dragFolder(e dom.Event) {
	e.(dom.DragEvent).Set("dragItemId", e.Target().ID())
}

func createBookmark(bkmID string, bkmTitle string, bkmURL string, bkmFavicon string, bkmStarred bool, starred bool) dom.HTMLElement {

	// Link (actually a clickable div).
	a := d.CreateElement("div").(dom.HTMLDivElement)
	a.SetTitle(bkmURL)
	a.AppendChild(d.CreateTextNode(bkmURL))
	a.SetAttribute("onclick", "openInParent('"+bkmURL+"');")
	// Main div.
	md := d.CreateElement("div").(dom.HTMLDivElement)
	md.SetClass(ClassItemBookmark)
	// Favicon.
	fav := d.CreateElement("img").(dom.HTMLImageElement)
	fav.Src = bkmFavicon
	fav.SetClass("favicon")
	// Star.
	str := d.CreateElement("div").(dom.HTMLDivElement)
	str.SetAttribute("onclick", "starBookmark('"+bkmID+"');")

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
	}

	md.AppendChild(str)
	md.AppendChild(fav)
	md.AppendChild(a)

	if !starred {
		md.AddEventListener("dragstart", false, dragBookmark)
	}

	return md
}

type folderStruct struct {
	fld     dom.HTMLDivElement
	subFlds dom.HTMLUListElement
}

func createFolder(fldID string, fldTitle string, nbChildrenFolders int) folderStruct {

	// Main div.
	md := d.CreateElement("div").(*dom.HTMLDivElement)
	md.SetTitle(fldTitle)
	md.SetClass(ClassItemFolder + " " + ClassItemFolderClosed)
	md.SetID("folder-" + fldID)
	md.SetDraggable(true)
	md.SetAttribute("onclick", "getChildrenFolders(event, "+fldID+");")
	// Subfolders.
	ul := d.CreateElement("ul").(*dom.HTMLUListElement)
	ul.SetID("subfolders-" + fldID)

	md.AppendChild(d.CreateTextNode(fldTitle))
	md.AppendChild(ul)

	return folderStruct{fld: *md, subFlds: *ul}
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

type newFolderStruct struct {
	FolderId    int64
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

		newFld := createFolder(string(data.FolderId), data.FolderTitle, 0)

		rootFld := d.GetElementByID("subfolders-1")
		rootFld.InsertBefore(newFld.fld, rootFld.FirstChild())
		rootFld.InsertBefore(newFld.subFlds, rootFld.FirstChild())
	}()
}

func dropRename(e dom.Event) {

	draggedItem := e.Target()
	draggedItemID := e.(dom.DropEvent).Get("dragItemId")
	draggedItemIDDigit := strings.Split(draggedItemID, "-")[1]

	if strings.HasPrefix(draggedItemID, "folder") {
		draggedFldName := d.GetElementByID(draggedItemID).InnerHTML()
	}

}

func renameFolder(e dom.Event) {

	go func() {

		fmt.Println("renameFolder")

		fldID := d.GetElementByID("rename-hidden-input-box-form").(*dom.HTMLInputElement).Value
		fldName := d.GetElementByID("rename-input-box-form").(*dom.HTMLInputElement).Value

		fldIDDigit := strings.Split(fldID, "-")[1]

		if strings.HasPrefix(fldID, "folder") {
			req, err := http.NewRequest("GET", "/renameFolder/?folderId="+fldIDDigit+"&folderName="+fldName, nil)

			if err != nil {
				return
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Println("renameFolder response code error")
				return
			}

			d.GetElementByID()

		}

	}()

}

func main() {

	// Bind enter key to add or rename a folder.
	d.AddEventListener("keydown", false, func(e dom.Event) {
		ke := e.(*dom.KeyboardEvent)
		if ke.KeyCode == 13 {
			e.PreventDefault()
		}
	})

}
