package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

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
	jQuery   = jquery.NewJQuery
)

func init() {
	window = dom.GetWindow()
	document = window.Document()
}

// getBranchNodes remotely gets the children nodes of the
// node with id "parentId"
func getBranchNodes(parentId string) []types.Node {
	var (
		data  []byte
		err   error
		nodes []types.Node
	)

	if data, err = xhr.Send("GET", "/getBranchNodes/?parentId="+parentId, nil); err != nil {
		errors.New("error getting nodes of  " + parentId)
	}

	// decoding response
	if err = json.NewDecoder(strings.NewReader(string(data))).Decode(&nodes); err != nil {
		fmt.Println("error decoding the JSON")
	}

	return nodes
}

// starBookmark star the bookmark with the
// given id
func starBookmark(id string) error {
	var (
		err  error
		data []byte
		b    types.Bookmark
	)

	if data, err = xhr.Send("GET", "/starBookmark/?star=true&bookmarkId="+id, nil); err != nil {
		return errors.New("error starring bookmark " + id)
	}

	datar := bytes.NewReader(data)
	decoder := json.NewDecoder(datar)

	if err = decoder.Decode(&b); err != nil {
		return errors.New("error decoding reponse from star " + id)
	}

	// adding bookmark to the star div
	document.GetElementByID("star").AppendChild(createStarredBookmarkNode(id, b.Title, b.URL, b.Favicon))

	// changing bookmark star icon
	jQuery(fmt.Sprintf("span#%sstarspan", id)).RemoveClass("mdi-star-outline")
	jQuery(fmt.Sprintf("span#%sstarspan", id)).AddClass("mdi-star")

	return nil
}

// unstarBookmark star the bookmark with the
// given id
func unstarBookmark(id string) error {
	var (
		err  error
		data []byte
		b    types.Bookmark
	)

	if data, err = xhr.Send("GET", "/starBookmark/?star=false&bookmarkId="+id, nil); err != nil {
		return errors.New("error unstarring bookmark " + id)
	}

	datar := bytes.NewReader(data)
	decoder := json.NewDecoder(datar)

	if err = decoder.Decode(&b); err != nil {
		return errors.New("error decoding reponse from star " + id)
	}

	// removing bookmark from the star div
	jQuery(fmt.Sprintf("button#%sstarred", id)).Remove()

	// changing bookmark star icon
	jQuery(fmt.Sprintf("span#%sstarspan", id)).RemoveClass("mdi-star")
	jQuery(fmt.Sprintf("span#%sstarspan", id)).AddClass("mdi-star-outline")

	return nil
}

// createBookmark remotely creates the bookmark b
func createBookmark(b types.Bookmark) error {
	var (
		err     error
		payload []byte
	)

	// TODO: deal with errors
	payload, _ = json.Marshal(b)

	if _, err = xhr.Send("POST", "/addBookmark/", payload); err != nil {
		return errors.New("error creating bookmark " + b.Title)
	}

	// removing create form
	jQuery(fmt.Sprintf("#%dcreateBookmark", b.Folder.Id)).Remove()

	// getting and cleaning the bookmark parent directory div
	parentD := document.GetElementByID(fmt.Sprintf("%dfolderBody", b.Folder.Id)).(*dom.HTMLDivElement)
	parentD.SetInnerHTML("")

	// lazily getting the bookmark parent folder children nodes
	// and refreshing them
	cnodes := getBranchNodes(fmt.Sprintf("%d", b.Folder.Id))
	for _, n := range cnodes {
		nid := fmt.Sprintf("%d", n.Key)
		displayNode(n, parentD, nid)
	}

	return nil
}

