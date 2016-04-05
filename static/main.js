/*----------------
 *
 * vars and consts
 *
 *----------------*/

var CLASS_ITEM_OVER = "folder-over";
var CLASS_ITEM_FOLDER = "folder";
var CLASS_ITEM_FOLDER_OPEN = "fa-folder-open-o";
var CLASS_ITEM_FOLDER_CLOSED = "fa-folder-o";

var CLASS_ITEM_BOOKMARK = "bookmark";

/*-------------------------------
 *
 * Bind enter key to add a folder
 *
 *-------------------------------*/
document.addEventListener("keydown", keyDownPress, false);

function keyDownPress(e) {

    if (e.which == 13) {
        if(document.getElementById('add-folder').disabled) {
            renameFolder();
        }
        else {
            addFolder();
        }
    }

}

/*----------------
 *
 * Utils fonctions
 *
 *----------------*/

// http://gomakethings.com/climbing-up-and-down-the-dom-tree-with-vanilla-javascript/
/**
 * Get closest DOM element up the tree that contains a class, ID, or data attribute
 * @param  {Node} elem The base element
 * @param  {String} selector The class, id, data attribute, or tag to look for
 * @return {Node} Null if no match
 */
var getClosest = function (elem, selector) {

    var firstChar = selector.charAt(0);

    // Get closest match
    for ( ; elem && elem !== document; elem = elem.parentNode ) {

        // If selector is a class
        if ( firstChar === '.' ) {
            if ( elem.classList.contains( selector.substr(1) ) ) {
                return elem;
            }
        }

        // If selector is an ID
        if ( firstChar === '#' ) {
            if ( elem.id === selector.substr(1) ) {
                return elem;
            }
        } 

        // If selector is a data attribute
        if ( firstChar === '[' ) {
            if ( elem.hasAttribute( selector.substr(1, selector.length - 2) ) ) {
                return elem;
            }
        }

        // If selector is a tag
        if ( elem.tagName.toLowerCase() === selector ) {
            return elem;
        }

    }

    return false;

};

/*
 * toogleDisplayImport shows/hiddes the import box form 
 * */
function toogleDisplayImport() {

    visibility = document.getElementById('import-input-box').style.display

    if (visibility == 'none') {
        document.getElementById('import-input-box').style.display = 'block';
    }
    else {
        document.getElementById('import-input-box').style.display = 'none';
    }

}

/*
 * showRenameBox shows the rename box form 
 * */
function showRenameBox() {
    document.getElementById('rename-input-box').style.visibility = 'visible';
   
    document.getElementById('add-folder').disabled = true;
    document.getElementById('add-folder-button').disabled = true;
}

/*
 * hideRenameBox hides the rename box form 
 * */
function hideRenameBox() {
    document.getElementById('rename-input-box').style.visibility = 'hidden';
    document.getElementById('rename-input-box-form').value = "";
    document.getElementById('rename-hidden-input-box-form').value = "";

    document.getElementById('add-folder').disabled = false;
    document.getElementById('add-folder-button').disabled = false;
}

/*
 * setRenameFormValue initializes the rename input form with val 
 * */
function setRenameFormValue(val) {
    document.getElementById('rename-input-box-form').value = val;
    document.getElementById('rename-input-box-form').select();
}

/*
 * setRenameHiddenFormValue initializes the hidden rename input form with val 
 * */
function setRenameHiddenFormValue(val) {
    document.getElementById('rename-hidden-input-box-form').value = val;
}

/*
 * overItem changes the look of the item when the mouse is over
 * */
function overItem(ev) {

    ev.preventDefault();
    addClass(ev.target, CLASS_ITEM_OVER);

}

/*
 * leaveItem changes the look of the item when the mouse leaves
 * */
function leaveItem(ev) {

    ev.preventDefault();
    removeClass(ev.target, CLASS_ITEM_OVER);

}

/*
 * hasClass returns true if element has the class clname
 * */
function hasClass(element, clname) {

    return (' ' + element.className + ' ').indexOf(' ' + clname + ' ') > -1;

}

/*
 * addClass adds the class clname to the given element
 * */
