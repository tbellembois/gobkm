package types

import (
	"encoding/json"
	"strings"
)

// Folder containing the bookmarks
type Folder struct {
	Id                int         `json:"id"`
	Title             string      `json:"title"`
	Parent            *Folder     `json:"parent"`
	Folders           []*Folder   `json:"folders"`
	Bookmarks         []*Bookmark `json:"bookmarks"`
	NbChildrenFolders int         `json:"nbchildrenfolders"`
}

// Bookmark
type Bookmark struct {
	Id      int     `json:"id"`
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Favicon string  `json:"favicon"` // base64 encoded image
	Starred bool    `json:"starred"`
	Folder  *Folder `json:"folder"` // reference to the folder to help
	Tags    []*Tag  `json:"tags"`
}

// Tag represents a bookmark tag
type Tag struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
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
	title1 := url1[strings.Index(url1, "//")+2:]
	title2 := url2[strings.Index(url2, "//")+2:]
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
