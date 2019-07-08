package main

import (
	"encoding/json"
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

func createButton(icon string, id string, visibility string) *dom.HTMLButtonElement {
	b := document.CreateElement("button").(*dom.HTMLButtonElement)
	b.SetAttribute("type", "button")
	b.SetAttribute("data-role", "none")
	b.SetClass("btn btn-outline-dark " + visibility)
	b.SetID(id)

	blabel := document.CreateElement("span").(*dom.HTMLSpanElement)
	blabel.SetID(id + "menu")
	blabel.SetClass("mdi mdi-24px mdi-" + icon)

	b.AppendChild(blabel)

	return b
}

func displayNode(n types.Node, e *dom.HTMLUListElement) {

	l := len(n.Children)

	switch l {
	case 0:
		//
		// e has no children
		//
		id := fmt.Sprintf("%d", n.Key)

		li := document.CreateElement("li").(*dom.HTMLLIElement)
		li.SetAttribute("data-icon", "false")
		li.SetID(id)

		favicon := document.CreateElement("img").(*dom.HTMLImageElement)
		favicon.SetClass("ui-li-icon")
		favicon.SetAttribute("src", n.Icon)

		mainSpan := document.CreateElement("span").(*dom.HTMLSpanElement)

		menuButton := createButton("menu", id+"menu", "visible")
		cutButton := createButton("content-cut", id+"cut", "invisible")
		deleteButton := createButton("delete", id+"delete", "invisible")

		link := document.CreateElement("a").(*dom.HTMLAnchorElement)
		link.SetAttribute("href", n.URL)
		link.SetAttribute("target", "_blank")
		link.SetID(id + "link")
		link.SetInnerHTML(n.Title)

		mainSpan.AppendChild(link)
		mainSpan.AppendChild(menuButton)
		mainSpan.AppendChild(cutButton)
		mainSpan.AppendChild(deleteButton)

		li.AppendChild(favicon)
		li.AppendChild(mainSpan)

		e.AppendChild(li)

		jQuery("#"+id+"menu").On("click", func(e jquery.Event) {
			e.StopPropagation()
			fmt.Println("clicked link menu " + id)
			jQuery("#" + id + "cut").RemoveClass("invisible")
			jQuery("#" + id + "delete").RemoveClass("invisible")
		})

	default:
		//
		// e has children
		//
		id := fmt.Sprintf("%d", n.Key)

		li := document.CreateElement("li").(*dom.HTMLLIElement)
		li.SetAttribute("data-icon", "false")
		li.SetAttribute("data-role", "collapsible")
		li.SetID(id)

		ul := document.CreateElement("ul").(*dom.HTMLUListElement)
		ul.SetAttribute("data-role", "listview")
		ul.SetID(string(n.Key))

		count := document.CreateElement("span").(*dom.HTMLSpanElement)
		count.SetClass("ui-li-count")
		count.SetInnerHTML(fmt.Sprintf("%d", l))

		menuButton := createButton("menu", id+"menu", "visible")

		folderName := document.CreateElement("h1").(*dom.HTMLHeadingElement)
		folderName.SetInnerHTML(n.Title)

		folderName.AppendChild(count)
		folderName.AppendChild(menuButton)
		li.AppendChild(folderName)
		li.AppendChild(ul)

		e.AppendChild(li)

		jQuery("#"+id+"menu").On("click", func(e jquery.Event) {
			e.StopPropagation()
			fmt.Println("clicked folder menu " + id)
			jQuery("#" + id + "cut").RemoveClass("invisible")
			jQuery("#" + id + "delete").RemoveClass("invisible")
		})

		for _, c := range n.Children {
			displayNode(*c, ul)
		}

	}

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

		jQuery(document).On("mobileinit", func(e jquery.Event) {
			fmt.Println("mobile init")
		})

		// getting tree div
		rootUL = document.GetElementByID("tree").(*dom.HTMLUListElement)

		// get remote folders and bookmarks
		go getNodes()

	})

	// exporting functions to be called from other JS files
	js.Global.Set("global", map[string]interface{}{})

}