function addClass(element, clname) {

    //match = '(?:^|\s)' + clname + '(?!\S)'
    match = clname;
    re = new RegExp(match, 'g');

    if ( ! element.className.match(re)) {
        element.className += ' ' + clname;
    }

}

/*
 * removeClass remove the class clname from the given element
 * */
function removeClass(element, clname) {
    
    //match = '(?:^|\s)' + clname + '(?!\S)'
    match = clname;
    re = new RegExp(match, 'g');

    element.className = element.className.replace( re , '' );

}

/*
 * open the url in the parent window
 * */
function openInParent(url) {

    //window.opener.location.href = url;
    parent.window.open(url);

}

/*
 * undisplayChildrenFolders hiddes the children of the folder folderIdNumber
 * */
function undisplayChildrenFolders(folderIdNumber) {

    console.log("undisplayChildrenFolders:folderIdNumber=" + folderIdNumber);

    document.getElementById("subfolders-" + folderIdNumber).innerHTML = '';
    document.getElementById("folder-" + folderIdNumber).className = CLASS_ITEM_FOLDER + " " + CLASS_ITEM_FOLDER_CLOSED;

}

/*
 * hasChildrenFolders returns true if the folder with folderIdNumber
 * has children in the current HTML page (no DB stuff here)
 * */
function hasChildrenFolders(folderIdNumber) {

    console.log("hasChildrenFolders:folderIdNumber=" + folderIdNumber);

    folder = document.getElementById("folder-" + folderIdNumber);

    return hasClass(folder, CLASS_ITEM_FOLDER_OPEN)

//    // getting the folder children ul
//    subfolders = document.getElementById("subfolders-" + folderIdNumber);
//    
//    isEmpty = subfolders.childNodes.length > 0;
//
//    console.log("hasChildrenFolders:isEmpty=" + isEmpty);
//
//    return isEmpty;

}

/*
 * createBookmark creates a new Bookmark structure
 * */
function createBookmark(bkmId, bkmTitle, bkmURL, bkmFavicon) {

    // bookmark title
    var newATitle = document.createTextNode(bkmTitle);

    // tags attributes
    var attAOnClick = document.createAttribute("onclick");
    var attATitle = document.createAttribute("title");
    var attAId = document.createAttribute("id");
    var attDivId = document.createAttribute("id");
    var attDivClass = document.createAttribute("class");
    var attDivDraggable = document.createAttribute("draggable");
    var attImageSrc = document.createAttribute("src");
    var attImageClass = document.createAttribute("class");

    // tags (the bookmark link is not a A but a clickable DIV)
    var newDivElem = document.createElement("div");
    var newA = document.createElement("div");
    var newImage = document.createElement("img");

    // tags attributes init
    attDivId.value = "bookmark-" + bkmId;
    attDivClass.value = CLASS_ITEM_BOOKMARK;
    attDivDraggable.value = "true";
    attAOnClick.value = "openInParent('" + bkmURL  + "');";
    attATitle.value = bkmURL;
    attAId.value = "bookmark-link-" + bkmId;
    attImageSrc.value = bkmFavicon;
    attImageClass.value = "favicon";

    // tags attributes linking
    newA.setAttributeNode(attAOnClick);
    newA.setAttributeNode(attAId);
    newA.setAttributeNode(attATitle);
    newDivElem.setAttributeNode(attDivId);
    newDivElem.setAttributeNode(attDivClass);
    newDivElem.setAttributeNode(attDivDraggable);
    newImage.setAttributeNode(attImageSrc);
    newImage.setAttributeNode(attImageClass);

    // final structure build
    newA.appendChild(newATitle);
    newDivElem.appendChild(newImage);
    newDivElem.appendChild(newA);
    
    // adding drag and drop events
    newDivElem.addEventListener("dragstart", dragBookmark);

    console.log(newA);

    return newDivElem;

}

/*
 * createFolder creates a new folder structure
 * */