// moveBookmark remotely moves the bookmark b
func moveBookmark(b types.Bookmark) error {
	var (
		err     error
		payload []byte
	)

	// TODO: deal with errors
	payload, _ = json.Marshal(b)

	if _, err = xhr.Send("PUT", "/moveBookmark/", payload); err != nil {
		return errors.New("error moving bookmark " + b.Title)
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
	for _, n := range cnodes {
		nid := fmt.Sprintf("%d", n.Key)
		displayNode(n, parentD, nid)
	}

	// resetting the cutednodeid
	jQuery("input[type=hidden][name=cutednodeid]").SetVal("")

	return nil
}

// moveFolder remotely moves the folder f
func moveFolder(f types.Folder) error {
	var (
		err     error
		payload []byte
	)

	// TODO: deal with errors
	payload, _ = json.Marshal(f)

	if _, err = xhr.Send("PUT", "/moveFolder/", payload); err != nil {
		return errors.New("error moving folder " + f.Title)
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
	for _, n := range cnodes {
		nid := fmt.Sprintf("%d", n.Key)
		displayNode(n, parentD, nid)
	}

	// resetting the cutednodeid
	jQuery("input[type=hidden][name=cutednodeid]").SetVal("")

	return nil
}

// createFolder remotely creates the folder f
func createFolder(f types.Folder) error {
	var (
		err     error
		payload []byte
	)

	// TODO: deal with errors
	payload, _ = json.Marshal(f)

	if _, err = xhr.Send("POST", "/addFolder/", payload); err != nil {
		return errors.New("error creating folder " + f.Title)
	}

	// removing create form
	jQuery(fmt.Sprintf("#%dcreateFolder", f.Parent.Id)).Remove()

	// getting and cleaning the folder parent directory div
	parentD := document.GetElementByID(fmt.Sprintf("%dfolderBody", f.Parent.Id)).(*dom.HTMLDivElement)
	parentD.SetInnerHTML("")

	// lazily getting the folder parent folder children nodes
	// and refreshing them
	cnodes := getBranchNodes(fmt.Sprintf("%d", f.Parent.Id))
	for _, n := range cnodes {
		nid := fmt.Sprintf("%d", n.Key)
		displayNode(n, parentD, nid)
	}

	return nil
}

// deleteFolder remotely deletes the folder
// with id "itemId"
func deleteFolder(itemId string) error {
	var (
		err error
	)

	if _, err = xhr.Send("DELETE", "/deleteFolder/?itemId="+itemId, nil); err != nil {
		return errors.New("error deleting folder " + itemId)
	}

	jQuery("div#" + itemId + "folderMainDiv").Remove()

	return nil
}

// deleteBookmark remotely deletes the bookmark
// with id "itemId"
func deleteBookmark(itemId string) error {
	var (
		err error
	)

	if _, err = xhr.Send("DELETE", "/deleteBookmark/?itemId="+itemId, nil); err != nil {
		return errors.New("error deleting bookmark " + itemId)
	}

	jQuery("div#" + itemId + "bookmarkMainDiv").Remove()

	return nil
}

// createStarredBookmarkNode creates a starred bookmark HTML element
func createStarredBookmarkNode(id, title, URL, icon string) *dom.HTMLButtonElement {

	buttonDiv := document.CreateElement("button").(*dom.HTMLButtonElement)
	buttonDiv.SetID(id + "starred")
	buttonDiv.SetClass("btn btn-outline-dark")
	buttonDiv.SetInnerHTML(title)
	buttonDiv.AddEventListener("click", false, func(event dom.Event) {
		window.Open(URL, "", "")
	})

	// favicon := document.CreateElement("img").(*dom.HTMLImageElement)
	// favicon.SetClass("favicon")
	// favicon.SetAttribute("src", icon)

	//buttonDiv.AppendChild(favicon)

	return buttonDiv

}

// createBookmarkNode creates a bookmark HTML element
func createBookmarkNode(id, title, URL, icon string, starred bool) *dom.HTMLDivElement {

	mainDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	mainDiv.SetClass("row bookmark")
	mainDiv.SetID(id + "bookmarkMainDiv")

	linkDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	linkDiv.SetClass("col col-10")

	buttonDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	buttonDiv.SetClass("col col-2")

	link := document.CreateElement("a").(*dom.HTMLAnchorElement)
	link.SetAttribute("href", URL)
	link.SetAttribute("target", "_blank")
	link.SetID(id + "bookmarkLink")
	link.SetInnerHTML(title)

	favicon := document.CreateElement("img").(*dom.HTMLImageElement)
	favicon.SetClass("favicon")
	favicon.SetAttribute("src", icon)

	menuButton := createButton("menu", id+"menu", "visible", "bookmarkbtn", "float-right")
	cutButton := createButton("content-cut", id+"cut", "invisible", "bookmarkbtn", "float-right")
	deleteButton := createButton("delete-outline", id+"delete", "invisible", "bookmarkbtn", "float-right")
	var starButton *dom.HTMLButtonElement
	if starred {
		starButton = createButton("star", id+"star", "invisible", "bookmarkbtn", "float-right")
	} else {
		starButton = createButton("star-outline", id+"star", "invisible", "bookmarkbtn", "float-right")
	}

	linkDiv.AppendChild(favicon)
	linkDiv.AppendChild(link)

	buttonDiv.AppendChild(menuButton)
	buttonDiv.AppendChild(cutButton)
	buttonDiv.AppendChild(deleteButton)
	buttonDiv.AppendChild(starButton)

	mainDiv.AppendChild(linkDiv)
	mainDiv.AppendChild(buttonDiv)

	return mainDiv
}

// createFolderNode creates a folder HTML element
func createFolderNode(id, parentid, title, count string) (*dom.HTMLDivElement, *dom.HTMLDivElement) {

	mainDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	mainDiv.SetClass("card")
	mainDiv.SetID(id + "folderMainDiv")

	// mainDiv.SetAttribute("ondrop", "dzDrop(event)")
	// mainDiv.SetAttribute("ondragover", "dzAllowDrop(event)")

	actionDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	actionDiv.SetID(id + "actionDiv")

	titleDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	titleDiv.SetClass("card-header")
	titleDiv.SetID("heading" + id)

	childrenDiv := document.CreateElement("div").(*dom.HTMLDivElement)
	childrenDiv.SetClass("collapse")
	childrenDiv.SetAttribute("aria-labelledby", "heading"+id)
	childrenDiv.SetAttribute("data-parent", "#collapse"+parentid)
	childrenDiv.SetID("collapse" + id)

	titleH := document.CreateElement("h5").(*dom.HTMLHeadingElement)
	titleH.SetClass("mb-0")

	titleB := document.CreateElement("button").(*dom.HTMLButtonElement)
	titleB.SetClass("btn btn-link")
	titleB.SetAttribute("data-toggle", "collapse")
	titleB.SetAttribute("data-target", "#collapse"+id)
	titleB.SetInnerHTML(title + " (" + count + ")")
	titleB.SetID(id + "folderLink")

	titleH.AppendChild(titleB)

	childrenBody := document.CreateElement("div").(*dom.HTMLDivElement)
	childrenBody.SetClass("card-body")
	childrenBody.SetID(id + "folderBody")

	menuButton := createButton("menu", id+"menu", "visible", "folderbtn", "float-right")
	cutButton := createButton("content-cut", id+"cut", "invisible", "folderbtn", "float-right")
	deleteButton := createButton("delete-outline", id+"delete", "invisible", "folderbtn", "float-right")
	pasteButton := createButton("content-paste", id+"paste", "invisible", "folderbtn", "float-right")
	addFolderButton := createButton("folder-plus-outline", id+"addFolder", "invisible", "folderbtn", "float-right")
	addBookmarkButton := createButton("bookmark-plus-outline", id+"addBookmark", "invisible", "folderbtn", "float-right")

	titleDiv.AppendChild(titleH)
	titleDiv.AppendChild(menuButton)
	titleDiv.AppendChild(cutButton)
	titleDiv.AppendChild(deleteButton)
	titleDiv.AppendChild(pasteButton)
	titleDiv.AppendChild(addFolderButton)
	titleDiv.AppendChild(addBookmarkButton)

	childrenDiv.AppendChild(childrenBody)

	mainDiv.AppendChild(titleDiv)
	mainDiv.AppendChild(actionDiv)
	mainDiv.AppendChild(childrenDiv)

	return mainDiv, childrenBody
}

// createButton creates a button with a materialdesign icon
// - icon is the materialdesign icon name without the heading mdi-
// - id is the button id
// - visibility = visible | invisible
// - classes are the additionnal classes for the button
func createButton(icon string, id string, visibility string, classes ...string) *dom.HTMLButtonElement {
	b := document.CreateElement("button").(*dom.HTMLButtonElement)
	b.SetAttribute("type", "button")
	b.SetAttribute("data-role", "none")
	b.SetClass(icon + " btn btn-outline-dark bg-light " + visibility)
	b.SetID(id)

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
	ifoldername.SetAttribute("placeholder", "folder name")
	ifoldername.SetClass("form-control")
	dc1.AppendChild(ifoldername)

	submit := createButton("check", id+"createFolderSubmit", "visible", "float-left")
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
	ibookmarkname.SetAttribute("placeholder", "bookmark name")
	ibookmarkname.SetClass("form-control")
	ibookmarkname.SetID(id + "createBookmarkInputName")
	dc1.AppendChild(ibookmarkname)

	ibookmarkurl := document.CreateElement("input").(*dom.HTMLInputElement)
	ibookmarkurl.SetAttribute("type", "text")
	ibookmarkurl.SetAttribute("placeholder", "bookmark URL")
	ibookmarkurl.SetClass("form-control")
	ibookmarkurl.SetID(id + "createBookmarkInputUrl")
	dc2.AppendChild(ibookmarkurl)

	ibookmarktags := document.CreateElement("select").(*dom.HTMLSelectElement)
	ibookmarktags.SetAttribute("placeholder", "tags")
	ibookmarktags.SetAttribute("multiple", "multiple")
	ibookmarktags.SetClass("form-control")
	ibookmarktags.SetID(id + "createBookmarkInputTags")
	dc3.AppendChild(ibookmarktags)

	submit := createButton("check", id+"createBookmarkSubmit", "visible", "float-left")
	dc4.AppendChild(submit)

	dr.AppendChild(dc1)
	dr.AppendChild(dc2)
	dr.AppendChild(dc3)
	dr.AppendChild(dc4)

	return dr
}

// hideActionButtons hide all the actions buttons
// except menu
func hideActionButtons() {
	jQuery(".content-cut").AddClass("invisible")
	jQuery(".delete-outline").AddClass("invisible")
	jQuery(".content-paste").AddClass("invisible")
	jQuery(".star-outline").AddClass("invisible")
	jQuery(".star").AddClass("invisible")
	jQuery(".folder-plus-outline").AddClass("invisible")
	jQuery(".bookmark-plus-outline").AddClass("invisible")
}

// hideForms hide all the create/update forms
func hideForms() {
	jQuery(".addFolder").Remove()
	jQuery(".addBookmark").Remove()
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
		fmt.Println("clic on " + id)
		hideActionButtons()
		hideForms()
		resetConfirmButtons()
	})
	jQuery("#"+id+"bookmarkMainDiv").On("click", func(e jquery.Event) {
		fmt.Println("clic on " + id)
		hideActionButtons()
		hideForms()
		resetConfirmButtons()
	})
	jQuery("#"+id+"folderMainDiv").On("click", func(e jquery.Event) {
		fmt.Println("clic on " + id)
		hideActionButtons()
		hideForms()
		resetConfirmButtons()
	})
	jQuery("#"+id+"folderMainDiv").On("customdrop", func(e jquery.Event) {
		e.StopPropagation()
		fmt.Println("customdrop on " + id)

		bkmurl := jQuery("input[name=droppedurl]").Val()
		// TODO deal with errors
		fldid, _ := strconv.Atoi(id)
		go createBookmark(types.Bookmark{Title: bkmurl, URL: bkmurl, Folder: &types.Folder{Id: fldid}})
	})

	//
	// buttons event binding
	//
	// menu
	jQuery("#"+id+"menu").On("click", func(e jquery.Event) {
		e.StopPropagation()

		// make all other buttons and forms invisible
		hideActionButtons()
		hideForms()
		resetConfirmButtons()

		jQuery("#" + id + "cut").RemoveClass("invisible")
		jQuery("#" + id + "delete").RemoveClass("invisible")
		jQuery("#" + id + "star").RemoveClass("invisible")
		jQuery("#" + id + "addFolder").RemoveClass("invisible")
		jQuery("#" + id + "addBookmark").RemoveClass("invisible")

		// something to paste ?
		if jQuery("input[type=hidden][name=cutednodeid]").Val() != "" {
			jQuery("#" + id + "paste").RemoveClass("invisible")
		}
	})

	// cut
	jQuery("#"+id+"cut").On("click", func(e jquery.Event) {
		e.StopPropagation()

		fmt.Println("clicked cut " + id)
		jQuery("input[type=hidden][name=cutednodeid]").SetVal(id)
		hideActionButtons()
	})

	// delete
	jQuery("#"+id+"delete").On("click", func(e jquery.Event) {
		e.StopPropagation()
		fmt.Println("clicked delete " + id)

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

	// paste, addFolder, addBookmark
	if !isBookmark {
		jQuery("#"+id+"paste").On("click", func(e jquery.Event) {
			e.StopPropagation()
			fmt.Println("clicked link paste " + id)

			// getting the cutted folder id
			cuttedid := jQuery("input[type=hidden][name=cutednodeid]").Val()
			cuttedidInt, _ := strconv.Atoi(cuttedid)

			if strings.Index(cuttedid, "-") == -1 {
				// passing the folder with its new parent to the moveFolder function
				// we do not need the folder name
				fmt.Println("move folder")
				go moveFolder(types.Folder{Id: cuttedidInt, Parent: &types.Folder{Id: idInt}})
			} else {
				// passing the folder with its new parent to the moveFolder function
				// we do not need the folder name
				fmt.Println("move bookmark")
				go moveBookmark(types.Bookmark{Id: cuttedidInt, Folder: &types.Folder{Id: idInt}})
			}

			hideActionButtons()
		})

		jQuery("#"+id+"addFolder").On("click", func(e jquery.Event) {
			e.StopPropagation()
			fmt.Println("clicked link add folder " + id)

			hideActionButtons()
			hideForms()
			resetConfirmButtons()

			f := createAddFolderForm(id)
			jQuery("#" + id + "actionDiv").Append(f)

			// add event binding
			jQuery("#"+id+"createFolderSubmit").On("click", func(e jquery.Event) {
				e.StopPropagation()

				folderName := jQuery("#" + id + "createFolderInput").Val()

				fmt.Println("create folder " + folderName + " of " + id)

				go createFolder(types.Folder{Title: folderName, Parent: &types.Folder{Id: idInt}})
			})
		})

		jQuery("#"+id+"addBookmark").On("click", func(e jquery.Event) {
			e.StopPropagation()
			fmt.Println("clicked link add bookmark " + id)

			hideActionButtons()
			hideForms()
			resetConfirmButtons()

			jQuery("#" + id + "actionDiv").Append(createAddBookmarkForm(id))

			// select2ify the form
			js.Global.Call("select2ify", id)

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

				fmt.Println("create bookmark in " + id)
			})
		})
	}

	// star
	if isBookmark {
		jQuery("#"+id+"star").On("click", func(e jquery.Event) {

			e.StopPropagation()
			fmt.Println("clicked link star " + id)

			if jQuery("span#" + id + "starspan").HasClass("mdi-star") {
				go unstarBookmark(id)
			} else {
				go starBookmark(id)
			}

		})
	}
}

