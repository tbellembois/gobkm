// search the bookmarks with the given tag
function searchTag(name) {
    $("input#search-form-input").val(name);
    $("input#search-form-input").trigger( "keyup" );
}

// unstar the given bookmark
function unStar(id) {
    $.ajax({
        url: "/starBookmark/",
        data: {
            bookmarkId: id,
            star: false
        }
    }).done(function(result) {
        $("li#bookmark-starred-" + id).remove();
    }).always(function() {
    });
}

// display an alert message
function displayMessage(msgText, type) {
    var d = $("<div>");
    d.attr("role", "alert");
    d.addClass("alert alert-" + type);
    d.css("z-index", "99");
    d.text(msgText);
    $("body").prepend(d.delay(1000).fadeOut("slow"));
}

// clear the search results
function clearSearchResults() {
    $("ul#search-list").html("");
    $("input#search-form-input").val("");
}

// toggle the given element
function toggleElement(id) {
    var el = document.getElementById(id);
    el.style.display = el.style.display == "none" ? "block" : "none";
}

// helper function to create a starred bookmark li
function createStarredBookmarkLi(bID, bTitle, bURL, bFavicon) {
    newBookmark = " \
        <li id='bookmark-starred-ID'> \
            <div class='bookmark'> \
                <img src=\"SRC\" alt='' class='favicon'> \
                <span class='bookmark-starred-link'  title=\"URL\" onclick='window.open(\"URL\", \"_blank\");'>TITLE</span> \
                <a href='#' onclick=\"unStar('-ID');\" class='fas fa-star-half' aria-hidden='true' title='remove bookmark'></a> \
            </div> \
        </li> \
    "
    newBookmark = newBookmark.replace("ID", bID);
    newBookmark = newBookmark.replace("ID", bID);
    newBookmark = newBookmark.replace("SRC", bFavicon);
    newBookmark = newBookmark.replace("URL", bURL);
    newBookmark = newBookmark.replace("URL", bURL);
    newBookmark = newBookmark.replace("TITLE", bTitle);
    return newBookmark;
}

// helper function to create a bookmark li (used in search results)
function createBookmarkLi(bID, bTitle, bURL, bFavicon) {
    newBookmark = " \
    <li class='col-sm-12 list-group-item'> \
            <img src='" + bFavicon + "' class='favicon'> \
            <span onclick='window.open(\"" + bURL + "\", \"_blank\");' class='bookmark-link'>" + bTitle + "</span> \
        </a> \
    </li>"
    return newBookmark;
}