function createFolder(folderId, folderTitle, nbChildrenFolders) {

     // new folder DIV creation
    // <div>
    //   title
    // </div>
    // <ul></ul>
    /* main div attributes */
    var newDivTitle = document.createTextNode(folderTitle);
    var attDivId = document.createAttribute("id");
    var attDivClass = document.createAttribute("class");
    var attDivDraggable = document.createAttribute("draggable");
    var attDivOnclick = document.createAttribute("onclick");

    attDivId.value = "folder-" + folderId;
    attDivClass.value = CLASS_ITEM_FOLDER + " " + CLASS_ITEM_FOLDER_CLOSED;
    attDivDraggable.value = "true";
    attDivOnclick.value = "getChildrenFolders(event, " + folderId  + ");";

    /* ul attributes */
    var attrUlId = document.createAttribute("id");
    attrUlId.value = "subfolders-" + folderId;

    /* div and ul creation  */
    var newFolder = document.createElement("div");
    var newUl = document.createElement("ul");

    /* div attributes linking */
    newFolder.setAttributeNode(attDivId);
    newFolder.setAttributeNode(attDivClass);
    newFolder.setAttributeNode(attDivDraggable);
    newFolder.setAttributeNode(attDivOnclick);
    
    /* ul attributes linking */
    newUl.setAttributeNode(attrUlId);

    /* final structure building  */
    newFolder.appendChild(newDivTitle);

    // adding drag and drop events
    newFolder.addEventListener("drop", dropFolder);
    newFolder.addEventListener("dragover", overFolder);
    newFolder.addEventListener("dragstart", dragFolder);
    newFolder.addEventListener("dragleave", leaveFolder);

    return [newFolder, newUl];
}

/*
 * displaySubfolder displays a folder struct with the given folderId and folderTitle
 * as a children of parentFolderId
 * */
function displaySubfolder(parentFolderId, folderId, folderTitle, nbChildrenFolders) {

    console.log("parentFolderId:" + parentFolderId);
    console.log("folderId:" + folderId);
    console.log("folderTitle:" + folderTitle);

    // checking if the folder does not exist
    if(document.getElementById('folder-' + folderId)) {
        return;
    }

    newFolderAndUl = createFolder(folderId, folderTitle, nbChildrenFolders);
  
    // then adding the new DIV to the DOM
    // and the new UL to contains the children
    document.getElementById("subfolders-" + parentFolderId).appendChild(newFolderAndUl[0]);
    document.getElementById("subfolders-" + parentFolderId).appendChild(newFolderAndUl[1]);

}

/*
 * displayBookmark displays a bookmark struct with the given bkmId, bkmTitle and bkmURL
 * as a children of parentFolderId
 * */
function displayBookmark(parentFolderId, bkmId, bkmTitle, bkmURL, bkmFavicon) {

    console.log("parentFolderId:" + parentFolderId);
    console.log("bkmId:" + bkmId);
    console.log("bkmTitle:" + bkmTitle);

    // checking if the bookmark does not exist
    if(document.getElementById('bookmark-' + bkmId)) {
        return;
    }

    newBookmark = createBookmark(bkmId, bkmTitle, bkmURL, bkmFavicon);
 
    // then adding the new DIV to the DOM
    // and the new UL to contains the children
    document.getElementById("subfolders-" + parentFolderId).appendChild(newBookmark);

}

/*-------------------
 *
 * Callback functions
 *
 * ------------------*/

/*
 * addFolder adds a root folder with the name of the input#addfolder value
 */
 function addFolder() {

    // getting the folder name from the form
    var folderName = document.getElementById("add-folder").value

    console.log("addFolder:folderName=" + folderName);

    var requestAddFolder = new XMLHttpRequest();

    requestAddFolder.open('GET', encodeURI(GoBkmProxyURL + "/addFolder/?folderName=" + folderName + "&t=" + Math.random()), true);

    requestAddFolder.onreadystatechange = function() {

      if (requestAddFolder.readyState == 4 && requestAddFolder.status == 200) {

            if ( requestAddFolder.responseText.length == 0) { return };

            // Success!
            var data = JSON.parse(requestAddFolder.responseText);

            console.log("addFolder:FolderId=" + data.FolderId);
            console.log("addFolder:FolderTitle=" + data.FolderTitle);

            // creating a new folder DIV
            newFolderStruct = createFolder(data.FolderId, data.FolderTitle) ;

            // appending the new folder to the root
            document.getElementById("subfolders-1").appendChild(newFolderStruct[0]);
            document.getElementById("subfolders-1").appendChild(newFolderStruct[1]);
            document.getElementById("add-folder").value = "";
      
      } else if (requestAddFolder.status != 200) {

        // We reached our target server, but it returned an error
        alert("Oups, an error occured ! (addFolder)");

      }

    };

    requestAddFolder.onerror = function() {
      // There was a connection error of some sort
    };

    requestAddFolder.send();

}

