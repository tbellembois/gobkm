package models

import (
	"github.com/tbellembois/gobkm/types"
)

// Datastore is a folders and bookmarks storage interface.
type Datastore interface {
	FlushErrors() error

	SearchBookmarks(string) []*types.Bookmark
	GetBookmark(int) *types.Bookmark
	GetBookmarkTags(int) []*types.Tag
	GetFolderBookmarks(int) types.Bookmarks
	SaveBookmark(*types.Bookmark) int64
	UpdateBookmark(*types.Bookmark)
	DeleteBookmark(*types.Bookmark)

	GetFolder(int) *types.Folder
	GetFolderSubfolders(int) []*types.Folder
	SaveFolder(*types.Folder) int64
	UpdateFolder(*types.Folder)
	DeleteFolder(*types.Folder)

	GetTags() []*types.Tag
	GetStars() []*types.Bookmark
	GetTag(int) *types.Tag
	SaveTag(*types.Tag) int64
}