//
// Menu callback functions.
//
// Called on create bookmark.
function newbookmarkCallBack(itemKey, opt, rootMenu, originalEvent) {
    tree = $('#tree').tree();
    nodeId = tree.getSelections()[0];
    
    // calling ajax
    $.ajax({
        method: "GET",
        url: "/addBookmark/",
        data: {
            destinationFolderId: nodeId,
            bookmarkUrl: "https://golang.org/",
        },
        dataType: "json",
    }).done(function(result) {
        displayMessage("bookmark created !", "success");
        // add node
        var parent = tree.getNodeById(nodeId);
        tree.expand(parent);
        tree.addNode({ id: -result.id, text: "new bookmark", url: "https://golang.org/", hasChildren: false, lazy: false, icon: "" }, parent);
        // make the new bookmark editable
        newNode = tree.getNodeById(-result.id);
        tree.unselect(parent);
        tree.select(newNode);
        editCallBack();
    }).fail(function() {
        displayMessage("error !", "alert");
    }).always(function() {
    });
}
// Called on create subfolder.
function newsubfolderCallBack(itemKey, opt, rootMenu, originalEvent) {
    tree = $('#tree').tree();
    nodeId = tree.getSelections()[0];
    
    // calling ajax
    $.ajax({
        method: "GET",
        url: "/addFolder/",
        data: {
            parentId: nodeId,
            folderName: "new folder",
        },
        dataType: "json",
    }).done(function(result) {
        displayMessage("folder created !", "success");
        // add node
        var parent = tree.getNodeById(nodeId);
        tree.expand(parent);
        tree.addNode({ id: result.id, text: "new folder", url: "", hasChildren: true, lazy: true, icon: "" }, parent);
        // make the new folder editable
        newNode = tree.getNodeById(result.id);
        tree.unselect(parent);
        tree.select(newNode);
        editCallBack();
    }).fail(function() {
        displayMessage("error !", "alert");
    }).always(function() {
    });
}
// Called on edit.
function editCallBack(itemKey, opt, rootMenu, originalEvent) {
    tree = $('#tree').tree();
    nodeId = tree.getSelections()[0];

    // hide possible open edit div
    $("div[class=edit]").hide();
    // clean possible selected values not committed
    $(".select2").val(null).trigger('change');

    // show edit div
    $("li[data-id=" + nodeId + "]").removeClass("active");
    $("div[edit=" + nodeId + "]").show();
    
    // initialize select2
    $("#select-tag-" + nodeId).select2({
        placeholder: 'select or enter a tag',
        tags: true,
        ajax: {
            url: '/getTags/',
            dataType: 'json',
            processResults: function (data) {
                // replacing name by text expected by select2
                var newdata = $.map(data, function (obj) {
                    obj.text = obj.text || obj.name;
                    obj.id = obj.id || obj.id;
                    return obj;
                });
                return {
                    results: newdata
                };
            },
        },
    });

    // getting the bookmark tags
    $.ajax({
        url: "/getBookmarkTags/",
        data: { bookmarkId: nodeId }
    }).then(function( data) {
        $.each(data, function( index, value ) {
            // create the option and append to Select2
            var option = new Option(value.name, value.id, true, true);
            $("#select-tag-" + nodeId).append(option).trigger('change');

            // manually trigger the `select2:select` event
            $("#select-tag-" + nodeId).trigger({
                type: 'select2:select',
                params: {
                    data: data
                }
            });
        });
    });

}

// Called on delete.
function deleteCallBack(itemKey, opt, rootMenu, originalEvent) {
    tree = $('#tree').tree();
    nodeId = tree.getSelections()[0];
    n = tree.getNodeById(nodeId);      
    isfolder = (n.attr("isfolder") == "true");

    // building ajax url
    url = "/deleteBookmark/";
    if (isfolder) {
        url = "/deleteFolder/";
    }
    // calling ajax
    $.ajax({
        method: "GET",
        url: url,
        data: {
            itemId: nodeId,
        },
        dataType: "json",
    }).done(function() {
        displayMessage("deleted !", "success");
        // remove node
        tree.removeNode(n);
    }).fail(function() {
        displayMessage("error !", "alert");
    }).always(function() {
    });
}
// Called on cut.
function cutCallBack(itemKey, opt, rootMenu, originalEvent) {
    tree = $('#tree').tree();
    nodeId = tree.getSelections()[0];
    // getting the cutted node
    n = tree.getNodeById(nodeId);
    // getting the cutted node data
    data = tree.getDataById(nodeId);
    // getting the isfolder attribute
    isfolder = n.attr("isfolder");
    // setting the hidden input fields
    $('input[name="cutted-nodeid"]').val(nodeId);
    $('input[name="cutted-nodeisfolder"]').val(isfolder);
    // sending a message
    displayMessage("cutted " + data.text, "success");
}
// Called on paste.
function pasteCallBack(itemKey, opt, rootMenu, originalEvent) {
    tree = $('#tree').tree();
    nodeId = tree.getSelections()[0];
    // getting the node to paste from the hidden input fields
    topasteNodeId = $('input[name="cutted-nodeid"]').val();
    topasteIsFolder = $('input[name="cutted-nodeisfolder"]').val();
    // getting the pasted and destination nodes
    dn = tree.getNodeById(nodeId);
    pn = tree.getNodeById(topasteNodeId);
    // getting the pasted and destination nodes data
    dndata = tree.getDataById(nodeId);
    pndata = tree.getDataById(topasteNodeId);
    // getting the isfolder attribute
    isfolder = pn.attr("isfolder");           

    // building ajax url
    isfolder = (topasteIsFolder == "true");
    url = "/moveBookmark/";
    if (isfolder) {
        url = "/moveFolder/";
    }

    // calling ajax
    $.ajax({
        method: "GET",
        url: url,
        data: {
            sourceItemId: topasteNodeId,
            destinationItemId: nodeId,
        },
        dataType: "json",
    }).done(function() {
        displayMessage("pasted " + pndata.text + " into " + dndata.text, "success");
        // expand destination node then move node
        tree.expand(dn);
        tree.addNode(tree.getDataById(topasteNodeId), dn);
        tree.removeNode(pn);
        // delete hidden input fields values
        $('input[name="cutted-nodeid"]').val("");
        $('input[name="cutted-nodeisfolder"]').val("");
    }).fail(function() {
        displayMessage("error !", "alert");
    }).always(function() {
    });
}
// Called on star.
function starCallBack(itemKey, opt, rootMenu, originalEvent) {
    tree = $('#tree').tree();
    nodeId = tree.getSelections()[0];

    // calling ajax
    $.ajax({
        method: "GET",
        url: "/starBookmark/",
        data: {
            bookmarkId: nodeId,
            star: true,
        },
        dataType: "json",
    }).done(function(result) {
        // add new starred bookmark
        newID = result.Id
        newFavicon = result.Favicon
        newURL = result.URL
        newTitle = result.Title
        $("ul#starred").append(createStarredBookmarkLi(newID, newTitle, newURL, newFavicon));
        displayMessage("stared " + newTitle, "success");
    }).fail(function() {
        displayMessage("error !", "alert");
    }).always(function() {
    });
}

