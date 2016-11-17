package models

import (
	"database/sql"

	log "github.com/Sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3" // register sqlite3 driver
	"github.com/tbellembois/gobkm/types"
)

const (
	dbdriver = "sqlite3"
)

// SQLiteDataStore implements the Datastore interface
// to store the folders and bookmarks in SQLite3.
type SQLiteDataStore struct {
	*sql.DB
	err error
}

// NewDBstore returns a database connection to the given dataSourceName
// ie. a path to the sqlite database file.
func NewDBstore(dataSourceName string) (*SQLiteDataStore, error) {
	log.WithFields(log.Fields{
		"dataSourceName": dataSourceName,
	}).Debug("NewDBstore:params")

	var (
		db  *sql.DB
		err error
	)

	if db, err = sql.Open(dbdriver, dataSourceName); err != nil {
		log.WithFields(log.Fields{
			"dataSourceName": dataSourceName,
		}).Error("NewDBstore:error opening the database")
		return nil, err
	}
	return &SQLiteDataStore{db, nil}, nil
}

// FlushErrors returns the last DB errors and flushes it.
func (db *SQLiteDataStore) FlushErrors() error {
	// Saving the last thrown error.
	lastError := db.err
	// Resetting the error.
	db.err = nil
	// Returning the last error.
	return lastError
}

// CreateDatabase creates the database tables.
func (db *SQLiteDataStore) CreateDatabase() {
	log.Info("Creating database")
	// Activate the foreign keys feature.
	if _, db.err = db.Exec("PRAGMA foreign_keys = ON"); db.err != nil {
		log.Error("CreateDatabase: error executing the PRAGMA request")
		return
	}

	// Tables creation if needed.
	if _, db.err = db.Exec("CREATE TABLE IF NOT EXISTS folder ( id integer PRIMARY KEY, title string NOT NULL, parentFolderId integer, nbChildrenFolders integer, FOREIGN KEY (parentFolderId) references folder(id) ON DELETE CASCADE)"); db.err != nil {
		log.Error("CreateDatabase: error executing the CREATE TABLE request for table bookmark")
		return
	}
	if _, db.err = db.Exec("CREATE TABLE IF NOT EXISTS bookmark ( id integer PRIMARY KEY, title string NOT NULL, url string NOT NULL, favicon string, starred integer, folderId integer, FOREIGN KEY (folderId) references folder(id) ON DELETE CASCADE)"); db.err != nil {
		log.Error("CreateDatabase: error executing the CREATE TABLE request for table bookmark")
		return
	}
	// Looking for folders.
	var count int
	if db.err = db.QueryRow("SELECT COUNT(*) as count FROM folder").Scan(&count); db.err != nil {
		log.Error("CreateDatabase: error executing the SELECT COUNT(*) request for table folder")
		return
	}
	// Inserting the / folder if not present.
	if count > 0 {
		log.Info("CreateDatabase: folder table not empty, leaving")
		return
	}
	if _, db.err = db.Exec("INSERT INTO folder(id, title) values(\"1\", \"/\")"); db.err != nil {
		log.Error("CreateDatabase: error inserting the root folder")
		return
	}
}

