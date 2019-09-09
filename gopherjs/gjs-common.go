package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/tbellembois/gobkm/types"
	"honnef.co/go/js/dom"
	"honnef.co/go/js/xhr"
)

var (
	window   dom.Window
	document dom.Document
	rootDiv  *dom.HTMLDivElement
	jQuery   func(...interface{}) jquery.JQuery
)

func init() {
	window = dom.GetWindow()
	document = window.Document()
	jQuery = jquery.NewJQuery
}

// processResults is called by select2 to return expected field names
// see select2 plugin documentation
func processResults(data map[string]interface{}) map[string]interface{} {
	var r map[string]interface{}
	r = make(map[string]interface{}, 0)

	var s []interface{}
	for _, value := range data {
		id := value.(map[string]interface{})["id"]
		name := value.(map[string]interface{})["name"]

		s = append(s, map[string]interface{}{
			"id":   id,
			"text": name,
		})
	}
	r["results"] = s

	return r
}

// select2ify turns a select into select2
// see select2 plugin documentation
func select2ify(id string) {
	var p map[string]interface{}

	p = make(map[string]interface{}, 0)
	p["placeholder"] = "tags"
	p["tags"] = "tags"
	p["ajax"] = map[string]interface{}{
		"url":            "/getTags",
		"dataType":       "json",
		"processResults": processResults,
	}

	jQuery(fmt.Sprintf("select#%supdateBookmarkInputTags", id)).Call("select2", p)
	jQuery(fmt.Sprintf("select#%screateBookmarkInputTags", id)).Call("select2", p)
}

// getBranchNodes remotely gets the children nodes of "nodeID"
func getBranchNodes(nodeID string) types.Folder {
	var (
		data  []byte
		err   error
		nodes types.Folder
	)

	if data, err = xhr.Send("GET", "/getBranchNodes/?parentId="+nodeID, nil); err != nil {
		fmt.Printf("error getting nodes of %s", nodeID)
	}

	// decoding response
	if err = json.NewDecoder(strings.NewReader(string(data))).Decode(&nodes); err != nil {
		fmt.Println("error decoding the JSON")
	}

	return nodes
}

// searchBookmark remotely search bookmarks
// s: the search string
func searchBookmark(s string) {
	var (
		err  error
		data []byte
		bkms []types.Bookmark
	)

	if data, err = xhr.Send("GET", "/searchBookmarks/?search="+s, nil); err != nil {
		fmt.Println("error search bookmark")
	}

	datar := bytes.NewReader(data)
	decoder := json.NewDecoder(datar)

	if err = decoder.Decode(&bkms); err != nil {
		fmt.Println("error decoding reponse from search")
	}

	// cleaning former search results and appending new ones
	jQuery("#searchResults").SetHtml("")
	for _, b := range bkms {
		jQuery("#searchResults").Append(createSearchBookmarkNode(b))
	}

}

// starBookmark remotely stars the bookmark
// id: the bookmark id
func starBookmark(id string) {
	var (
		err  error
		data []byte
		b    types.Bookmark
	)

	if data, err = xhr.Send("GET", "/starBookmark/?star=true&bookmarkId="+id, nil); err != nil {
		fmt.Printf("error starring bookmark %s", id)
	}

	datar := bytes.NewReader(data)
	decoder := json.NewDecoder(datar)

	if err = decoder.Decode(&b); err != nil {
		fmt.Printf("error decoding reponse from star %s", id)
	}

	// adding bookmark to the star div
	document.GetElementByID("star").AppendChild(createStarredBookmarkNode(b))

	// changing bookmark star icon
	jQuery(fmt.Sprintf("span#%sstarspan", id)).RemoveClass("mdi-star-outline")
	jQuery(fmt.Sprintf("span#%sstarspan", id)).AddClass("mdi-star")
}

// unstarBookmark star the bookmark with the
// id: the bookmark id
func unstarBookmark(id string) {
	var (
		err  error
		data []byte
		b    types.Bookmark
	)

	if data, err = xhr.Send("GET", "/starBookmark/?star=false&bookmarkId="+id, nil); err != nil {
		fmt.Printf("error unstarring bookmark %s", id)
	}

	datar := bytes.NewReader(data)
	decoder := json.NewDecoder(datar)

	if err = decoder.Decode(&b); err != nil {
		fmt.Printf("error decoding reponse from star %s", id)
	}

	// removing bookmark from the star div
	jQuery(fmt.Sprintf("button#%sstarred", id)).Remove()

	// changing bookmark star icon
	jQuery(fmt.Sprintf("span#%sstarspan", id)).RemoveClass("mdi-star")
	jQuery(fmt.Sprintf("span#%sstarspan", id)).AddClass("mdi-star-outline")
}

// updateBookmark remotely updates the bookmark b
func updateBookmark(b types.Bookmark) {
	var (
		err     error
		payload []byte
	)

	if payload, err = json.Marshal(b); err != nil {
		fmt.Printf("error marshalling bookmark %s", b.Title)
	}

	if _, err = xhr.Send("POST", "/renameBookmark/", payload); err != nil {
		fmt.Printf("error updating bookmark %s", b.Title)
	}

	// removing create form
	jQuery(fmt.Sprintf("#%dupdateBookmark", b.Id)).Remove()

	// updating html
	jQuery(fmt.Sprintf("a#%dbookmarkLink", b.Id)).SetHtml(b.Title)
	jQuery(fmt.Sprintf("a#%dbookmarkLink", b.Id)).SetAttr("href", b.URL)

	// deleting former tags
	jQuery(fmt.Sprintf(".%dbadge", b.Id)).Remove()

	// adding new ones
	for _, t := range b.Tags {
		jQuery(createSmallTagNode(*t, b.Id)).InsertAfter(fmt.Sprintf("a#%dbookmarkLink", b.Id))
	}

	// hiding buttons
	hideActionButtons("")

}

