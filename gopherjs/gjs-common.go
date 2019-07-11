package main

import (
	"encoding/json"
	"errors"
	"fmt"
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
	rootUL   *dom.HTMLUListElement
	jQuery   = jquery.NewJQuery
)

func init() {
	window = dom.GetWindow()
	document = window.Document()
}

// TODO: use types.Node instead
type folder struct {
	folderName string
	parentId   string
}

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

func createFolder(folderName string, parentId string) error {
	var (
		//data []byte
		err error
		//node types.Node
	)

	if _, err = xhr.Send("GET", "/addFolder/?folderName="+folderName+"&parentId="+parentId, nil); err != nil {
		return errors.New("error creating folder " + folderName)
	}

	// removing create form
	jQuery("#" + parentId + "createFolder").Remove()

	ul := document.GetElementByID(parentId + "ul").(*dom.HTMLUListElement)
	ul.SetInnerHTML("")

	// getting children nodes
	cnodes := getBranchNodes(parentId)
	for _, n := range cnodes {
		displayNode(n, ul)
	}

	// // decoding response to extract the new folder id
	// node = types.Node{}
	// if err = json.NewDecoder(strings.NewReader(string(data))).Decode(&node); err != nil {
	// 	fmt.Println("error decoding the JSON")
	// }
	// fid := fmt.Sprintf("%d", node.Key)

	// // creating folder node
	// f, _ := createFolderNode(fid, node.Title, "0")
	// // appending to the parent ul as the first child
	// jQuery("#" + parentId + "ul").Prepend(f)

	// // binding button events
	// bindButtonEvents(fid, false)

	// refreshing JQM
	//jQuery("#tree").Trigger("create")

	return nil
}

func deleteFolder(itemId string) error {
	var (
		err error
	)

	if _, err = xhr.Send("GET", "/deleteFolder/?itemId="+itemId, nil); err != nil {
		return errors.New("error deleting folder " + itemId)
	}

	return nil
}

func createBookmarkNode(id, title, URL, icon string) *dom.HTMLLIElement {
	li := document.CreateElement("li").(*dom.HTMLLIElement)
	li.SetAttribute("data-icon", "false")
	li.SetID(id)

	favicon := document.CreateElement("img").(*dom.HTMLImageElement)
	favicon.SetClass("ui-li-icon")
	favicon.SetAttribute("src", icon)

	mainSpan := document.CreateElement("span").(*dom.HTMLSpanElement)

	menuButton := createButton("menu", id+"menu", "visible", "float-right")
	cutButton := createButton("content-cut", id+"cut", "invisible", "float-right")
	deleteButton := createButton("delete-outline", id+"delete", "invisible", "float-right")
	starButton := createButton("star-outline", id+"delete", "invisible", "float-right")

	link := document.CreateElement("a").(*dom.HTMLAnchorElement)
	link.SetAttribute("href", URL)
	link.SetAttribute("target", "_blank")
	link.SetID(id + "link")
	link.SetInnerHTML(title)

	mainSpan.AppendChild(link)
	mainSpan.AppendChild(menuButton)
	mainSpan.AppendChild(cutButton)
	mainSpan.AppendChild(deleteButton)
	mainSpan.AppendChild(starButton)

	li.AppendChild(favicon)
	li.AppendChild(mainSpan)

	return li
}

func createFolderNode(id, title, count string) (*dom.HTMLLIElement, *dom.HTMLUListElement) {

	li := document.CreateElement("li").(*dom.HTMLLIElement)
	li.SetAttribute("data-icon", "false")
	li.SetAttribute("data-role", "collapsible")
	li.SetID(id)

	ul := document.CreateElement("ul").(*dom.HTMLUListElement)
	ul.SetAttribute("data-role", "listview")
	ul.SetID(id + "ul")

	c := document.CreateElement("span").(*dom.HTMLSpanElement)
	c.SetClass("ui-li-count")
	c.SetInnerHTML(count)

	menuButton := createButton("menu", id+"menu", "visible", "float-right")
	cutButton := createButton("content-cut", id+"cut", "invisible", "float-right")
	deleteButton := createButton("delete-outline", id+"delete", "invisible", "float-right")
	pasteButton := createButton("content-paste", id+"paste", "invisible", "float-right")
	addFolderButton := createButton("folder-plus-outline", id+"addFolder", "invisible", "float-right")
	addBookmarkButton := createButton("bookmark-plus-outline", id+"addBookmark", "invisible", "float-right")

	folderName := document.CreateElement("h1").(*dom.HTMLHeadingElement)
	folderName.SetInnerHTML(title)
	folderName.AddEventListener("click", false, func(event dom.Event) {
		// getting the href attribute to check if the click was on
		// the h1 element or on one of its children
		// if not, stopping the event propagation to prevent toggleing
		// the li
		h := event.Target().GetAttribute("href")
		if h != "#" {
			event.StopImmediatePropagation()
		}
	})

	folderName.AppendChild(c)
	folderName.AppendChild(menuButton)
	folderName.AppendChild(cutButton)
	folderName.AppendChild(deleteButton)
	folderName.AppendChild(pasteButton)
	folderName.AppendChild(addFolderButton)
	folderName.AppendChild(addBookmarkButton)

	li.AppendChild(folderName)
	li.AppendChild(ul)

	return li, ul
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
	blabel.SetID(id)
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
	// avoiding the propagation of the event to the parent h1
	// that would lead JQM to toggle the sublist (li)
	//ifoldername.SetAttribute("onclick", "event.stopPropagation()")
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
	// avoiding the propagation of the event to the parent h1
	// that would lead JQM to toggle the sublist (li)
	//ibookmarkname.SetAttribute("onclick", "event.stopPropagation()")
	ibookmarkname.SetClass("form-control")
	dc1.AppendChild(ibookmarkname)

	ibookmarkurl := document.CreateElement("input").(*dom.HTMLInputElement)
	ibookmarkurl.SetAttribute("type", "text")
	ibookmarkurl.SetAttribute("placeholder", "bookmark URL")
	// avoiding the propagation of the event to the parent h1
	// that would lead JQM to toggle the sublist (li)
	//ibookmarkurl.SetAttribute("onclick", "event.stopPropagation()")
	ibookmarkurl.SetClass("form-control")
	dc2.AppendChild(ibookmarkurl)

	ibookmarktags := document.CreateElement("select").(*dom.HTMLSelectElement)
	ibookmarktags.SetAttribute("placeholder", "tags")
	// avoiding the propagation of the event to the parent h1
	// that would lead JQM to toggle the sublist (li)
	//ibookmarktags.SetAttribute("onclick", "event.stopPropagation()")
	ibookmarktags.SetAttribute("multiple", "multiple")
	ibookmarktags.SetClass("form-control")
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
	jQuery(".folder-plus-outline").AddClass("invisible")
	jQuery(".bookmark-plus-outline").AddClass("invisible")
}

