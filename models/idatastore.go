package models

import (
	"github.com/tbellembois/gobkm/types"
)

// Datastore is a folders and bookmarks storage interface
type Datastore interface {
	FlushErrors() error

	SearchBookmarks(string) []*types.Bookmark
	GetAllBookmarks() []*types.Bookmark
	GetBookmark(int) *types.Bookmark
	GetFolderBookmarks(int) types.Bookmarks
	GetNoIconBookmarks() []*types.Bookmark
	GetStarredBookmarks() []*types.Bookmark
	SaveBookmark(*types.Bookmark) int64
	UpdateBookmark(*types.Bookmark)
	DeleteBookmark(*types.Bookmark)

	GetFolder(int) *types.Folder
	GetFolderSubfolders(int) []*types.Folder
	GetRootFolders() []*types.Folder
	SaveFolder(*types.Folder) int64
	UpdateFolder(*types.Folder)
	DeleteFolder(*types.Folder)
}