// createBookmark remotely creates the bookmark b
func createBookmark(b types.Bookmark) {
	var (
		err     error
		payload []byte
	)

	if payload, err = json.Marshal(b); err != nil {
		fmt.Printf("error marshalling bookmark %s", b.Title)
	}

	if _, err = xhr.Send("POST", "/addBookmark/", payload); err != nil {
		fmt.Printf("error creating bookmark %s", b.Title)
	}

	// removing create form
	jQuery(fmt.Sprintf("#%dcreateBookmark", b.Folder.Id)).Remove()

	// getting and cleaning the bookmark parent directory div
	parentD := document.GetElementByID(fmt.Sprintf("%dfolderBody", b.Folder.Id)).(*dom.HTMLDivElement)
	parentD.SetInnerHTML("")

	// lazily getting the bookmark parent folder children nodes
	// and refreshing them
	cnodes := getBranchNodes(fmt.Sprintf("%d", b.Folder.Id))
	for _, c := range cnodes.Folders {
		displayNode(*c, parentD)
	}
	for _, c := range cnodes.Bookmarks {
		displayNode(*c, parentD)
	}
}

// moveBookmark remotely moves the bookmark b
func moveBookmark(b types.Bookmark) {
	var (
		err     error
		payload []byte
	)

	if payload, err = json.Marshal(b); err != nil {
		fmt.Printf("error marshalling bookmark %s", b.Title)
	}

	if _, err = xhr.Send("PUT", "/moveBookmark/", payload); err != nil {
		fmt.Printf("error moving bookmark %s", b.Title)
	}

	// getting the cutted bookmark id
	cuttedid := jQuery("input[type=hidden][name=cutednodeid]").Val()

	// deleting the old bookmark div
	jQuery(fmt.Sprintf("#%sbookmarkMainDiv", cuttedid)).Remove()

	// getting and cleaning the folder parent directory div
	parentD := document.GetElementByID(fmt.Sprintf("%dfolderBody", b.Folder.Id)).(*dom.HTMLDivElement)
	parentD.SetInnerHTML("")

	// lazily getting the folder parent folder children nodes
	// and refreshing them
	cnodes := getBranchNodes(fmt.Sprintf("%d", b.Folder.Id))
	for _, c := range cnodes.Folders {
		displayNode(*c, parentD)
	}
	for _, c := range cnodes.Bookmarks {
		displayNode(*c, parentD)
	}

	// resetting the cutednodeid
	jQuery("input[type=hidden][name=cutednodeid]").SetVal("")
}

// moveFolder remotely moves the folder f
func moveFolder(f types.Folder) {
	var (
		err     error
		payload []byte
	)

	if payload, err = json.Marshal(f); err != nil {
		fmt.Printf("error marshalling folder %s", f.Title)
	}

	if _, err = xhr.Send("PUT", "/moveFolder/", payload); err != nil {
		fmt.Printf("error moving folder %s", f.Title)
	}

	// getting the cutted folder id
	cuttedid := jQuery("input[type=hidden][name=cutednodeid]").Val()

	// deleting the old folder div
	jQuery(fmt.Sprintf("#%sfolderMainDiv", cuttedid)).Remove()

	// getting and cleaning the folder parent directory div
	parentD := document.GetElementByID(fmt.Sprintf("%dfolderBody", f.Parent.Id)).(*dom.HTMLDivElement)
	parentD.SetInnerHTML("")

	// lazily getting the folder parent folder children nodes
	// and refreshing them
	cnodes := getBranchNodes(fmt.Sprintf("%d", f.Parent.Id))
	for _, c := range cnodes.Folders {
		displayNode(*c, parentD)
	}
	for _, c := range cnodes.Bookmarks {
		displayNode(*c, parentD)
	}

	// resetting the cutednodeid
	jQuery("input[type=hidden][name=cutednodeid]").SetVal("")
}

// createFolder remotely creates the folder f
func createFolder(f types.Folder) {
	var (
		err     error
		payload []byte
	)

	if payload, err = json.Marshal(f); err != nil {
		fmt.Printf("error marshalling folder %s", f.Title)
	}

	if _, err = xhr.Send("POST", "/addFolder/", payload); err != nil {
		fmt.Printf("error creating folder %s", f.Title)
	}

	// removing create form
	jQuery(fmt.Sprintf("#%dcreateFolder", f.Parent.Id)).Remove()

	// getting and cleaning the folder parent directory div
	parentD := document.GetElementByID(fmt.Sprintf("%dfolderBody", f.Parent.Id)).(*dom.HTMLDivElement)
	parentD.SetInnerHTML("")

	// lazily getting the folder parent folder children nodes
	// and refreshing them
	cnodes := getBranchNodes(fmt.Sprintf("%d", f.Parent.Id))
	for _, c := range cnodes.Folders {
		displayNode(*c, parentD)
	}
	for _, c := range cnodes.Bookmarks {
		displayNode(*c, parentD)
	}
}

// updateFolder remotely updates the folder f
func updateFolder(f types.Folder) {
	var (
		err     error
		payload []byte
	)

	if payload, err = json.Marshal(f); err != nil {
		fmt.Printf("error marshalling folder %s", f.Title)
	}

	if _, err = xhr.Send("POST", "/renameFolder/", payload); err != nil {
		fmt.Printf("error renaming folder %s", f.Title)
	}

	// removing create form
	jQuery(fmt.Sprintf("#%dupdateFolder", f.Id)).Remove()

	// updating html
	jQuery(fmt.Sprintf("button#%dfolderLink", f.Id)).SetHtml(f.Title)

	// hiding buttons
	hideActionButtons("")
}