/*
 * renameFolder rename a folder with the name of the input#rename-folder value
 */
function renameFolder() {

    // getting the folder id and name from the forms
    folderId = document.getElementById('rename-hidden-input-box-form').value;
    folderName = document.getElementById('rename-input-box-form').value;

    // extracting the digit from the id
    _folderIdNumber = folderId.split("-")[1];

    console.log("renameFolder:folderId=" + folderId);
    console.log("renameFolder:folderName=" + folderName);

    if (folderId.startsWith("folder")) {

        var requestRenameFolder = new XMLHttpRequest();

        requestRenameFolder.open('GET', encodeURI(GoBkmProxyURL + "/renameFolder/?folderId=" + _folderIdNumber + "&folderName=" + folderName + "&t=" + Math.random()), true);

        requestRenameFolder.onreadystatechange = function() {

          if (requestRenameFolder.status >= 200 && requestRenameFolder.status < 400) {

            // Success!
            document.getElementById(draggedFolderId).innerHTML = folderName;

          } else {

            // We reached our target server, but it returned an error
            alert("Oups, an error occured ! (renameFolder)");

          }
        };

        requestRenameFolder.onerror = function() {
          // There was a connection error of some sort
        };

        requestRenameFolder.send();

    } else {

        var requestRenameBookmark = new XMLHttpRequest();

        requestRenameBookmark.open('GET', encodeURI(GoBkmProxyURL + "/renameBookmark/?bookmarkId=" + _folderIdNumber + "&bookmarkName=" + folderName + "&t=" + Math.random()), true);

        requestRenameBookmark.onreadystatechange = function() {

          if (requestRenameBookmark.status >= 200 && requestRenameBookmark.status < 400) {

            // Success!
            document.getElementById("bookmark-link-" + _folderIdNumber).innerHTML = folderName;

          } else {

            // We reached our target server, but it returned an error
            alert("Oups, an error occured ! (renameBookmark)");

          }
        };

        requestRenameBookmark.onerror = function() {
          // There was a connection error of some sort
        };

        requestRenameBookmark.send();

    }

        // hidding the rename box and clearing its values
        hideRenameBox();

}

/*
 * getChildrenFolders is called when a click is performed on a folder
 * and retrieves from the server the subfolders of the clicked folder
 * */
