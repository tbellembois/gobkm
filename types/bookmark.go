package types

import (
	"encoding/json"
	"strings"
)

// Folder containing the bookmarks
type Folder struct {
	Id                int
	Title             string
	Parent            *Folder
	NbChildrenFolders int
}

// Bookmark
type Bookmark struct {
	Id      int
	Title   string
	URL     string
	Favicon string // base64 encoded image
	Starred bool
	Folder  *Folder
	Tags    []*Tag
}

// Tag
type Tag struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// Node
type Node struct {
	Key      int     `json:"id"`
	Title    string  `json:"text"`
	URL      string  `json:"url"`
	Folder   bool    `json:"hasChildren"`
	Lazy     bool    `json:"lazy"`
	Icon     string  `json:"icon"`
	Children []*Node `json:"children"`
	Tags     []*Tag  `json:"tag"`
}

// Bookmarks implements the sort interface
type Bookmarks []*Bookmark

func (b Bookmarks) Len() int {
	return len(b)
}

func (b Bookmarks) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b Bookmarks) Less(i, j int) bool {
	url1 := b[i].Title
	url2 := b[j].Title
	title1 := url1[strings.Index(url1, "//")+2 : len(url1)]
	title2 := url2[strings.Index(url2, "//")+2 : len(url2)]
	return title1 < title2
}

func (bk *Bookmark) String() string {
	var out []byte
	var err error

	if out, err = json.Marshal(bk); err != nil {
		return ""
	}
	return string(out)
}

// PathString returns the bookmark full path as a string
func (bk *Bookmark) PathString() string {
	var (
		p *Folder
		r string
	)
	for p = bk.Folder; p != nil; p = p.Parent {
		r += "/" + p.Title
	}
	return r
}

func (fd *Folder) String() string {
	var out []byte
	var err error

	if out, err = json.Marshal(fd); err != nil {
		return ""
	}
	return string(out)
}

// IsRootFolder returns true if the given Folder has no parent
func (fd *Folder) IsRootFolder() bool {
	return fd.Parent == nil
}

// HasChildrenFolders returns true if the given Folder has children
func (fd *Folder) HasChildrenFolders() bool {
	return fd.NbChildrenFolders > 0
}