// deleteFolder remotely deletes the folder
// id: the bookmark id
func deleteFolder(id string) {
	var (
		err error
	)

	if _, err = xhr.Send("DELETE", "/deleteFolder/?itemId="+id, nil); err != nil {
		fmt.Printf("error deleting folder %s", id)
	}

	jQuery("div#" + id + "folderMainDiv").Remove()
}

// deleteBookmark remotely deletes the bookmark
// id: the bookmark id
func deleteBookmark(id string) {
	var (
		err error
	)

	if _, err = xhr.Send("DELETE", "/deleteBookmark/?itemId="+id, nil); err != nil {
		fmt.Printf("error deleting bookmark %s", id)
	}

	jQuery("div#" + id + "bookmarkMainDiv").Remove()
}

// createStarredBookmarkNode creates a starred bookmark HTML element
func createStarredBookmarkNode(b types.Bookmark) *dom.HTMLButtonElement {

	buttonDiv := document.CreateElement("button").(*dom.HTMLButtonElement)
	buttonDiv.SetID(fmt.Sprintf("-%dstarred", b.Id))
	buttonDiv.SetClass("btn btn-outline-dark")
	buttonDiv.SetInnerHTML(b.Title)
	buttonDiv.AddEventListener("click", false, func(event dom.Event) {
		window.Open(b.URL, "", "")
	})

	return buttonDiv
}

// createTagNode creates a tag HTML element
func createTagNode(id, title string) *dom.HTMLButtonElement {

	buttonDiv := document.CreateElement("button").(*dom.HTMLButtonElement)
	buttonDiv.SetID(id + "tag")
	buttonDiv.SetClass("btn btn-outline-dark")
	buttonDiv.SetInnerHTML(title)

	buttonDiv.AddEventListener("click", false, func(event dom.Event) {
		jQuery("input#searchInput").SetVal(title)
		jQuery("input#searchInput").Focus()
		jQuery("input#searchInput").Trigger("keypress")
	})

	return buttonDiv
}

// createTagNode creates a small tag HTML element
func createSmallTagNode(t types.Tag, bkmid int) *dom.HTMLSpanElement {

	tagSpan := document.CreateElement("span").(*dom.HTMLSpanElement)
	tagSpan.SetClass(fmt.Sprintf("%dbadge badge badge-pill badge-dark", bkmid))
	tagSpan.SetAttribute("value", fmt.Sprintf("%d", t.Id))
	tagSpan.SetInnerHTML(t.Name)
	return tagSpan

}

// createSearchBookmarkNode creates a bookmark HTML element
func createSearchBookmarkNode(n types.Bookmark) *dom.HTMLDivElement {

	topDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	topDiv.SetClass("container-fluid")

	mainDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	mainDiv.SetClass("row bookmark")
	mainDiv.SetID(fmt.Sprintf("%dsearchBookmarkMainDiv", n.Id))

	actionDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	actionDiv.SetClass("row")
	actionDiv.SetID(fmt.Sprintf("%dactionDiv", n.Id))

	linkDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	linkDiv.SetClass("col col-10")

	buttonDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	buttonDiv.SetClass("col col-2")

	link := document.CreateElement("a").(*dom.HTMLAnchorElement)
	link.SetAttribute("href", n.URL)
	link.SetAttribute("target", "_blank")
	link.SetID(fmt.Sprintf("%dbookmarkLink", n.Id))
	link.SetInnerHTML(n.Title)

	favicon := document.CreateElement("img").(*dom.HTMLImageElement)
	favicon.SetClass("favicon")
	favicon.SetAttribute("src", n.Favicon)

	treeButton := createButton("file-tree", fmt.Sprintf("%dtree", n.Id), false, "bookmarkbtn", "float-right")
	treeButton.AddEventListener("click", false, func(event dom.Event) {
		// building and reversing parent id slice
		a := make([]int, 0)
		p := n.Folder
		for p != nil {
			// todo: fix this
			if p.Id == 1 {
				p.Id = 0
			}
			a = append(a, p.Id)
			p = p.Parent
		}
		for i := len(a)/2 - 1; i >= 0; i-- {
			opp := len(a) - 1 - i
			a[i], a[opp] = a[opp], a[i]
		}

		for _, i := range a {
			if !jQuery(fmt.Sprintf("div#collapse%d", i)).HasClass("show") {
				jQuery(fmt.Sprintf("button#%dfolderLink", i)).Call("trigger", "click")
			}
		}

		jQuery(fmt.Sprintf("#-%dbookmarkMainDiv", n.Id)).SetCss("border-left", "20px solid yellow")

		go func() {
			time.Sleep(500 * time.Millisecond)
			document.GetElementByID(fmt.Sprintf("-%dbookmarkMainDiv", n.Id)).Underlying().Call("scrollIntoView")
		}()
	})

	buttonDiv.AppendChild(treeButton)

	linkDiv.AppendChild(favicon)
	linkDiv.AppendChild(link)
	for _, t := range n.Tags {
		linkDiv.AppendChild(createSmallTagNode(*t, n.Id))
	}

	mainDiv.AppendChild(linkDiv)
	mainDiv.AppendChild(buttonDiv)
	topDiv.AppendChild(mainDiv)
	topDiv.AppendChild(actionDiv)

	return topDiv
}