//
// Import callback
//
function importCallback(){

    var file = $("#import-file").get(0).files[0],
    formData = new FormData();
    formData.append( 'file', file );

    // calling ajax
    $.ajax({
        method: "POST",
        url: "/import/",
        contentType: false,
        cache      : false,
        processData: false,
        data       : formData,
    }).done(function(result) {
        displayMessage("import success", "success");
        window.setTimeout(function(){location.reload()},2000)
    }).fail(function() {
        displayMessage("error !", "alert");
    }).always(function() {
    });
}

//
// Bookmark edition callback.
//
function editOkCallBack(nodeId) {
    //console.log(params);
    tree = $('#tree').tree();
    isfolder = $("input#input-fld-" + nodeId).val();
    ftitle = $("input#input-title-" + nodeId).val();
    furl = $("input#input-url-" + nodeId).val();
    ftag = $("select#select-tag-" + nodeId).select2('data');
    fid = $("input#input-id-" + nodeId).val();

    url = "/renameBookmark/";
    if (isfolder == "true") {
        url = "/renameFolder/";
    }
    selectedIds = [];
    $.each(ftag, function( index, value ) {
        selectedIds.push(value.id);
    });
    console.log(selectedIds);
    $.ajax({
        method: "GET",
        url: url,
        data: {
            itemId: fid,
            itemName: ftitle,
            itemUrl: furl,
            itemTag: selectedIds,
        },
        dataType: "json",
    }).done(function(data) {
        displayMessage("updated " + ftitle, "success");
        $("li[data-id=" + fid + "]").find("span[data-role=display]").first().text(ftitle);
        $("li[data-id=" + fid + "]").find("a[link-id=" + fid + "]").first().attr("href", furl).attr("title", furl);       
    }).fail(function(data) {
        displayMessage("error !", "alert");
    }).always(function(data) {
        $("div[edit=" + nodeId + "]").hide();
    });
}

