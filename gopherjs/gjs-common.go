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

func displayNode(n types.Node, h *dom.HTMLUListElement) {

	l := len(n.Children)

	switch l {
	case 0:
		id := fmt.Sprintf("%d", n.Key)

		newLIChildElement := document.CreateElement("li").(*dom.HTMLLIElement)
		newLIChildElement.SetID(id)

		newIMGChildElement := document.CreateElement("img").(*dom.HTMLImageElement)
		newIMGChildElement.SetClass("ui-li-icon")
		newIMGChildElement.SetAttribute("src", n.Icon)

		newSPANChildElement := document.CreateElement("a").(*dom.HTMLAnchorElement)
		newSPANChildElement.SetAttribute("href", "#")
		newSPANChildElement.SetInnerHTML(n.Title)
		newSPANChildElement.SetID(id + "link")

		newLIChildElement.AppendChild(newIMGChildElement)
		newLIChildElement.AppendChild(newSPANChildElement)

		h.AppendChild(newLIChildElement)

		jQuery(document).On("vclick", id, func(e jquery.Event) {
			fmt.Println("toto clicked on " + jQuery(e.Target).Val())
		})

	default:
		id := fmt.Sprintf("%d", n.Key)

		newLIChildElement := document.CreateElement("li").(*dom.HTMLLIElement)
		newLIChildElement.SetAttribute("data-role", "collapsible")
		newLIChildElement.SetAttribute("data-iconpos", "right")
		newLIChildElement.SetID(id)

		newULChildElement := document.CreateElement("ul").(*dom.HTMLUListElement)
		newULChildElement.SetAttribute("data-role", "listview")
		newULChildElement.SetAttribute("data-inset", "true")
		newULChildElement.SetID(string(n.Key))

		newSPANChildElement := document.CreateElement("span").(*dom.HTMLSpanElement)
		newSPANChildElement.SetClass("ui-li-count")
		newSPANChildElement.SetInnerHTML(fmt.Sprintf("%d", l))

		newH2ChildElement := document.CreateElement("h1").(*dom.HTMLHeadingElement)
		newH2ChildElement.SetInnerHTML(n.Title)

		newH2ChildElement.AppendChild(newSPANChildElement)
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
