package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/tbellembois/gobkm/types"
	"honnef.co/go/js/dom"
	"honnef.co/go/js/xhr"
)

var (
	window   dom.Window
	document dom.Document
	rootUL   *dom.HTMLUListElement
)

func init() {
	window = dom.GetWindow()
	document = window.Document()
}

func displayNode(n types.Node, h *dom.HTMLUListElement) {

	l := len(n.Children)

	switch l {
	case 0:
		newLIChildElement := document.CreateElement("li").(*dom.HTMLLIElement)
		newSPANChildElement := document.CreateElement("span").(*dom.HTMLSpanElement)
		newSPANChildElement.SetInnerHTML(n.Title)

		newLIChildElement.AppendChild(newSPANChildElement)

		h.AppendChild(newLIChildElement)
	default:
		newULChildElement := document.CreateElement("ul").(*dom.HTMLUListElement)
		newULChildElement.SetAttribute("data-role", "listview")
		newULChildElement.SetAttribute("data-theme", "b")

		newLIChildElement := document.CreateElement("li").(*dom.HTMLLIElement)
		newLIChildElement.SetAttribute("data-role", "collapsible")
		newLIChildElement.SetAttribute("data-iconpos", "right")
		newLIChildElement.SetAttribute("data-inset", "false")

		newH2ChildElement := document.CreateElement("h2").(*dom.HTMLHeadingElement)
		newH2ChildElement.SetInnerHTML(n.Title)

		newLIChildElement.AppendChild(newH2ChildElement)
		newLIChildElement.AppendChild(newULChildElement)

		h.AppendChild(newLIChildElement)

		for _, c := range n.Children {
			displayNode(*c, newULChildElement)
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

	// TODO: call
	//$("tree").trigger("create");
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