// createBookmarkNode creates a bookmark HTML element
func createBookmarkNode(n types.Bookmark) *dom.HTMLDivElement {

	// negating bookmark ids to avoid conflict with folders
	n.Id = -n.Id

	topDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	topDiv.SetClass("container-fluid")

	mainDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	mainDiv.SetClass("row bookmark")
	mainDiv.SetID(fmt.Sprintf("%dbookmarkMainDiv", n.Id))

	actionDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	actionDiv.SetClass("row")
	actionDiv.SetID(fmt.Sprintf("%dactionDiv", n.Id))

	linkDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	linkDiv.SetClass("col col-10")

	buttonDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	buttonDiv.SetClass("col col-2")

	link := document.CreateElement("a").(*dom.HTMLAnchorElement)
	link.SetAttribute("href", n.URL)
	link.SetAttribute("target", "_blank")
	link.SetID(fmt.Sprintf("%dbookmarkLink", n.Id))
	link.SetInnerHTML(n.Title)

	favicon := document.CreateElement("img").(*dom.HTMLImageElement)
	favicon.SetClass("favicon")
	favicon.SetAttribute("src", n.Favicon)

	menuButton := createButton("menu", fmt.Sprintf("%dmenu", n.Id), false, "bookmarkbtn", "float-right")
	cutButton := createButton("content-cut", fmt.Sprintf("%dcut", n.Id), true, "bookmarkbtn", "float-right")
	deleteButton := createButton("delete-outline", fmt.Sprintf("%ddelete", n.Id), true, "bookmarkbtn", "float-right")
	editButton := createButton("pencil-outline", fmt.Sprintf("%dedit", n.Id), true, "bookmarkbtn", "float-right")
	var starButton *dom.HTMLButtonElement
	if n.Starred {
		starButton = createButton("star", fmt.Sprintf("%dstar", n.Id), true, "bookmarkbtn", "float-right")
	} else {
		starButton = createButton("star-outline", fmt.Sprintf("%dstar", n.Id), true, "bookmarkbtn", "float-right")
	}

	linkDiv.AppendChild(favicon)
	linkDiv.AppendChild(link)
	for _, t := range n.Tags {
		linkDiv.AppendChild(createSmallTagNode(*t, n.Id))
	}

	buttonDiv.AppendChild(menuButton)
	buttonDiv.AppendChild(cutButton)
	buttonDiv.AppendChild(deleteButton)
	buttonDiv.AppendChild(starButton)
	buttonDiv.AppendChild(editButton)

	mainDiv.AppendChild(linkDiv)
	mainDiv.AppendChild(buttonDiv)
	topDiv.AppendChild(mainDiv)
	topDiv.AppendChild(actionDiv)

	return topDiv
}

// createFolderNode creates a folder HTML element
func createFolderNode(n types.Folder) (*dom.HTMLDivElement, *dom.HTMLDivElement) {

	mainDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	mainDiv.SetClass("card")
	mainDiv.SetID(fmt.Sprintf("%dfolderMainDiv", n.Id))

	actionDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	actionDiv.SetID(fmt.Sprintf("%dactionDiv", n.Id))

	titleDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	titleDiv.SetClass("card-header")
	titleDiv.SetID(fmt.Sprintf("heading%d", n.Id))

	childrenDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	childrenDiv.SetClass("collapse")
	childrenDiv.SetAttribute("aria-labelledby", fmt.Sprintf("heading%d", n.Id))

	var pid int
	if n.Parent == nil {
		pid = 0
	} else {
		pid = n.Parent.Id
	}

	childrenDiv.SetAttribute("data-parent", fmt.Sprintf("#collapse%d", pid))
	childrenDiv.SetID(fmt.Sprintf("collapse%d", n.Id))

	titleH := document.CreateElement("h5").(*dom.HTMLHeadingElement)
	titleH.SetClass("mb-0")

	titleB := document.CreateElement("button").(*dom.HTMLButtonElement)
	titleB.SetClass("btn btn-link")
	titleB.SetAttribute("data-toggle", "collapse")
	titleB.SetAttribute("data-target", fmt.Sprintf("#collapse%d", n.Id))
	titleB.SetInnerHTML(n.Title)
	titleB.SetID(fmt.Sprintf("%dfolderLink", n.Id))

	countS := document.CreateElement("span").(*dom.HTMLSpanElement)
	countS.SetClass("count badge badge-primary badge-pill")
	countS.SetInnerHTML(fmt.Sprintf("%d", len(n.Bookmarks)))

	titleH.AppendChild(titleB)
	titleH.AppendChild(countS)

	childrenBody := document.CreateElement("div").(*dom.HTMLDivElement)
	childrenBody.SetClass("card-body")
	childrenBody.SetID(fmt.Sprintf("%dfolderBody", n.Id))

	menuButton := createButton("menu", fmt.Sprintf("%dmenu", n.Id), false, "folderbtn", "float-right")
	cutButton := createButton("content-cut", fmt.Sprintf("%dcut", n.Id), true, "folderbtn", "float-right")
	deleteButton := createButton("delete-outline", fmt.Sprintf("%ddelete", n.Id), true, "folderbtn", "float-right")
	pasteButton := createButton("content-paste", fmt.Sprintf("%dpaste", n.Id), true, "folderbtn", "float-right")
	addFolderButton := createButton("folder-plus-outline", fmt.Sprintf("%daddFolder", n.Id), true, "folderbtn", "float-right")
	addBookmarkButton := createButton("bookmark-plus-outline", fmt.Sprintf("%daddBookmark", n.Id), true, "folderbtn", "float-right")
	editButton := createButton("pencil-outline", fmt.Sprintf("%dedit", n.Id), true, "folderbtn", "float-right")

	titleDiv.AppendChild(titleH)
	titleDiv.AppendChild(menuButton)
	titleDiv.AppendChild(cutButton)
	titleDiv.AppendChild(deleteButton)
	titleDiv.AppendChild(pasteButton)
	titleDiv.AppendChild(addFolderButton)
	titleDiv.AppendChild(addBookmarkButton)
	titleDiv.AppendChild(editButton)

	childrenDiv.AppendChild(childrenBody)

	mainDiv.AppendChild(titleDiv)
	mainDiv.AppendChild(actionDiv)
	mainDiv.AppendChild(childrenDiv)

	return mainDiv, childrenBody
}