function getChildrenFolders(ev, folderIdNum) {

    console.log("getChildrenFolders:folderIdNum=" + folderIdNum);

    if (hasChildrenFolders(folderIdNum)) {

        undisplayChildrenFolders(folderIdNum);
        console.log("folder " + folderIdNum + " has children folders");
        return;

    }

    /*
     * getting the children folders from the server
     * */
    var requestChildrenFolders = new XMLHttpRequest();

    requestChildrenFolders.open('GET', encodeURI(GoBkmProxyURL + "/getChildrenFolders/?folderId=" + folderIdNum + "&t=" + Math.random()), true);

    requestChildrenFolders.onreadystatechange = function() {

      if (requestChildrenFolders.status >= 200 && requestChildrenFolders.status < 400) {

        if ( requestChildrenFolders.responseText.length == 0) { return };

        // Success!
        var data = JSON.parse(requestChildrenFolders.responseText);

        for (var i = 0, len = data.length; i < len; i++) {

            var fld = data[i];
            console.log(fld.Title + " " + fld.NbChildrenFolders);
            displaySubfolder(folderIdNum, fld.Id, fld.Title, fld.NbChildrenFolders);

        }

        document.getElementById("folder-" + folderIdNum).className = CLASS_ITEM_FOLDER + " " + CLASS_ITEM_FOLDER_OPEN;

      } else {

        // We reached our target server, but it returned an error
        alert("Oups, an error occured ! (getChildrenFolders)");

      }
    };

    requestChildrenFolders.onerror = function() {
      // There was a connection error of some sort
    };

    requestChildrenFolders.send();
   
    /*
     * getting the folder bookmarks from the server
     * */
    var requestFolderBookmarks = new XMLHttpRequest();

    requestFolderBookmarks.open('GET', encodeURI(GoBkmProxyURL + "/getFolderBookmarks/?folderId=" + folderIdNum + "&t=" + Math.random()), true);

    requestFolderBookmarks.onreadystatechange = function() {

      if (requestFolderBookmarks.status >= 200 && requestFolderBookmarks.status < 400) {

        if ( requestFolderBookmarks.responseText.length == 0) { return };

        // Success!
        var data = JSON.parse(requestFolderBookmarks.response);

        for (var i = 0, len = data.length; i < len; i++) {

            var bkm = data[i];
            console.log("getChildrenFolders:bkm.Title=" + bkm.Title);
            displayBookmark(folderIdNum, bkm.Id, bkm.Title, bkm.URL, bkm.Favicon);

        }

      } else {
        // We reached our target server, but it returned an error
       
       alert("Oups, an error occured !");

      }
    };

    requestFolderBookmarks.onerror = function() {
      // There was a connection error of some sort
    };

    requestFolderBookmarks.send();

}

/*--------------------------
 *
 * onover on leave fonctions
 *
 *--------------------------*/

// the dragged item
var dragItem;

function leaveFolder(ev) {
    leaveItem(ev);
}

function leaveRename(ev) {
    leaveItem(ev);
}

function leaveDelete(ev) {
    leaveItem(ev);
}

function overDelete(ev) {
    overItem(ev);
}

function overFolder(ev) {
    overItem(ev);
}

function overRename(ev) {
    overItem(ev);
}

/*--------------------------
 *
 * ondrag fonctions
 *
 *--------------------------*/

function createDragIcon() {

    var dragIcon = document.createElement('img');

    dragIcon.src = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABQAAAAUCAYAAACNiR0NAAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAAN1wAADdcBQiibeAAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAAAUdEVYdFRpdGxlAFBsYWluIEFycm93cyAxmwo0GAAAAAt0RVh0QXV0aG9yAFNvZWJZ2vkGAAAAIXRFWHRDcmVhdGlvbiBUaW1lADIwMDYtMTItMjZUMDA6MDA6MDBAP0jiAAAANnRFWHRTb3VyY2UAaHR0cHM6Ly9vcGVuY2xpcGFydC5vcmcvZGV0YWlsLzI2NTE3Ly1ieS0tMjY1MTdRQYl9AAAAWHRFWHRDb3B5cmlnaHQAQ0MwIFB1YmxpYyBEb21haW4gRGVkaWNhdGlvbiBodHRwOi8vY3JlYXRpdmVjb21tb25zLm9yZy9wdWJsaWNkb21haW4vemVyby8xLjAvxuO9+QAAAilJREFUOI2tlL9rGmEYxx8v0RK8oaTpeRVeDhEJRhw6yJ2HP87T6FYczhYJ59Clq0MWq6EZOuhU7NxFSpXmL4iLWUQr3BLwR0tbioo/DoWCW2PjdWiFEKp3TfOdXt7n+37e7wPP+wJokM/nO/T5fIdavJgGj97tdh+4XK4DADBoga4Vy7LpwWBwORqNFn6///n/JjTwPP/YbDZjJEnqWJZ9AgB3bpyO47hjSZLmoiieiaJ4JknSnOO445vytpxO51ee51/A704wj8eTdjgcnwFg659pRqPRhBCyXt9HCFmNRqNJM4iiKDtN06cIofgqD0IoTtP0KUVR9uu1jeXCYrHsRiKRV4lE4qUsy8NGo5FaBZzNZk2GYeLJZDKNYZhHUZRPk8lkvEz0UBCEk2KxOF0sFko2m/2ys7PzQK0Tk8lE5HK5j4qiKOVy+XssFitbrVZ2kyTJZDgcdguCcE+n00Gr1epNp9OxGlCW5Umz2RwCwG4gELjb7Xb3er3eMwAAwHGcEAQhn8/nv3U6nZ/BYPBIDej1eo/a7fZFoVDoCoLw1mw2o7/57odCodcMw3QdDsfeKpjNZrMzDNOLRqNvcBwn1C4HANgmCGJ/1dgQBLEPANtaQFd1u4MNcPtPDwDAkMlkzpU/SqVS56DyOaj9NheVSuVkOBwuxuOxUqvV3gPAj3UHNtYVAQD6/f4HvV7/qFqtyqVS6SkAXK7zb6oBAWBer9ffLROrmX8B333LVRLDopMAAAAASUVORK5CYII=';

    return dragIcon;

}

