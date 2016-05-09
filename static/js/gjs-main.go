package main

import "honnef.co/go/js/dom"

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

	d.GetElementByID("test").AddEventListener("click", false, toogleDisplayImport)
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

func renameFolder() {
	print("renameFolder")
}

func addFolder() {
	print("addFolder")
}

func dragBookmark(e dom.Event) {
	e.Set("dragItemId", e.Target().ID())
}

func createBookmark(bkmID string, bkmTitle string, bkmURL string, bkmFavicon string, bkmStarred bool, starred bool) dom.HTMLElement {

	// Link (actually a clickable div).
	a := d.CreateElement("div").(dom.HTMLDivElement)
	a.SetTitle(bkmURL)
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

func main() {

	// Bind enter key to add or rename a folder.
	d.AddEventListener("keydown", false, func(e dom.Event) {
		ke := e.(*dom.KeyboardEvent)
		if ke.KeyCode == 13 {
			e.PreventDefault()

			dsbl := d.GetElementByID("add-folder").HasAttribute("disabled")
			if dsbl {
				renameFolder()
			} else {
				addFolder()
			}
		}
	})

}