// createButton creates a button with a materialdesign icon
// - icon is the materialdesign icon name without the heading mdi-
// - id is the button id
// - hideen hides the button
// - classes are the additionnal classes for the button
func createButton(icon string, id string, hidden bool, classes ...string) *dom.HTMLButtonElement {
	b := document.CreateElement("button").(*dom.HTMLButtonElement)
	b.SetAttribute("type", "button")
	b.SetAttribute("data-role", "none")
	b.SetClass(icon + " btn btn-outline-dark bg-light")
	b.SetID(id)

	if hidden {
		b.Class().Add("invisible")
	}

	for _, c := range classes {
		b.Class().Add(c)
	}

	blabel := document.CreateElement("span").(*dom.HTMLSpanElement)
	blabel.SetID(id + "span")
	blabel.SetClass("mdi mdi-" + icon)

	b.AppendChild(blabel)

	return b
}

// createAddFolderForm returns a folder creation form
// with id = id+"createFolder" for the main div
// and id = id+"createFolderSubmit" for the submit button
func createAddFolderForm(id string) *dom.HTMLDivElement {
	dr := document.CreateElement("div").(*dom.HTMLDivElement)
	dr.SetID(id + "createFolder")
	dr.SetClass("row addFolder mt-2 mb-2 ml-5 mr-5")
	dc1 := document.CreateElement("div").(*dom.HTMLDivElement)
	dc1.SetClass("col col-11")
	dc2 := document.CreateElement("div").(*dom.HTMLDivElement)
	dc2.SetClass("col col-1")

	ifoldername := document.CreateElement("input").(*dom.HTMLInputElement)
	ifoldername.SetID(id + "createFolderInput")
	ifoldername.SetAttribute("type", "text")
	ifoldername.SetAttribute("placeholder", "name")
	ifoldername.SetClass("form-control")
	dc1.AppendChild(ifoldername)

	submit := createButton("check-bold", id+"createFolderSubmit", false, "float-left")
	dc2.AppendChild(submit)

	dr.AppendChild(dc1)
	dr.AppendChild(dc2)

	return dr
}

// createUpdateFolderForm returns a folder update form
// with id = id+"updateFolder" for the main div
// and id = id+"updateFolderSubmit" for the submit button
func createUpdateFolderForm(id string) *dom.HTMLDivElement {
	dr := document.CreateElement("div").(*dom.HTMLDivElement)
	dr.SetID(id + "updateFolder")
	dr.SetClass("row updateFolder mt-2 mb-2 ml-5 mr-5")
	dc1 := document.CreateElement("div").(*dom.HTMLDivElement)
	dc1.SetClass("col col-11")
	dc2 := document.CreateElement("div").(*dom.HTMLDivElement)
	dc2.SetClass("col col-1")

	ifoldername := document.CreateElement("input").(*dom.HTMLInputElement)
	ifoldername.SetID(id + "updateFolderInput")
	ifoldername.SetAttribute("type", "text")
	ifoldername.SetAttribute("placeholder", "name")
	ifoldername.SetClass("form-control")
	dc1.AppendChild(ifoldername)

	submit := createButton("check-bold", id+"updateFolderSubmit", false, "float-left")
	dc2.AppendChild(submit)

	dr.AppendChild(dc1)
	dr.AppendChild(dc2)

	return dr
}

// createAddFolderForm returns a bookmark creation form
// with id = id+"createBookmark" for the main div
// and id = id+"createBookmarkSubmit" for the submit button
func createAddBookmarkForm(id string) *dom.HTMLDivElement {
	dr := document.CreateElement("div").(*dom.HTMLDivElement)
	dr.SetID(id + "createBookmark")
	dr.SetClass("row addBookmark mt-2 mb-2 ml-5 mr-5")
	dc1 := document.CreateElement("div").(*dom.HTMLDivElement)
	dc1.SetClass("col col-12")
	dc2 := document.CreateElement("div").(*dom.HTMLDivElement)
	dc2.SetClass("col col-12")
	dc3 := document.CreateElement("div").(*dom.HTMLDivElement)
	dc3.SetClass("col col-12")
	dc4 := document.CreateElement("div").(*dom.HTMLDivElement)
	dc4.SetClass("col col-1")

	ibookmarkname := document.CreateElement("input").(*dom.HTMLInputElement)
	ibookmarkname.SetAttribute("type", "text")
	ibookmarkname.SetAttribute("placeholder", "name")
	ibookmarkname.SetClass("form-control")
	ibookmarkname.SetID(id + "createBookmarkInputName")
	dc1.AppendChild(ibookmarkname)

	ibookmarkurl := document.CreateElement("input").(*dom.HTMLInputElement)
	ibookmarkurl.SetAttribute("type", "text")
	ibookmarkurl.SetAttribute("placeholder", "URL")
	ibookmarkurl.SetClass("form-control")
	ibookmarkurl.SetID(id + "createBookmarkInputUrl")
	dc2.AppendChild(ibookmarkurl)

	ibookmarktags := document.CreateElement("select").(*dom.HTMLSelectElement)
	ibookmarktags.SetAttribute("placeholder", "tags")
	ibookmarktags.SetAttribute("multiple", "multiple")
	ibookmarktags.SetClass("form-control")
	ibookmarktags.SetID(id + "createBookmarkInputTags")
	dc3.AppendChild(ibookmarktags)

	submit := createButton("check-bold", id+"createBookmarkSubmit", false, "float-left")
	dc4.AppendChild(submit)

	dr.AppendChild(dc1)
	dr.AppendChild(dc2)
	dr.AppendChild(dc3)
	dr.AppendChild(dc4)

	return dr
}