/*
 * dragBookmark is called when we start dragging a bookmark
 * */
function dragBookmark(ev) {

    console.log('dragBookmark');

    dragItem = ev.target;
    
    ev.dataTransfer.setData("dragItemId", ev.target.id);

    dragIcon = createDragIcon();
    ev.dataTransfer.setDragImage(dragIcon, -10, -10);

}

/*
 * dragFolder is called when we start dragging a folder
 * */

function dragFolder(ev) {

    console.log('dragFolder');

    dragItem = ev.target;
    
    ev.dataTransfer.setData("dragItemId", ev.target.id);

    dragIcon = createDragIcon();
    ev.dataTransfer.setDragImage(dragIcon, -10, -10);

}


/*--------------------------
 *
 * ondrop fonctions
 *
 *--------------------------*/

/*
 * dropRename is called when a folder or bookmark is dropped on the rename box
 * */
function dropRename(ev) {

    draggedfolder = ev.target;
    
    renameBox = document.getElementById('rename-box');

    draggedItemId = ev.dataTransfer.getData("dragItemId");
    _draggedItemIdNumber = draggedItemId.split("-")[1];

     if (draggedItemId.startsWith("folder")) {
        
        draggedFolderId = draggedItemId;
        draggedFolderName = document.getElementById(draggedFolderId).innerHTML.trim();

        console.log("dropRename:draggedFolderId=" + draggedFolderId);
        console.log("dropRename:draggedFolderName=" + draggedFolderName);

        showRenameBox();

        setRenameFormValue(draggedFolderName);
        
        setRenameHiddenFormValue(draggedFolderId);

        removeClass(renameBox, CLASS_ITEM_OVER);

     } else {
     
        draggedBookmarkId = draggedItemId;
        draggedBookmarkName = document.getElementById("bookmark-link-" + _draggedItemIdNumber).innerHTML.trim();
       
        showRenameBox();

        setRenameFormValue(draggedBookmarkName);
        
        setRenameHiddenFormValue(draggedBookmarkId);

        removeClass(renameBox, CLASS_ITEM_OVER);

     }
}

/*
 * dropDelete is called when a folder or bookmark is dropped on the delete box
 * */