// PopulateDatabase populate the database with sample folders and bookmarks.
func (db *SQLiteDataStore) PopulateDatabase() {
	log.Info("Populating database")
	// Leaving silently on past errors...
	if db.err != nil {
		return
	}

	var (
		folders   []*types.Folder
		bookmarks []*types.Bookmark
		count     int
	)

	// Leaving if database is already populated.
	if db.err = db.QueryRow("SELECT COUNT(*) as count FROM folder").Scan(&count); db.err != nil || count > 1 {
		log.Info("Database not empty, leaving")
		return
	}

	// Getting the root folder.
	folderRoot := db.GetFolder(1)
	// Creating new sample folders.
	folder1 := types.Folder{Id: 1, Title: "IT", Parent: folderRoot}
	folder2 := types.Folder{Id: 2, Title: "Development", Parent: &folder1}
	// Creating new sample bookmark.
	bookmark1 := types.Bookmark{Id: 1, Title: "GoLang", Starred: true, URL: "https://golang.org/", Favicon: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAABHNCSVQICAgIfAhkiAAAAb9JREFUOI3tkj9oU1EUh797c3lgjA4xL61FX0yhMQqmW5QgFim4+GcyQ3Hp1MlBqFIyOGUobRScnYoQikNA0Ao6WJS2UIdiK7SUVGtfIZg0iMSA+Iy5Dg9fGnyLu2e6nHPu9zv3/K7QWuMXjfqebjQbOM5PIuEjHI6Ywq9P/TlUdm09+3KeNxtlAHbLWzTrNeTBQxjhHuLHohrgwqkBRi5dpO+4JQDEh80NfePOXaIDJ3FigximBUAyk+5SOvFphR/tNovvyzg769TKmxQLecS5a9d1dOQ2zp7N6bjF1PAZlJKMv1hFpVxIa+0t96+cBWD82TLr2zaGaVGbvYcEqLx+gmFajKZiqANBeo/2MZcb89RHUzEAeiNh5nJjGKZF9VUJAFks5FGVrc7IuuW7VH518slMGlHdpljII/sTSW+7j5ohEIrP9S9cnnxIaShOaSjOzNoOBNz81ceLHqg/kRRqv0ggGGLCdm3t+fqRmZtZ15HKEhN2Go1ABUO06VjfBdDSLQS0IFNd4fytSQAWHuR4B8gW7lWJP8B7rtA8zU7zfH4V8f0brew0ou37j/wBHigx2D2d/LvHJ/Vv8R8AvwHjjZMncK4ImgAAAABJRU5ErkJggg==", Folder: &folder2}
	folders = append(folders, &folder1, &folder2)
	bookmarks = append(bookmarks, &bookmark1)

	// DB save.
	for _, fld := range folders {
		db.SaveFolder(fld)
	}
	for _, bkm := range bookmarks {
		db.SaveBookmark(bkm)
	}
}

// GetBookmark returns a Bookmark instance with the given id.
func (db *SQLiteDataStore) GetBookmark(id int) *types.Bookmark {
	log.WithFields(log.Fields{
		"id": id,
	}).Debug("GetBookmark")
	// Leaving silently on past errors...
	if db.err != nil {
		return nil
	}

	var (
		folderID sql.NullInt64
		starred  sql.NullInt64
	)

	// Querying the bookmark.
	bkm := new(types.Bookmark)
	db.err = db.QueryRow("SELECT id, title, url, favicon, starred, folderId FROM bookmark WHERE id=?", id).Scan(&bkm.Id, &bkm.Title, &bkm.URL, &bkm.Favicon, &starred, &folderID)
	switch {
	case db.err == sql.ErrNoRows:
		log.WithFields(log.Fields{
			"id": id,
		}).Debug("GetBookmark:no bookmark with that ID")
		return nil
	case db.err != nil:
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("GetBookmark:SELECT query error")
		return nil
	default:
		log.WithFields(log.Fields{
			"Id":       bkm.Id,
			"Title":    bkm.Title,
			"folderId": folderID,
			"Favicon":  bkm.Favicon,
		}).Debug("GetBookmark:bookmark found")
		// Starred bookmark ?
		if int(starred.Int64) != 0 {
			bkm.Starred = true
		}
		// Retrieving the parent folder if it is not the root (/).
		if folderID.Int64 != 0 {
			bkm.Folder = db.GetFolder(int(folderID.Int64))
			if db.err != nil {
				log.WithFields(log.Fields{
					"err": db.err,
				}).Error("GetBookmark:parent Folder retrieving error")
				return nil
			}
		}
	}
	return bkm
}

// GetFolder returns a Folder instance with the given id.
func (db *SQLiteDataStore) GetFolder(id int) *types.Folder {
	log.WithFields(log.Fields{
		"id": id,
	}).Debug("GetFolder")
	// Leaving silently on past errors...
	if db.err != nil || id == 0 {
		return nil
	}

	// Querying the folder.
	var parentFldID sql.NullInt64
	fld := new(types.Folder)
	db.err = db.QueryRow("SELECT id, title, parentFolderId FROM folder WHERE id=?", id).Scan(&fld.Id, &fld.Title, &parentFldID)
	switch {
	case db.err == sql.ErrNoRows:
		log.WithFields(log.Fields{
			"id": id,
		}).Debug("GetFolder:no folder with that ID")
		return nil
	case db.err != nil:
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("GetFolder:SELECT query error")
		return nil
	default:
		log.WithFields(log.Fields{
			"Id":          fld.Id,
			"Title":       fld.Title,
			"parentFldId": parentFldID,
		}).Debug("GetFolder:folder found")
		// recursively retrieving the parents
		if parentFldID.Int64 != 0 {
			fld.Parent = db.GetFolder(int(parentFldID.Int64))
		}
	}
	return fld
}

// GetStarredBookmarks returns the starred bookmarks.
func (db *SQLiteDataStore) GetStarredBookmarks() []*types.Bookmark {
	// Leaving silently on past errors...
	if db.err != nil {
		return nil
	}

	// Querying the bookmarks.
	var (
		rows *sql.Rows
		bkms []*types.Bookmark
	)
	rows, db.err = db.Query("SELECT id, title, url, favicon, starred, folderId FROM bookmark WHERE starred ORDER BY title")
	defer func() {
		if db.err = rows.Close(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("GetStarredBookmarks:error closing rows")
		}
	}()

	switch {
	case db.err == sql.ErrNoRows:
		log.Debug("GetStarredBookmarks:no bookmarks")
		return nil
	case db.err != nil:
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("GetStarredBookmarks:SELECT query error")
		return nil
	default:
		for rows.Next() {
			// Building a new Bookmark instance with each row.
			bkm := new(types.Bookmark)
			var fldID sql.NullInt64
			db.err = rows.Scan(&bkm.Id, &bkm.Title, &bkm.URL, &bkm.Favicon, &bkm.Starred, &fldID)
			if db.err != nil {
				log.WithFields(log.Fields{
					"err": db.err,
				}).Error("GetStarredBookmarks:error scanning the query result row")
				return nil
			}
			// Retrieving the bookmark folder.
			bkm.Folder = db.GetFolder(int(fldID.Int64))
			bkms = append(bkms, bkm)
		}
		if db.err = rows.Err(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("GetStarredBookmarks:error looping rows")
			return nil
		}
		return bkms
	}
}

// GetNoIconBookmarks returns the bookmarks with no favicon.
func (db *SQLiteDataStore) GetNoIconBookmarks() []*types.Bookmark {
	// Leaving silently on past errors...
	if db.err != nil {
		return nil
	}
	var (
		rows *sql.Rows
		bkms []*types.Bookmark
	)

	// Querying the bookmarks.
	rows, db.err = db.Query("SELECT id, title, url, favicon, starred, folderId FROM bookmark WHERE favicon='' ORDER BY title")
	defer func() {
		if db.err = rows.Close(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("GetNoIconBookmarks:error closing rows")
		}
	}()

	switch {
	case db.err == sql.ErrNoRows:
		log.Debug("GetNoIconBookmarks:no bookmarks")
		return nil
	case db.err != nil:
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("GetNoIconBookmarks:SELECT query error")
		return nil
	default:
		for rows.Next() {
			// Building a new Bookmark instance with each row.
			bkm := new(types.Bookmark)
			var fldID sql.NullInt64
			var starred sql.NullInt64
			db.err = rows.Scan(&bkm.Id, &bkm.Title, &bkm.URL, &bkm.Favicon, &starred, &fldID)
			if db.err != nil {
				log.WithFields(log.Fields{
					"err": db.err,
				}).Error("GetNoIconBookmarks:error scanning the query result row")
				return nil
			}
			// Starred bookmark ?
			if int(starred.Int64) != 0 {
				bkm.Starred = true
			}
			// Retrieving the bookmark folder.
			bkm.Folder = db.GetFolder(int(fldID.Int64))
			bkms = append(bkms, bkm)
		}
		if db.err = rows.Err(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("GetNoIconBookmarks:error looping rows")
			return nil
		}
		return bkms
	}
}

// GetAllBookmarks returns all the bookmarks as an array of *Bookmark.
func (db *SQLiteDataStore) GetAllBookmarks() []*types.Bookmark {
	// Leaving silently on past errors...
	if db.err != nil {
		return nil
	}
	var (
		rows *sql.Rows
		bkms []*types.Bookmark
	)

	// Querying the bokmarks.
	rows, db.err = db.Query("SELECT * FROM bookmark ORDER BY title")
	defer func() {
		if db.err = rows.Close(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("GetAllBookmarks:error closing rows")
		}
	}()

	switch {
	case db.err == sql.ErrNoRows:
		log.Debug("GetAllBookmarks:no bookmarks")
		return nil
	case db.err != nil:
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("GetAllBookmarks:SELECT query error")
		return nil
	default:
		for rows.Next() {
			bkm := new(types.Bookmark)
			var fldID sql.NullInt64
			var starred sql.NullInt64
			// Building a new Bookmark instance with each row.
			db.err = rows.Scan(&bkm.Id, &bkm.Title, &bkm.URL, &bkm.Favicon, &starred, &fldID)
			if db.err != nil {
				log.WithFields(log.Fields{
					"err": db.err,
				}).Error("GetAllBookmarks:error scanning the query result row")
				return nil
			}
			// Starred bookmark ?
			if int(starred.Int64) != 0 {
				bkm.Starred = true
			}
			// Retrieving the bookmark folder.
			bkm.Folder = db.GetFolder(int(fldID.Int64))
			bkms = append(bkms, bkm)
		}
		if db.err = rows.Err(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("GetAllBookmarks:error looping rows")
			return nil
		}
		return bkms
	}
}

// SearchBookmarks returns the bookmarks with the title containing the given string.
func (db *SQLiteDataStore) SearchBookmarks(s string) []*types.Bookmark {
	log.WithFields(log.Fields{
		"s": s,
	}).Debug("SearchBookmarks")
	// Leaving silently on past errors...
	if db.err != nil {
		return nil
	}
	var (
		rows *sql.Rows
		bkms []*types.Bookmark
	)

	// Querying the bookmarks.
	rows, db.err = db.Query("SELECT id, title, url, favicon, starred, folderId FROM bookmark WHERE title LIKE ? ORDER BY title", "%"+s+"%")
	defer func() {
		if db.err = rows.Close(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("SearchBookmarks:error closing rows")
		}
	}()
	switch {
	case db.err == sql.ErrNoRows:
		log.Debug("SearchBookmarks:no bookmarks")
		return nil
	case db.err != nil:
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("SearchBookmarks:SELECT query error")
		return nil
	default:
		for rows.Next() {
			// Building a new Bookmark instance with each row.
			bkm := new(types.Bookmark)
			var parentFldID sql.NullInt64
			var starred sql.NullInt64
			db.err = rows.Scan(&bkm.Id, &bkm.Title, &bkm.URL, &bkm.Favicon, &starred, &parentFldID)
			// Starred bookmark ?
			if int(starred.Int64) != 0 {
				bkm.Starred = true
			}
			if db.err != nil {
				log.WithFields(log.Fields{
					"err": db.err,
				}).Error("SearchBookmarks:error scanning the query result row")
				return nil
			}
			log.WithFields(log.Fields{
				"bkm": bkm,
			}).Debug("SearchBookmarks:bookmark found")
			bkms = append(bkms, bkm)
		}
		if db.err = rows.Err(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("SearchBookmarks:error looping rows")
			return nil
		}
		return bkms
	}
}

// GetFolderBookmarks returns the bookmarks of the given folder id.
func (db *SQLiteDataStore) GetFolderBookmarks(id int) []*types.Bookmark {
	log.WithFields(log.Fields{
		"id": id,
	}).Debug("GetFolderBookmarks")
	// Leaving silently on past errors...
	if db.err != nil {
		return nil
	}
	var (
		rows *sql.Rows
		bkms []*types.Bookmark
	)

	// Querying the bookmarks.
	rows, db.err = db.Query("SELECT id, title, url, favicon, starred, folderId FROM bookmark WHERE folderId is ? ORDER BY title", id)
	defer func() {
		if db.err = rows.Close(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("GetFolderBookmarks:error closing rows")
		}
	}()

	switch {
	case db.err == sql.ErrNoRows:
		log.Debug("GetFolderBookmarks:no bookmarks")
		return nil
	case db.err != nil:
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("GetFolderBookmarks:SELECT query error")
		return nil
	default:
		for rows.Next() {
			// Building a new Bookmark instance with each row.
			bkm := new(types.Bookmark)
			var parentFldID sql.NullInt64
			var starred sql.NullInt64
			db.err = rows.Scan(&bkm.Id, &bkm.Title, &bkm.URL, &bkm.Favicon, &starred, &parentFldID)
			// Starred bookmark ?
			if int(starred.Int64) != 0 {
				bkm.Starred = true
			}
			if db.err != nil {
				log.WithFields(log.Fields{
					"err": db.err,
				}).Error("GetFolderBookmarks:error scanning the query result row")
				return nil
			}
			log.WithFields(log.Fields{
				"bkm": bkm,
			}).Debug("GetFolderBookmarks:bookmark found")
			bkms = append(bkms, bkm)
		}
		if db.err = rows.Err(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("GetFolderBookmarks:error looping rows")
			return nil
		}
		return bkms
	}
}

// GetFolderSubfolders returns the children folders as an array of *Folder
func (db *SQLiteDataStore) GetFolderSubfolders(id int) []*types.Folder {
	log.WithFields(log.Fields{
		"id": id,
	}).Debug("GetChildrenFolders")
	// Leaving silently on past errors...
	if db.err != nil {
		return nil
	}

	var (
		rows *sql.Rows
		flds []*types.Folder
	)
	// Querying the folders.
	rows, db.err = db.Query("SELECT id, title, parentFolderId, nbChildrenFolders FROM folder WHERE parentFolderId is ? ORDER BY title", id)
	defer func() {
		if db.err = rows.Close(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("GetFolderSubfolders:error closing rows")
		}
	}()
	switch {
	case db.err == sql.ErrNoRows:
		log.Debug("GetChildrenFolders:no folders")
		return nil
	case db.err != nil:
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("GetChildrenFolders:SELECT query error")
		return nil
	default:
		for rows.Next() {
			// Building a new Folder instance with each row.
			fld := new(types.Folder)
			var parentFldID sql.NullInt64
			db.err = rows.Scan(&fld.Id, &fld.Title, &parentFldID, &fld.NbChildrenFolders)
			if db.err != nil {
				log.WithFields(log.Fields{
					"err": db.err,
				}).Error("GetChildrenFolders:error scanning the query result row")
				return nil
			}
			flds = append(flds, fld)
		}
		if db.err = rows.Err(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("GetChildrenFolders:error looping rows")
			return nil
		}
		return flds
	}
}

// GetRootFolders returns the root folders as an array of *Folder
func (db *SQLiteDataStore) GetRootFolders() []*types.Folder {
	// Leaving silently on past errors...
	if db.err != nil {
		return nil
	}
	var (
		rows *sql.Rows
		flds []*types.Folder
	)

	// Querying the folders.
	rows, db.err = db.Query("SELECT id, title, parentFolderId, nbChildrenFolders FROM folder WHERE parentFolderId==1 ORDER BY title")
	defer func() {
		if db.err = rows.Close(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("GetNoIconBookmarks:error closing rows")
		}
	}()
	switch {
	case db.err == sql.ErrNoRows:
		log.Debug("GetRootFolders:no folders")
		return nil
	case db.err != nil:
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("GetRootFolders:SELECT query error")
		return nil
	default:
		for rows.Next() {
			// Building a new Folder instance with each row.
			fld := new(types.Folder)
			var parentFldID sql.NullInt64
			db.err = rows.Scan(&fld.Id, &fld.Title, &parentFldID, &fld.NbChildrenFolders)
			if db.err != nil {
				log.WithFields(log.Fields{
					"err": db.err,
				}).Error("GetRootFolders:error scanning the query result row")
				return nil
			}
			flds = append(flds, fld)
		}
		if db.err = rows.Err(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("GetRootFolders:error looping rows")
			return nil
		}
		return flds
	}
}

// SaveFolder saves the given new Folder into the db and returns the folder id.
// Called only on folder creation or rename
// so only the Title has to be set.
func (db *SQLiteDataStore) SaveFolder(f *types.Folder) int64 {
	log.WithFields(log.Fields{
		"f": f,
	}).Debug("SaveFolder")
	// Leaving silently on past errors...
	if db.err != nil {
		return 0
	}
	var stmt *sql.Stmt

	// Preparing the query.
	// id will be auto incremented
	if stmt, db.err = db.Prepare("INSERT INTO folder(title, parentFolderId, nbChildrenFolders) values(?,?,?)"); db.err != nil {
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("SaveFolder:SELECT request prepare error")
		return 0
	}
	defer func() {
		if db.err = stmt.Close(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("SaveFolder:error closing stmt")
		}
	}()

	// Executing the query.
	var res sql.Result
	if f.Parent != nil {
		res, db.err = stmt.Exec(f.Title, f.Parent.Id, f.NbChildrenFolders)
	} else {
		res, db.err = stmt.Exec(f.Title, 1, f.NbChildrenFolders)
	}
	id, _ := res.LastInsertId() // we should check the error here too...
	if db.err != nil {
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("SaveFolder:INSERT query error")
		return 0
	}
	return id
}

// UpdateBookmark updates the given bookmark.
func (db *SQLiteDataStore) UpdateBookmark(b *types.Bookmark) {
	log.WithFields(log.Fields{
		"b": b,
	}).Debug("UpdateBookmark")
	// Leaving silently on past errors...
	if db.err != nil {
		return
	}

	var (
		stmt *sql.Stmt
		tx   *sql.Tx
	)
	// Beginning a new transaction.
	// TODO: is a transaction needed here?
	tx, db.err = db.Begin()
	if db.err != nil {
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("Update bookmark:transaction begin failed")
		return
	}

	// Preparing the update request.
	stmt, db.err = tx.Prepare("UPDATE bookmark SET title=?, url=?, folderId=?, starred=?, favicon=? WHERE id=?")
	if db.err != nil {
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("Update bookmark:UPDATE request prepare error")
		return
	}
	defer func() {
		if db.err = stmt.Close(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("UpdateBookmark:error closing stmt")
		}
	}()

	// Executing the query.
	if b.Folder != nil {
		_, db.err = stmt.Exec(b.Title, b.URL, b.Folder.Id, b.Starred, b.Favicon, b.Id)
	} else {
		_, db.err = stmt.Exec(b.Title, b.URL, 1, b.Starred, b.Favicon, b.Id)
	}
	// Rolling back on errors, or commit.
	if db.err != nil {
		if db.err = tx.Rollback(); db.err != nil {
			// Just logging the error.
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("UpdateBookmark: UPDATE query transaction rollback error")
		}
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("UpdateBookmark: UPDATE query error")
		return
	}
	if db.err = tx.Commit(); db.err != nil {
		// Just logging the error.
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("UpdateBookmark: UPDATE query transaction commit error")
	}
}

// SaveBookmark saves the new given Bookmark into the db
func (db *SQLiteDataStore) SaveBookmark(b *types.Bookmark) int64 {
	log.WithFields(log.Fields{
		"b": b,
	}).Debug("SaveBookmark")
	// Leaving silently on past errors...
	if db.err != nil {
		return 0
	}

	// Preparing the query.
	var stmt *sql.Stmt
	stmt, db.err = db.Prepare("INSERT INTO bookmark(title, url, folderId, favicon) values(?,?,?,?)")
	if db.err != nil {
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("SaveBookmark:INSERT request prepare error")
		return 0
	}
	defer func() {
		if db.err = stmt.Close(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("SaveBookmark:error closing stmt")
		}
	}()

	// Executing the query.
	var res sql.Result
	if b.Folder != nil {
		res, db.err = stmt.Exec(b.Title, b.URL, b.Folder.Id, b.Favicon)
	} else {
		res, db.err = stmt.Exec(b.Title, b.URL, 1, b.Favicon)
	}
	if db.err != nil {
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("SaveBookmark:INSERT query error")
		return 0
	}
	id, _ := res.LastInsertId()
	return id
}

// DeleteBookmark delete the given Bookmark from the db
func (db *SQLiteDataStore) DeleteBookmark(b *types.Bookmark) {
	log.WithFields(log.Fields{
		"b": b,
	}).Debug("DeleteBookmark")
	// Leaving silently on past errors...
	if db.err != nil {
		return
	}

	// Executing the query.
	_, db.err = db.Exec("DELETE from bookmark WHERE id=?", b.Id)
	if db.err != nil {
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("DeleteBookmark:DELETE query error")
		return
	}
	return
}

// UpdateFolder updates the given folder.
func (db *SQLiteDataStore) UpdateFolder(f *types.Folder) {
	log.WithFields(log.Fields{
		"f": f,
	}).Debug("UpdateFolder")
	// Leaving silently on past errors...
	if db.err != nil {
		return
	}

	var oldParentFolderID sql.NullInt64
	// Retrieving the parentFolderId of the folder to be updated.
	if db.err = db.QueryRow("SELECT parentFolderId from folder WHERE id=?", f.Id).Scan(&oldParentFolderID); db.err != nil {
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("UpdateFolder:SELECT query error")
		return
	}
	log.WithFields(log.Fields{
		"oldParentFolderId": oldParentFolderID,
		"f.Parent":          f.Parent,
	}).Debug("UpdateFolder")

	// Preparing the update request for the folder.
	var stmt *sql.Stmt
	stmt, db.err = db.Prepare("UPDATE folder SET title=?, parentFolderId=?, nbChildrenFolders=(SELECT count(*) from folder WHERE parentFolderId=?) WHERE id=?")
	if db.err != nil {
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("UpdateFolder:UPDATE request prepare error")
		return
	}
	defer func() {
		if db.err = stmt.Close(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("UpdateFolder:error closing stmt")
		}
	}()

	// Executing the query.
	if f.Parent != nil {
		_, db.err = stmt.Exec(f.Title, f.Parent.Id, f.Id, f.Id)
	} else {
		_, db.err = stmt.Exec(f.Title, 1, f.Id, f.Id)
	}
	if db.err != nil {
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("UpdateFolder:UPDATE query error")
		return
	}

	// Preparing the update request for the old and new parent folders (to update the nbChildrenFolders).
	stmt, db.err = db.Prepare("UPDATE folder SET nbChildrenFolders=(SELECT count(*) from folder WHERE parentFolderId=?) WHERE id=?")
	if db.err != nil {
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("UpdateFolder:UPDATE old parent request prepare error")
		return
	}
	defer func() {
		if db.err = stmt.Close(); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("UpdateFolder:error closing stmt")
		}
	}()

	// Executing the query for the old parent.
	if _, db.err = stmt.Exec(oldParentFolderID, oldParentFolderID); db.err != nil {
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("UpdateFolder:UPDATE old parent request error")
		return
	}
	// And the new.
	if f.Parent != nil {
		if _, db.err = stmt.Exec(f.Parent.Id, f.Parent.Id); db.err != nil {
			log.WithFields(log.Fields{
				"err": db.err,
			}).Error("UpdateFolder:UPDATE new parent request error")
			return
		}
	}
}

// DeleteFolder delete the given Folder from the db.
func (db *SQLiteDataStore) DeleteFolder(f *types.Folder) {
	log.WithFields(log.Fields{
		"f": f,
	}).Debug("DeleteFolder")
	// Leaving silently on past errors...
	if db.err != nil {
		return
	}

	// Executing the query.
	_, db.err = db.Exec("DELETE from folder WHERE id=?", f.Id)
	if db.err != nil {
		log.WithFields(log.Fields{
			"err": db.err,
		}).Error("DeleteFolder:DELETE query error")
		return
	}
	return
}