//
// Drop zone callback functions
//
// ondragover
function dzAllowDrop(ev) {
    ev.preventDefault();
}
// ondrop
function dzDrop(ev) {
    ev.preventDefault();
    
    // getting the url from dropped bookmark
    var data
    dttypes = ev.dataTransfer.types;
    if ($.inArray("text/uri-list", dttypes) != -1) {
        data = ev.dataTransfer.getData("text/uri-list");
    } else if($.inArray("text/x-moz-url", dttypes) != -1) {
        data = ev.dataTransfer.getData("text/x-moz-url");
    } else {
        alert("not supported by your browser");
        return;
    }
    //console.log(data);
    //console.log(ev.dataTransfer.types);
    // ["text/plain", "text/uri-list"]
    // ["text/plain", "text/x-moz-url"

    // getting the selected node id
    tree = $('#tree').tree();
    nodeId = tree.getSelections()[0];
    if (typeof nodeId == 'undefined') {
        nodeId = "1";
    }
    console.log(nodeId);

    // getting the destination node
    dn = tree.getNodeById(nodeId);

    // calling ajax
    $.ajax({
        url: "/addBookmark/",
        data: {
            destinationFolderId: nodeId,
            bookmarkUrl: data
        }
    }).done(function(result) {
        displayMessage("added !", "success");
        // expanded the destination node
        tree.expand(dn);
        // adding the new node
        tree.addNode({ id: -result.id, text: result.text, url: result.url, hasChildren: false, lazy: false, icon: result.icon }, dn);
    }).always(function() {
    });

}

//
// Tree callback functions.
//
// Called when a node is rendered.
function nodeDataBoundCallback (e, node, id, record) {
    // Adding the editing hidden form
    var eledit = $("<div></div>").attr("edit", id).addClass("edit");
    var eltitle = $("<div></div>").addClass("input-group mb-3").attr("id", id).append(
        $("<div></div>").addClass("input-group-prepend").append(
            $("<span></span>").addClass("input-group-text").attr("id", "edit-title-" + id).text("title")
        )).append(
            $("<input></input>").addClass("form-control").attr("id", "input-title-" + id).attr("aria-describedby", "edit-title-" + id).val(record.text)
        );
    var elurl = $("<div></div>").addClass("input-group mb-3").attr("id", id).append(
        $("<div></div>").addClass("input-group-prepend").append(
            $("<span></span>").addClass("input-group-text").attr("id", "edit-url-" + id).text("url")
        )).append(
            $("<input></input>").addClass("form-control").attr("id", "input-url-" + id).attr("aria-describedby", "edit-url-" + id).val(record.url)
        );
    var eltag =  $("<div></div>").addClass("input-group mb-3").attr("id", id).append(
        $("<div></div>").addClass("input-group-prepend").append(
            $("<span></span>").addClass("input-group-text").attr("id", "edit-tag-" + id).text("tags")
        )).append(
            $("<select></select>").addClass("form-control").addClass("select2").attr("multiple", "multiple").attr("id", "select-tag-" + id).attr("aria-describedby", "edit-tag-" + id).val(record.tag)
        );
    var elid = $("<input></input>").attr("type", "hidden").attr("id", "input-id-" + id).val(id);
    var elbuttonok = $("<button></button>").attr("onclick", "editOkCallBack('" + id + "');").attr("type", "button").addClass("btn btn-success").text("ok");
    var elbuttoncancel = $("<button></button>").attr("onclick", "$('div[edit=" + id + "]').hide();").attr("type", "button").addClass("btn btn-secondary").text("cancel");
    
    if (record.url != "") {
        var elfld = $("<input></input>").attr("type", "hidden").attr("id", "input-fld-" + id).val(false);
        eledit.append(elid).append(eltitle).append(elfld).append(elurl).append(eltag).append(elbuttonok).append("&nbsp;").append(elbuttoncancel);
        node.find("div[data-role=wrapper]").first().after(eledit);
    } else {
        var elfld = $("<input></input>").attr("type", "hidden").attr("id", "input-fld-" + id).val(true);
        eledit.append(elid).append(eltitle).append(elfld).append(elbuttonok).append("&nbsp;").append(elbuttoncancel);
        node.find("div[data-role=wrapper]").first().after(eledit)
    }
    eledit.hide();

    // Appending an link with the url for bookmarks.
    if (record.url != "") {
        var d = $("<div></div>").attr("class", "icon-link");
        if (record.url.startsWith("file")) {
            var a = $("<a></a>").attr("link-id", id).attr("href", "#").attr("onclick", "alert('" + record.url + "')").attr("title", record.url);
        } else {
            var a = $("<a></a>").attr("link-id", id).attr("href", record.url).attr("target", "_blank").attr("title", record.url);
        }
        var i = $("<i></i>").attr("class", "fas fa-link");
        d.append(a.append(i));
        node.find('div[data-role="wrapper"]').append(d);

        // Adding tags as title for bookmarks
        var title = "";
        $.each(record.tag, function( index, tag ) {
            title = title + " " + tag.name
        });
        node.attr("title", title);
    }
    
    // Appending an attribute to the node to identify folders.
    if (record.url != "") {
        node.attr("isfolder", false);
    } else {
        node.attr("isfolder", true);
    }
};
// Called when a node is enabled
function selectCallback (e, node) {
    // Deleting search results
    clearSearchResults();
    // Disabling editing on all other nodes
    // When cancelling a node edition, the node remains editable
    var currentpk = node.attr("data-id");
    $("span.editable").each( function( index, element ){
        //var pk = $(this).attr("data-pk");
        //if (pk != currentpk) {
        //    $(this).editable("disable");
        //}
    });
};