// createUpdateFolderForm returns a bookmark update form
// with id = id+"updateBookmark" for the main div
// and id = id+"updateBookmarkSubmit" for the submit button
func createUpdateBookmarkForm(id string) *dom.HTMLDivElement {
	dr := document.CreateElement("div").(*dom.HTMLDivElement)
	dr.SetID(id + "updateBookmark")
	dr.SetClass("row updateBookmark mt-2 mb-2 ml-5 mr-5")
	dc1 := document.CreateElement("div").(*dom.HTMLDivElement)
	dc1.SetClass("col col-12")
	dc2 := document.CreateElement("div").(*dom.HTMLDivElement)
	dc2.SetClass("col col-12")
	dc3 := document.CreateElement("div").(*dom.HTMLDivElement)
	dc3.SetClass("col col-12")
	dc4 := document.CreateElement("div").(*dom.HTMLDivElement)
	dc4.SetClass("col col-1")

	ibookmarkname := document.CreateElement("input").(*dom.HTMLInputElement)
	ibookmarkname.SetAttribute("type", "text")
	ibookmarkname.SetAttribute("placeholder", "name")
	ibookmarkname.SetClass("form-control")
	ibookmarkname.SetID(id + "updateBookmarkInputName")
	dc1.AppendChild(ibookmarkname)

	ibookmarkurl := document.CreateElement("input").(*dom.HTMLInputElement)
	ibookmarkurl.SetAttribute("type", "text")
	ibookmarkurl.SetAttribute("placeholder", "URL")
	ibookmarkurl.SetClass("form-control")
	ibookmarkurl.SetID(id + "updateBookmarkInputUrl")
	dc2.AppendChild(ibookmarkurl)

	ibookmarktags := document.CreateElement("select").(*dom.HTMLSelectElement)
	ibookmarktags.SetAttribute("placeholder", "tags")
	ibookmarktags.SetAttribute("multiple", "multiple")
	ibookmarktags.SetClass("form-control")
	ibookmarktags.SetID(id + "updateBookmarkInputTags")
	dc3.AppendChild(ibookmarktags)

	submit := createButton("check-bold", id+"updateBookmarkSubmit", false, "float-left")
	dc4.AppendChild(submit)

	dr.AppendChild(dc1)
	dr.AppendChild(dc2)
	dr.AppendChild(dc3)
	dr.AppendChild(dc4)

	return dr
}

// hideActionButtons hide all the actions buttons
// except for item id
func hideActionButtons(id string) {
	jQuery(".content-cut").Not("#" + id + "cut").AddClass("invisible")
	jQuery(".delete-outline").Not("#" + id + "delete").AddClass("invisible")
	jQuery(".content-paste").Not("#" + id + "paste").AddClass("invisible")
	jQuery(".star-outline").Not("#" + id + "star").AddClass("invisible")
	jQuery(".star").Not("#" + id + "star").AddClass("invisible")
	jQuery(".folder-plus-outline").Not("#" + id + "addFolder").AddClass("invisible")
	jQuery(".bookmark-plus-outline").Not("#" + id + "addBookmark").AddClass("invisible")
	jQuery(".pencil-outline").Not("#" + id + "edit").AddClass("invisible")

	jQuery("#searchResults").SetHtml("")
	jQuery("input#searchInput").SetVal("")

	for _, e := range document.GetElementsByClassName("card") {
		e.SetAttribute("style", "border: 1px solid rgba(0,0,0,.125)")
	}
	for _, e := range document.GetElementsByClassName("bookmark") {
		e.SetAttribute("style", "border: none")
	}
}

// hideForms hide all the create/update forms
func hideForms() {
	jQuery(".addFolder").Remove()
	jQuery(".addBookmark").Remove()
	jQuery(".updateFolder").Remove()
	jQuery(".updateBookmark").Remove()
}

// resetConfirmButtons reset all delete confirm buttons
func resetConfirmButtons() {
	jQuery("span.mdi-check").AddClass("mdi-delete-outline")
	jQuery("span.mdi-check").RemoveClass("mdi-check")
}