function dropDelete(ev) {

    console.log('dropFolder');

    ev.preventDefault();

    // retrieving elements
    draggedItemId = ev.dataTransfer.getData("dragItemId");
 
    deleteBox = document.getElementById('delete-box');

    console.log("dropFolder:draggedItemId=" + draggedItemId);

    if (draggedItemId.startsWith("folder")) {

        console.log("folder dropped delete");

        draggedFolderId = draggedItemId;
        draggedFolder = document.getElementById(draggedFolderId);
        _draggedFolderIdNumber = draggedFolderId.split("-")[1];

        var requestDeleteFolder = new XMLHttpRequest();

        requestDeleteFolder.open('GET', encodeURI(GoBkmProxyURL + "/deleteFolder/?folderId=" + _draggedFolderIdNumber + "&t=" + Math.random()), true);

        requestDeleteFolder.onreadystatechange = function() {

          if (requestDeleteFolder.status >= 200 && requestDeleteFolder.status < 400) {

                _children = document.getElementById("subfolders-" + _draggedFolderIdNumber);
                _children.parentNode.removeChild(_children);
                draggedFolder.parentNode.removeChild(draggedFolder);

          } else {

            // We reached our target server, but it returned an error
            alert("Oups, an error occured ! (getChildrenFolders)");

          }
        };

        requestDeleteFolder.onerror = function() {
          // There was a connection error of some sort
        };

        requestDeleteFolder.send();

    }
    else {

        console.log("bookmark dropped delete");
       
        draggedBookmarkId = draggedItemId;
        draggedBookmark = document.getElementById(draggedBookmarkId);
        _draggedBookmarkIdNumber = draggedBookmarkId.split("-")[1];

        var requestDeleteBookmark = new XMLHttpRequest();

        requestDeleteBookmark.open('GET', encodeURI(GoBkmProxyURL + "/deleteBookmark/?bookmarkId=" + _draggedBookmarkIdNumber + "&t=" + Math.random()), true);

        requestDeleteBookmark.onreadystatechange = function() {

          if (requestDeleteBookmark.status >= 200 && requestDeleteBookmark.status < 400) {

              draggedBookmark.parentNode.removeChild(draggedBookmark);

          } else {

            // We reached our target server, but it returned an error
            alert("Oups, an error occured ! (getChildrenFolders)");

          }
        };

        requestDeleteBookmark.onerror = function() {
          // There was a connection error of some sort
        };

        requestDeleteBookmark.send();

    }
    
    removeClass(deleteBox, CLASS_ITEM_OVER);
}

