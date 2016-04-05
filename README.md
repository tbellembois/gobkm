# GoBkm

GoBkm is an *ultra minimalist single user online bookmark manager* inspired by <http://sitebar.org/> written in [Go](https://golang.org/) and Javascript.  
It is designed to run on a remote server (I run it on a [RaspberryPi](https://www.raspberrypi.org/)) and accessed remotely.

The purpose of this project was to study the Go programming language in its different aspects (OOP, databases, HTML templates, learning curve).

GoBkm uses [Google](http://www.google.com) to retrieve sites favicon, the [FontAwesome](https://fontawesome.github.io/Font-Awesome/) library for folders and bookmarks icons and the [OpenClipart](https://openclipart.org/) library for the drag ghost icon.

![screenshot](screenshot.png)

## Usage

```bash
    ./gobkm # run GoBkm on localhost:8080
```

You can change the listening port with:
```bash
    ./gobkm -port [port_number]
```

Using an HTTP proxy (Apache/Nginx), specify its URL with:
```bash
    ./gobkm -port [port_number] -proxy [proxy_url]
```
## GUI

Drag and drop an URL from your Web browser address bar into a folder to bookmark it. Rename/delete folders and bookmarks by dragging them on the icons on the top.

## Bookmarklet

Drag the bookmarklet on your bookmark bar.

Click it to open the GoBkm bar, resize and place it wherever you want.

## Nginx proxy with user authentication

- create a `gobkm` user and group, and a home for the app

    ```bash
        groupadd --system gobkm
        useradd --system gobkm --gid gobkm
        mkdir /usr/local/gobkm
    ```

- drop the bkm binary into the `/usr/local/gobkm` directory

- setup permissions

    ```bash
        chown -R  gobkm:gobkm /usr/local/gobkm
        cd /usr/local/gobkm
    ```

- launch GoBkm

    ```bash
        cd /usr/local/gobkm
        su - gobkm -c "/usr/local/gobkm/gobkm -proxy http://proxy_url" &
    ```

- setup a Nginx server

    ```bash
    server {

        listen 80;
        # change proxy_url
        server_name proxy_url;
          
        root          /usr/local/gobkm;  
        charset utf-8;
       
        # uncomment to enable authentication
        # details at: http://nginx.org/en/docs/http/ngx_http_auth_basic_module.html
        #auth_basic "GoBkm";
        #auth_basic_user_file /usr/local/gobkm/gobkm.htpasswd;

        location / {
            # change the port if needed
            proxy_pass http://127.0.0.1:8080;
        }

    }
    ```

## Thanks

Thanks to [Sébastien Binet](https://github.com/sbinet) for the tutorial and help.

## Roadmap

- provide a systemd startup script
- translate the Javascript into gopherjs <https://github.com/gopherjs/gopherjs>
- ~~import export feature (HTML?)~~

## Known limitations

- no user management
- no authentication (relies on the HTTP proxy)
- folders and bookmarks are sorted by title (currently not configurable)
- supports only latest versions of Web browsers

## References

- <http://youmightnotneedjquery.com/>
- <http://gomakethings.com/climbing-up-and-down-the-dom-tree-with-vanilla-javascript/>
- <https://golang.org>
- <http://www.w3schools.com/>