// hideForms hide all the create/update forms
func hideForms() {
	jQuery(".addFolder").Remove()
	jQuery(".addBookmark").Remove()
}

func bindButtonEvents(id string, isBookmark bool) {
	//
	// folder click event binding
	//
	jQuery("#"+id+" > h1").On("click", func(e jquery.Event) {
		fmt.Println("clic on " + id)
		hideActionButtons()
		hideForms()
	})

	//
	// buttons event binding
	//
	// menu
	jQuery("#"+id+"menu").On("click", func(e jquery.Event) {
		//e.StopPropagation()

		// make all other buttons and forms invisible
		hideActionButtons()
		hideForms()

		jQuery("#" + id + "cut").RemoveClass("invisible")
		jQuery("#" + id + "delete").RemoveClass("invisible")
		jQuery("#" + id + "star-outline").RemoveClass("invisible")
		jQuery("#" + id + "addFolder").RemoveClass("invisible")
		jQuery("#" + id + "addBookmark").RemoveClass("invisible")

		// something to paste ?
		if jQuery("input[type=hidden][name=cutednodeid]").Val() != "" {
			jQuery("#" + id + "paste").RemoveClass("invisible")
		}
	})

	// cut
	jQuery("#"+id+"cut").On("click", func(e jquery.Event) {
		//e.StopPropagation()

		jQuery("input[type=hidden][name=cutednodeid]").SetVal(id)
	})

	// delete
	jQuery("#"+id+"delete").On("click", func(e jquery.Event) {
		//e.StopPropagation()
	})

	// paste, addFolder, addBookmark
	if !isBookmark {
		jQuery("#"+id+"paste").On("click", func(e jquery.Event) {
			//e.StopPropagation()
			fmt.Println("clicked link paste " + id)
			jQuery("input[type=hidden][name=cutednodeid]").SetVal("")
		})

		jQuery("#"+id+"addFolder").On("click", func(e jquery.Event) {
			//e.StopPropagation()
			fmt.Println("clicked link add folder " + id)

			hideActionButtons()
			hideForms()

			f := createAddFolderForm(id)
			jQuery("li#" + id + " > h1").Append(f)

			// add event binding
			jQuery("#"+id+"createFolderSubmit").On("click", func(e jquery.Event) {
				//e.StopPropagation()

				folderName := jQuery("#" + id + "createFolderInput").Val()

				fmt.Println("create subfolder " + folderName + " of " + id)

				go createFolder(folderName, id)
			})
		})

		jQuery("#"+id+"addBookmark").On("click", func(e jquery.Event) {
			//e.StopPropagation()
			fmt.Println("clicked link add bookmark " + id)

			hideActionButtons()
			hideForms()

			b := createAddBookmarkForm(id)
			jQuery("li#" + id + " > h1").Append(b)

			// select2ify the form
			js.Global.Call("select2ify")

			// add event binding
			jQuery("#"+id+"createBookmarkSubmit").On("click", func(e jquery.Event) {
				//e.StopPropagation()

				fmt.Println("create bookmark in " + id)
			})
		})
	}

	// star
	if isBookmark {
		jQuery("#"+id+"star").On("click", func(e jquery.Event) {
			//e.StopPropagation()
			fmt.Println("clicked link star " + id)
		})
	}
}

// displayNode recursively display a Node as an JQM listview widget
func displayNode(n types.Node, e *dom.HTMLUListElement) {

	// Node keys are negative for bookmarks and positive for folders
	id := fmt.Sprintf("%d", n.Key)
	isBookmark := n.Key < 0

	switch isBookmark {
	case true:
		//
		// bookmark
		//
		f := createBookmarkNode(id, n.Title, n.URL, n.Icon)
		e.AppendChild(f)

	default:
		//
		// folder
		//
		// children count
		c := len(n.Children)
		b, u := createFolderNode(id, n.Title, fmt.Sprintf("%d", c))
		e.AppendChild(b)

		for _, c := range n.Children {
			displayNode(*c, u)
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
		displayNode(n, rootUL)
	}

	// https://stackoverflow.com/questions/6977338/jquery-mobile-listview-refresh
	jQuery("#tree").Trigger("create")

}

func main() {

	// when page is loaded...
	document.AddEventListener("DOMContentLoaded", false, func(event dom.Event) {
		event.PreventDefault()

		// getting tree div
		rootUL = document.GetElementByID("tree").(*dom.HTMLUListElement)

		// get remote folders and bookmarks
		go getNodes()
	})

	// exporting functions to be called from other JS files
	js.Global.Set("global", map[string]interface{}{})

}