// displayNode recursively display a Node as an JQM listview widget
func displayNode(n types.Node, e *dom.HTMLDivElement, parentId string) {

	// Node keys are negative for bookmarks and positive for folders
	id := fmt.Sprintf("%d", n.Key)
	isBookmark := n.Key < 0

	// building starred bookmark list
	if n.Starred {
		document.GetElementByID("star").AppendChild(createStarredBookmarkNode(id, n.Title, n.URL, n.Icon))
	}

	switch isBookmark {
	case true:
		//
		// bookmark
		//
		f := createBookmarkNode(id, n.Title, n.URL, n.Icon, n.Starred)
		e.AppendChild(f)

	default:
		//
		// folder
		//
		// children count
		c := len(n.Children)
		b, u := createFolderNode(id, parentId, n.Title, fmt.Sprintf("%d", c))
		e.AppendChild(b)

		for _, c := range n.Children {
			displayNode(*c, u, id)
		}
	}

	bindButtonEvents(id, isBookmark)
}

func getNodes() {

	var (
		data  []byte
		err   error
		nodes []types.Node
	)

	if data, err = xhr.Send("GET", "/getTree/", nil); err != nil {
		fmt.Println("error getting the bookmarks")
	}

	if err = json.NewDecoder(strings.NewReader(string(data))).Decode(&nodes); err != nil {
		fmt.Println("error decoding the JSON")
	}

	for _, n := range nodes {
		displayNode(n, rootDiv, "1")
	}

	// https://stackoverflow.com/questions/6977338/jquery-mobile-listview-refresh
	jQuery("#tree").Trigger("create")

}

func main() {

	// when page is loaded...
	document.AddEventListener("DOMContentLoaded", false, func(event dom.Event) {
		event.PreventDefault()

		// getting tree div
		rootDiv = document.GetElementByID("collapse1").(*dom.HTMLDivElement)

		// get remote folders and bookmarks
		go getNodes()
	})

	// exporting functions to be called from other JS files
	js.Global.Set("global", map[string]interface{}{})

}