// bindButtonEvents binds all the events
// for the folder or bookmark (isbookmark = true)
// with the given id
func bindButtonEvents(id string, isBookmark bool) {

	idInt, _ := strconv.Atoi(id)

	//
	// folder click event binding
	//
	jQuery("#"+id+"folderLink").On("click", func(e jquery.Event) {
		hideActionButtons(id)
		hideForms()
		resetConfirmButtons()
	})
	jQuery("#"+id+"bookmarkMainDiv").On("click", func(e jquery.Event) {
		hideActionButtons(id)
		resetConfirmButtons()
	})
	jQuery("#"+id+"folderMainDiv").On("click", func(e jquery.Event) {
		hideActionButtons(id)
		resetConfirmButtons()
	})
	jQuery("#"+id+"folderMainDiv").On("customdrop", func(e jquery.Event) {
		document.GetElementByID(id+"folderMainDiv").SetAttribute("style", "border: 1px solid rgba(0,0,0,.125)")

		e.StopPropagation()

		bkmurl := jQuery("input[name=droppedurl]").Val()
		// TODO deal with errors
		fldid, _ := strconv.Atoi(id)
		go createBookmark(types.Bookmark{Title: bkmurl, URL: bkmurl, Folder: &types.Folder{Id: fldid}})
	})
	jQuery("#"+id+"folderMainDiv").On("dragenter", func(e jquery.Event) {
		e.StopPropagation()
		for _, e := range document.GetElementsByClassName("card") {
			e.SetAttribute("style", "border: 1px solid rgba(0,0,0,.125)")
		}
		document.GetElementByID(id+"folderMainDiv").SetAttribute("style", "border: 5px solid yellow;")
	})

	//
	// buttons event binding
	//
	// menu
	jQuery("#"+id+"menu").On("click", func(e jquery.Event) {
		e.StopPropagation()

		// make all other buttons and forms invisible
		hideActionButtons(id)
		hideForms()
		resetConfirmButtons()

		if jQuery(fmt.Sprintf("#%scut", id)).HasClass("invisible") {
			fmt.Println("show")

			jQuery(fmt.Sprintf("#%scut", id)).RemoveClass("invisible")
			jQuery("#" + id + "delete").RemoveClass("invisible")
			jQuery("#" + id + "star").RemoveClass("invisible")
			jQuery("#" + id + "addFolder").RemoveClass("invisible")
			jQuery("#" + id + "addBookmark").RemoveClass("invisible")
			jQuery("#" + id + "edit").RemoveClass("invisible")
		} else {
			fmt.Println("hide")

			jQuery(fmt.Sprintf("#%scut", id)).AddClass("invisible")
			jQuery("#" + id + "delete").AddClass("invisible")
			jQuery("#" + id + "star").AddClass("invisible")
			jQuery("#" + id + "addFolder").AddClass("invisible")
			jQuery("#" + id + "addBookmark").AddClass("invisible")
			jQuery("#" + id + "edit").AddClass("invisible")
		}

		// something to paste ?
		if jQuery("input[type=hidden][name=cutednodeid]").Val() != "" {
			jQuery("#" + id + "paste").RemoveClass("invisible")
		}
	})

	// cut
	jQuery("#"+id+"cut").On("click", func(e jquery.Event) {
		e.StopPropagation()

		jQuery("input[type=hidden][name=cutednodeid]").SetVal(id)
		hideActionButtons(id)
	})

	// delete
	jQuery("#"+id+"delete").On("click", func(e jquery.Event) {
		e.StopPropagation()

		if jQuery("#" + id + "delete > span").HasClass("mdi-check") {
			if strings.Index(id, "-") == -1 {
				go deleteFolder(id)
			} else {
				go deleteBookmark(id)
			}
		} else {
			jQuery("#" + id + "delete > span").RemoveClass("mdi-delete-outline")
			jQuery("#" + id + "delete > span").AddClass("mdi-check")
		}

	})

	// edit
	jQuery("#"+id+"edit").On("click", func(e jquery.Event) {
		e.StopPropagation()

		hideActionButtons(id)
		hideForms()
		resetConfirmButtons()

		if strings.Index(id, "-") == -1 {

			f := createUpdateFolderForm(id)
			jQuery("#" + id + "actionDiv").Append(f)

			// init input with current folder name
			jQuery("#" + id + "updateFolderInput").SetVal(jQuery("button#" + id + "folderLink").Html())
			jQuery("#" + id + "updateFolderInput").Focus()
			jQuery("#" + id + "updateFolderInput").Select()

			// add event binding
			jQuery("#"+id+"updateFolderSubmit").On("click", func(e jquery.Event) {
				e.StopPropagation()

				folderName := jQuery("#" + id + "updateFolderInput").Val()
				fldid, _ := strconv.Atoi(id)
				go updateFolder(types.Folder{Title: folderName, Id: fldid})
			})
		} else {

			b := createUpdateBookmarkForm(id)
			jQuery("#" + id + "actionDiv").Append(b)

			// init inputs with current bookmark
			jQuery("#" + id + "updateBookmarkInputName").SetVal(jQuery("a#" + id + "bookmarkLink").Html())
			jQuery("#" + id + "updateBookmarkInputUrl").SetVal(jQuery("a#" + id + "bookmarkLink").Attr("href"))
			jQuery("#" + id + "updateBookmarkInputName").Focus()
			jQuery("#" + id + "updateBookmarkInputName").Select()

			tagsElements := document.GetElementsByClassName(id + "badge")
			for _, t := range tagsElements {
				value := t.(*dom.HTMLSpanElement).GetAttribute("value")
				text := t.(*dom.HTMLSpanElement).InnerHTML()

				o := document.CreateElement("option").(*dom.HTMLOptionElement)
				o.SetAttribute("value", value)
				o.Selected = true
				o.SetInnerHTML(text)
				jQuery("#"+id+"updateBookmarkInputTags").Append(o).Call("trigger", "change")
			}

			// select2ify the form
			//js.Global.Call("select2ify", id)
			select2ify(id)

			// add event binding
			jQuery("#"+id+"updateBookmarkSubmit").On("click", func(e jquery.Event) {
				e.StopPropagation()

				bookmarkName := jQuery("#" + id + "updateBookmarkInputName").Val()
				bookmarkUrl := jQuery("#" + id + "updateBookmarkInputUrl").Val()
				bookmarkTags := jQuery("#"+id+"updateBookmarkInputTags").Call("select2", "data")

				b := types.Bookmark{Id: idInt, Title: bookmarkName, URL: bookmarkUrl}

				for _, t := range bookmarkTags.ToArray() {
					text := t.(map[string]interface{})["text"].(string)
					id := t.(map[string]interface{})["id"].(string)

					var tid int
					if text == id {
						tid = -1
					} else {
						tid, _ = strconv.Atoi(id)
					}
					b.Tags = append(b.Tags, &types.Tag{Id: tid, Name: text})
				}

				go updateBookmark(b)
			})
		}
	})

	// paste, addFolder, addBookmark
	if !isBookmark {
		jQuery("#"+id+"paste").On("click", func(e jquery.Event) {
			e.StopPropagation()

			// getting the cutted folder id
			cuttedid := jQuery("input[type=hidden][name=cutednodeid]").Val()
			cuttedidInt, _ := strconv.Atoi(cuttedid)

			if strings.Index(cuttedid, "-") == -1 {
				// passing the folder with its new parent to the moveFolder function
				// we do not need the folder name
				go moveFolder(types.Folder{Id: cuttedidInt, Parent: &types.Folder{Id: idInt}})
			} else {
				// passing the folder with its new parent to the moveFolder function
				// we do not need the folder name
				go moveBookmark(types.Bookmark{Id: cuttedidInt, Folder: &types.Folder{Id: idInt}})
			}

			hideActionButtons(id)
		})

		jQuery("#"+id+"addFolder").On("click", func(e jquery.Event) {
			e.StopPropagation()

			hideActionButtons(id)
			hideForms()
			resetConfirmButtons()

			f := createAddFolderForm(id)
			jQuery("#" + id + "actionDiv").Append(f)
			jQuery("#" + id + "createFolderInput").Focus()

			// add event binding
			jQuery("#"+id+"createFolderSubmit").On("click", func(e jquery.Event) {
				e.StopPropagation()

				folderName := jQuery("#" + id + "createFolderInput").Val()

				go createFolder(types.Folder{Title: folderName, Parent: &types.Folder{Id: idInt}})
			})
		})

		jQuery("#"+id+"addBookmark").On("click", func(e jquery.Event) {
			e.StopPropagation()

			hideActionButtons(id)
			hideForms()
			resetConfirmButtons()

			jQuery("#" + id + "actionDiv").Append(createAddBookmarkForm(id))
			jQuery("#" + id + "createBookmarkInputName").Focus()

			// select2ify the form
			//js.Global.Call("select2ify", id)
			select2ify(id)

			// add event binding
			jQuery("#"+id+"createBookmarkSubmit").On("click", func(e jquery.Event) {
				e.StopPropagation()

				b := types.Bookmark{}
				b.Title = jQuery("#" + id + "createBookmarkInputName").Val()
				b.URL = jQuery("#" + id + "createBookmarkInputUrl").Val()
				b.Folder = &types.Folder{Id: idInt}

				ts := make([]*types.Tag, 0)

				tags := jQuery("#"+id+"createBookmarkInputTags").Call("select2", "data")

				tags.Each(func(i int, data interface{}) {
					// TODO: deal with errors
					id, _ := strconv.Atoi(data.(map[string]interface{})["id"].(string))
					text := data.(map[string]interface{})["text"].(string)
					ts = append(ts, &types.Tag{
						Id:   id,
						Name: text,
					})
				})
				b.Tags = ts

				go createBookmark(b)

			})
		})
	}

	// star
	if isBookmark {
		jQuery("#"+id+"star").On("click", func(e jquery.Event) {

			e.StopPropagation()

			if jQuery("span#" + id + "starspan").HasClass("mdi-star") {
				go unstarBookmark(id)
			} else {
				go starBookmark(id)
			}

		})
	}
}

