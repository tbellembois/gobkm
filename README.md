# GoBkm

GoBkm is an *ultra minimalist single user online bookmark manager* inspired by <http://sitebar.org/> written in [Go](https://golang.org/) and GopherJS.  
It is designed to run on a remote server (I run it on a [RaspberryPi](https://www.raspberrypi.org/)) and accessed remotely.

The purpose of this project was to study the Go programming language (OOP, databases, HTML templates, learning curve).

![screenshot](screenshot.png)

## Installation

Use the Docker image <https://hub.docker.com/r/tbellembois/gobkm> with the `docker-compose.yml` file in the sources.

or

Download and uncompress the latest release from <https://github.com/tbellembois/gobkm/releases>.

or

```bash
    $ go get -u github.com/tbellembois/gobkm
```

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

Log to file
```bash
    ./gobkm -logfile /var/log/gobkm.log
```

Debug mode (by default only errors are shown):
```bash
    ./gobkm -debug
```

Specify sqlite database file path
```bash
    ./gobkm -db /var/gobkm/gobkm.db
```

## GUI

- drag and drop an URL from your Web browser address bar into a folder OR
- use the "new bookmark" icon in a folder menu 

### Stars

You can "star" your favorite bookmarks to keep them on the top of the window.

### Tags

You can tag bookmarks. This may be redondant with folders but it may help if you have bookmarks with the same topic in different folders.

## Bookmarklets

Click on the little "earth" icon at the bottom of the application and drag and drop the bookmarklet in your bookmark bar.
Searches in the search field are performed by bookmark names and tags.

## Nginx proxy (optional)

### GoBkm installation

You can use Nginx in front of GoBkm to use authentication and HTTPS.

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

### Nginx configuration

- setup a GoBkm server configuration file such as `/etc/nginx/servers-available/gobkm.conf`

    ```bash
    server {

        listen 80;
        # change proxy_url
        server_name proxy_url;
          
        root          /usr/local/gobkm;  
        charset utf-8;

        gzip on;
        gzip_disable "msie6";

        gzip_comp_level 6;
        gzip_min_length 1100;
        gzip_buffers 16 8k;
        gzip_proxied any;
        gzip_types
            text/plain
            text/css
            text/js
            text/xml
            text/javascript
            application/javascript
            application/x-javascript
            application/json
            application/xml
            application/rss+xml
            image/svg+xml;

        # uncomment and change to enable HTTPS
        #ssl on;
        #ssl_certificate /etc/nginx/ssl2/my-gobkm.crt;
        #ssl_certificate_key /etc/nginx/ssl2/my-gobkm.key;

        # uncomment to enable authentication
        # details at: http://nginx.org/en/docs/http/ngx_http_auth_basic_module.html
        #auth_basic "GoBkm";
        #auth_basic_user_file /usr/local/gobkm/gobkm.htpasswd;

        location / {

		  	# preflight OPTIONS requests response
			if ($request_method = 'OPTIONS') {
				add_header 'Access-Control-Allow-Credentials' 'true';
				add_header 'Access-Control-Allow-Origin' '*';
				add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS';
				#
				# Custom headers and headers various browsers *should* be OK with but aren't
				#
				add_header 'Access-Control-Allow-Headers' 'DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type';
				#
				# Tell client that this pre-flight info is valid for 20 days
				#
				add_header 'Access-Control-Max-Age' 1728000;
				add_header 'Content-Type' 'text/plain charset=UTF-8';
				add_header 'Content-Length' 0;
				return 204;
			}

            # change the port if needed
        	proxy_set_header Upgrade $http_upgrade;
        	proxy_set_header Connection 'upgrade';
        	proxy_pass http://127.0.0.1:8080;
        }

    }
    ```

- enable the new site

    ```bash
        $ ln -s /etc/nginx/servers-available/gobkm.conf /etc/nginx/servers-enabled/
        $ systemctl restart nginx
    ```

### SSL self-signed certificate generation (optional)

I strongly recommend [Let's Encrypt](https://letsencrypt.org/) for the certificates generation.

The following method is just here for archive purposes.

```bash
	# generate a root CA key
	openssl genrsa -out rootCA.key 2048
	# the a root CA
	openssl req -x509 -new -nodes -key rootCA.key -days 3650 -out rootCA.crt
	# generate a server key  
    openssl genrsa -out my-gobkm.key 2048
	# then a CSR (certificate signing request)
	openssl req -new -key my-gobkm.key  -out my-gobkm.csr
	# and finally auto signing the server certificate with the root CA
    openssl x509 -req -in my-gobkm.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out my-gobkm.crt -days 3650
```
To avoid security exceptions just import your `rootCA.crt` into your browser.

## systemd script (optional)

If you want to start GoBkm at boot you can install the provided systemd `gobkm.service` script.

```bash
    # example for Arch Linux
    cd /etc/systemd/system
    wget https://raw.githubusercontent.com/tbellembois/gobkm/master/gobkm.service
    vim gobkm.service
    # change the ExecStart line and other parameters according to your configuration
    systemctl enable gobkm.service
```

## Thanks

Thanks to [Sébastien Binet](https://github.com/sbinet) for the tutorial and help on Go.  
Thanks to [Dmitri Shuralyov](https://github.com/shurcooL) for the help on GopherJS.

## Known limitations

- no user management
- no authentication (relies on the HTTP proxy)
- folders and bookmarks are sorted by title (currently not configurable)

## Notes

Cross compiled for the RaspberryPi under Arch Linux with:
```bash
    # requires the package arm-linux-gnueabihf-gcc
    env GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=1 CC=/usr/bin/arm-linux-gnueabihf-gcc go build .
```

Javascript generation (DEPRECATED):
```bash
    gopherjs build static/js/gjs-main.go -o static/js/gjs-main.js
```

## Credits

- sites favicon retrieved from [Google](http://www.google.com)
- folders, bookmarks, rename and delete icons from the [FontAwesome](https://fontawesome.github.io/Font-Awesome/) library
- GoBKM SVG favicon build with [Inkscape](http://www.inkscape-fr.org/) from <https://github.com/golang-samples/gopher-vector> and <https://commons.wikimedia.org/wiki/File:Bookmark_empty_font_awesome.svg>
- favicon PNG generated from <https://realfavicongenerator.net>

## References

- <http://youmightnotneedjquery.com/>
- <http://gomakethings.com/climbing-up-and-down-the-dom-tree-with-vanilla-javascript/>
- <https://golang.org>
- <https://github.com/jteeuwen/go-bindata>
- <https://godoc.org/github.com/GeertJohan/go.rice>
- <http://www.w3schools.com/>
- <http://blog.teamtreehouse.com/uploading-files-ajax>
- <https://joeshaw.org/net-context-and-http-handler/>
- <https://medium.com/@matryer/the-http-handlerfunc-wrapper-technique-in-golang-c60bf76e6124#.xx66llwp4>
- <http://wpvkp.com/font-awesome-doesnt-display-in-firefox-maxcdn/>
