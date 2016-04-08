package types

import "encoding/json"

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
}

func (fd *Folder) String() string {

	var out []byte
	var err error

	if out, err = json.Marshal(fd); err != nil {
		return ""
	}

	return string(out)
}

func (bk *Bookmark) String() string {

	var out []byte
	var err error

	if out, err = json.Marshal(bk); err != nil {
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