// displayNode recursively display a Node as an JQM listview widget
func displayNode(n interface{}, e *dom.HTMLDivElement) {
	t := reflect.TypeOf(n)
	isBookmark := false
	id := "0"

	switch t {
	case reflect.TypeOf(types.Bookmark{}):
		b := n.(types.Bookmark)
		if b.Starred {
			document.GetElementByID("star").AppendChild(createStarredBookmarkNode(b))
		}
		e.AppendChild(createBookmarkNode(b))

		isBookmark = true
		id = fmt.Sprintf("-%d", b.Id)
	case reflect.TypeOf(types.Folder{}):
		f := n.(types.Folder)
		b, u := createFolderNode(f)
		e.AppendChild(b)
		for _, c := range f.Folders {
			displayNode(*c, u)
		}
		for _, c := range f.Bookmarks {
			displayNode(*c, u)
		}

		id = fmt.Sprintf("%d", f.Id)
	default:
		fmt.Println("unexpected type " + t.String())
	}

	bindButtonEvents(id, isBookmark)
}

func getTags() {

	var (
		data []byte
		err  error
		tags []types.Tag
	)

	if data, err = xhr.Send("GET", "/getTags/", nil); err != nil {
		fmt.Println("error getting the tags")
	}

	if err = json.NewDecoder(strings.NewReader(string(data))).Decode(&tags); err != nil {
		fmt.Println("error decoding the JSON")
	}

	for _, t := range tags {
		// adding tag to the tags div
		id := fmt.Sprintf("%d", t.Id)
		document.GetElementByID("tags").AppendChild(createTagNode(id, t.Name))
	}

}

func getNodes() {

	var (
		data     []byte
		err      error
		rootNode types.Folder
	)

	if data, err = xhr.Send("GET", "/getTree/", nil); err != nil {
		fmt.Println("error getting the bookmarks")
	}

	if err = json.NewDecoder(strings.NewReader(string(data))).Decode(&rootNode); err != nil {
		fmt.Println("error decoding the JSON")
	}

	displayNode(rootNode, rootDiv)

}

func main() {

	// when page is loaded...
	document.AddEventListener("DOMContentLoaded", false, func(event dom.Event) {
		event.PreventDefault()

		// getting tree div
		rootDiv = document.GetElementByID("collapse1").(*dom.HTMLDivElement)

		// get remote folders and bookmarks
		go getNodes()

		// get remote tags
		go getTags()

		// add search form listener
		var t *time.Timer
		t = time.NewTimer(0)
		document.GetElementByID("searchInput").AddEventListener("keypress", false, func(event dom.Event) {
			t.Stop()
			t = time.AfterFunc(1000000000, func() {
				if jQuery("input#searchInput").Val() != "" {
					go searchBookmark(jQuery("input#searchInput").Val())
				}
			})
		})
	})

	// exporting functions to be called from other JS files
	js.Global.Set("global", map[string]interface{}{})

}