// called when draggable is dropped on droppable
function dropFolder(ev) {

    console.log('dropFolder');

    ev.preventDefault();

    // retrieving elements
    droppedFolder = ev.target;
    droppedFolderId = droppedFolder.getAttribute("id"); 
    draggedItemId = ev.dataTransfer.getData("dragItemId");
 
    console.log("dropFolder:draggedItemId=" + draggedItemId);

    if (draggedItemId.startsWith("folder")) {
        console.log("folder dropped");

        draggedFolderId = draggedItemId;
        draggedFolder = document.getElementById(draggedFolderId);     

        // extracting the drag and drop folders id numbers
        _draggedFolderIdNumber = draggedFolderId.split("-")[1];
        _droppedFolderIdNumber = droppedFolderId.split("-")[1];

        // can not move a folder into itself
        if (_draggedFolderIdNumber == _droppedFolderIdNumber) {
            console.log("can not move a folder into itself");
            return;
        }

        // can not move a folder into its first parent
        draggedParentChildrenUlId = getClosest(draggedFolder, 'ul').getAttribute('id');
        if (draggedParentChildrenUlId == "subfolders-" + _droppedFolderIdNumber) {
            console.log("can not move a folder into its first parent");
            return;
        }

        // can not move a folder into one of its children
        if (droppedFolder.closest('ul#subfolders-' + _draggedFolderIdNumber)) {
            console.log("can not move a folder into one of its children");
            return;
        }

        // then getting the drag and drop folders children (ul elements)
        draggedFolderChildren = document.getElementById("subfolders-" + _draggedFolderIdNumber);
        droppedFolderChildren = document.getElementById("subfolders-" + _droppedFolderIdNumber);

        console.log(draggedFolderChildren);
        console.log(droppedFolderChildren);

        var requestMoveFolder = new XMLHttpRequest();

        requestMoveFolder.open('GET', encodeURI(GoBkmProxyURL + "/moveFolder/?sourceFolderId=" + _draggedFolderIdNumber + "&destinationFolderId=" + _droppedFolderIdNumber  + "&t=" + Math.random()), true);

        requestMoveFolder.onreadystatechange = function() {

          if (requestMoveFolder.status >= 200 && requestMoveFolder.status < 400) {

               // moving the dragged folder and its children 
               droppedFolderChildren.appendChild(draggedFolder);
               droppedFolderChildren.appendChild(draggedFolderChildren);

               removeClass(droppedFolder, CLASS_ITEM_OVER);
               removeClass(droppedFolder, CLASS_ITEM_FOLDER_CLOSED);
               addClass(droppedFolder, CLASS_ITEM_FOLDER_OPEN);


          } else {

            // We reached our target server, but it returned an error
            alert("Oups, an error occured ! (dropFolder)");

          }
        };

        requestMoveFolder.onerror = function() {
          // There was a connection error of some sort
        };

        requestMoveFolder.send();
 
    }
    else if (draggedItemId.startsWith("bookmark")) {
    
        console.log("bookmark dropped");

        draggedBookmarkId = draggedItemId;
        draggedBookmark = document.getElementById(draggedBookmarkId);     

        // extracting the drag and drop folders id numbers
        _draggedBookmarkIdNumber = draggedBookmarkId.split("-")[1];
        _droppedFolderIdNumber = droppedFolderId.split("-")[1];

        // can not move a bookmark into its first parent
        draggedParentChildrenUlId = getClosest(draggedBookmark, 'ul').getAttribute('id');
        if (draggedParentChildrenUlId == "subfolders-" + _droppedFolderIdNumber) {
            console.log("can not move a bookmark into its first parent");
            return;
        }

        droppedFolderChildren = document.getElementById("subfolders-" + _droppedFolderIdNumber);

        var requestMoveBookmark = new XMLHttpRequest();

        requestMoveBookmark.open('GET', encodeURI(GoBkmProxyURL + "/moveBookmark/?bookmarkId=" + _draggedBookmarkIdNumber + "&destinationFolderId=" + _droppedFolderIdNumber + "&t=" + Math.random()), true);

        requestMoveBookmark.onreadystatechange = function() {

          if (requestMoveBookmark.status >= 200 && requestMoveBookmark.status < 400) {

               // moving the dragged folder and its children 
               if (hasChildrenFolders(_droppedFolderIdNumber)) {           

                console.log('folder open')

                droppedFolderChildren.appendChild(draggedBookmark);
                removeClass(droppedFolder, CLASS_ITEM_FOLDER_CLOSED);
                addClass(droppedFolder, CLASS_ITEM_FOLDER_OPEN);

               }
               else {
               
                console.log('folder closed')

                draggedBookmark.parentNode.removeChild(draggedBookmark)
                
               }

               removeClass(droppedFolder, CLASS_ITEM_OVER);

          } else {

            // We reached our target server, but it returned an error
            alert("Oups, an error occured ! (dropFolder)");

          }
        };

        requestMoveBookmark.onerror = function() {
          // There was a connection error of some sort
        };

        requestMoveBookmark.send();

    }
    else {
    
        console.log("new bookmark dropped");
        
        _droppedFolderIdNumber = droppedFolderId.split("-")[1];

        droppedFolderChildren = document.getElementById("subfolders-" + _droppedFolderIdNumber);
 
        url = ev.dataTransfer.getData('URL');
        console.log("dropFolder:url=" + url);
        console.log("dropFolder:_droppedFolderIdNumber=" + _droppedFolderIdNumber);

        var requestAddBookmark = new XMLHttpRequest();

        requestAddBookmark.open('GET', encodeURI(GoBkmProxyURL + "/addBookmark/?bookmarkUrl=" + url + "&destinationFolderId=" + _droppedFolderIdNumber + "&t=" + Math.random()), true);

        requestAddBookmark.onreadystatechange = function() {

            if (requestAddBookmark.readyState == 4 && requestAddBookmark.status == 200) {

               var data = JSON.parse(requestAddBookmark.responseText);

               console.log(data.BookmarkId);
               console.log(data.BookmarkTitle);

               newBookmark = createBookmark(data.BookmarkId, data.BookmarkURL, data.BookmarkURL, '');
               droppedFolderChildren.appendChild(newBookmark);

               removeClass(droppedFolder, CLASS_ITEM_OVER);
               removeClass(droppedFolder, CLASS_ITEM_FOLDER_CLOSED);
               addClass(droppedFolder, CLASS_ITEM_FOLDER_OPEN);
           
            } else if (requestAddBookmark.status != 200) {

            // We reached our target server, but it returned an error
            alert("Oups, an error occured ! (addBookmark)");

          }
        };

        requestAddBookmark.onerror = function() {
          // There was a connection error of some sort
        };

        requestAddBookmark.send();

    }

 }