// when the page is loaded
$(function() {

    // establishing web socket connection
    if (window.location.protocol == "https:") {
        wsproto = "wss";
    } else {
        wsproto = "ws";
    }
    var ws= new WebSocket(wsproto + "://" + GoBkmProxyHost + "/socket/");

    // web socket receive action
    ws.onmessage = function(evt) {
        // lazyli reloading the tree
        // TODO: improve this
        tree = $('#tree').tree();
        tree.reload();
    };

    // retrieving tags
    $.ajax({
        method: "GET",
        url: "/getTags/",
    }).done(function(result) {
        $.each(result, function(id, val) {
            $("div#tags").append('<button type="button" class="btn btn-outline-primary" onclick="searchTag(\'' + val.name + '\')">' + val.name + '</button>');
        });
    }).fail(function() {
        displayMessage("error retrieving tag list !", "alert");
    }).always(function() {
    });

    // import/export button bidding
    $("button#export-box").click(function() {
        window.open('/export/?target=_blank');
    });
    $("button#import-box").click(function() {
        $("div#import-input-box").toggle(); 
    });
    
    // hide/collapse button bidding
    $('.collapse').on('shown.bs.collapse', function () {
        $("#collapse-button").removeClass("fa-angle-double-down").addClass("fa-angle-double-up");
    });
    $('.collapse').on('hidden.bs.collapse', function () {
        $("#collapse-button").removeClass("fa-angle-double-up").addClass("fa-angle-double-down");
    });

    //
    // search input bindding
    //
    // https://schier.co/blog/2014/12/08/wait-for-user-to-stop-typing-using-javascript.html
    // Init a timeout variable to be used below
    var timeout = null;

    // Listen for keystroke events in the search input
    $('#search-form-input').keyup(function (e) {

        search = $(this).val();
        if (search.length < 2) {
            return;
        }

        // Clear the timeout if it has already been set.
        // This will prevent the previous task from executing
        // if it has been less than <MILLISECONDS>
        clearTimeout(timeout);

        // Make a new timeout set to go off in 800ms
        timeout = setTimeout(function () {
            // console.log('Input Value:', searchInput.value);
            $.ajax({
                url: "/searchBookmarks/",
                data: {
                    search: search
                }
            }).done(function(result) {
                if (result == null) {
                    return;
                }
                //console.log(result);
                clearSearchResults();
                // Close button
                //$("div#search-result").append('<div id="close-search-results-button" title="close list" onclick="clearSearchResults();"><i class="fas fa-times-circle"></i></div>');
                $.each(result, function(key,value) {
                    //console.log(value);
                    $("ul#search-list").append(createBookmarkLi(value.Id, value.Title, value.URL, value.Favicon));
                });
            });
        }, 500);
    });

    // context menu initialization
    $.contextMenu({
        selector: '.gj-list-md-active',
        callback: function(key, options) {
            var m = "clicked: " + key;
            window.console && console.log(m) || alert(m); 
        },
        items: {
            "edit": {
                name: "edit",
                icon: function(){
                    return 'fas fa-edit';
                },
                callback: function(itemKey, opt, rootMenu, originalEvent) {
                    editCallBack(itemKey, opt, rootMenu, originalEvent); 
                }
            },
            "sep1": "",
            "newsubfolder": {
                name: "new subfolder",
                icon: function(){
                    return 'far fa-folder-open';
                },
                callback: function(itemKey, opt, rootMenu, originalEvent) {
                    newsubfolderCallBack(itemKey, opt, rootMenu, originalEvent); 
                },
                disabled: function(key, opt){        
                    var nodeId = tree.getSelections()[0];
                    // getting the selected node
                    var n = tree.getNodeById(nodeId);
                    // disabling on bookmarks
                    if (n.attr("isfolder") != "true"){
                        return true;
                    } else {
                        return false;
                    }
                }
            },
            "sep2": "",
            "newbookmark": {
                name: "new bookmark",
                icon: function(){
                    return 'far fa-bookmark';
                },
                callback: function(itemKey, opt, rootMenu, originalEvent) {
                    newbookmarkCallBack(itemKey, opt, rootMenu, originalEvent); 
                },
                disabled: function(key, opt){        
                    var nodeId = tree.getSelections()[0];
                    // getting the selected node
                    var n = tree.getNodeById(nodeId);
                    // disabling on bookmarks
                    if (n.attr("isfolder") != "true"){
                        return true;
                    } else {
                        return false;
                    }
                }
            },
            "sep3": "",
            "cut": {
                name: "cut",
                icon: function(){
                    return 'fas fa-cut';
                },
                callback: function(itemKey, opt, rootMenu, originalEvent) {
                    cutCallBack(itemKey, opt, rootMenu, originalEvent); 
                }
            },
            "sep4": "",
            "paste": {
                name: "paste",
                icon: function(){
                    return 'fas fa-paste';
                },
                callback: function(itemKey, opt, rootMenu, originalEvent) {
                    pasteCallBack(itemKey, opt, rootMenu, originalEvent); 
                },
                disabled: function(key, opt){        
                    // disable paste if no node has been cutted
                    if ($('input[name="cutted-nodeid"]').val() == ""){
                        return true;
                    } else {
                        return false;
                    }
                }
            },
            "sep5": "",
            "delete": {
                name: "delete",
                icon: function(){
                    return 'fas fa-trash-alt';
                },
                callback: function(itemKey, opt, rootMenu, originalEvent) {
                    deleteCallBack(itemKey, opt, rootMenu, originalEvent); 
                },
            },
            "sep6": "",
            "star": {
                name: "star",
                icon: function(){
                    return 'fas fa-star';
                },
                callback: function(itemKey, opt, rootMenu, originalEvent) {
                    starCallBack(itemKey, opt, rootMenu, originalEvent); 
                },
                disabled: function(key, opt){        
                    // disable star on folders
                    var nodeId = tree.getSelections()[0];
                    // getting the selected node
                    var n = tree.getNodeById(nodeId);

                    if (n.attr("isfolder") == "true"){
                        return true;
                    } else {
                        return false;
                    }
                }
            }
        }
    });

    // Create the tree inside the <div id="tree"> element.
    var tree = $('#tree').tree({
        //uiLibrary: 'bootstrap4',
        //iconsLibrary: 'fontawesome',
        uiLibrary: 'materialdesign',
        iconsLibrary: 'materialicons',
        //width: 350,
        border: false,
        primaryKey: 'id',
        //dataSource: '/getBranchNodes/',
        //lazyLoading: true,
        dataSource: '/getTree/',
        lazyLoading: false,
        imageUrlField: 'icon',
    });

    tree.on('nodeDataBound', nodeDataBoundCallback);
    tree.on('select', selectCallback);
    tree.on('expand', clearSearchResults);
});
