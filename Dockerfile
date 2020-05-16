# Version: 0.0.1
FROM golang:1.14.3-buster
LABEL author="Thomas Bellembois"

# copying sources
WORKDIR /go/src/github.com/tbellembois/gobkm/
COPY . .

# installing dependencies
RUN go get -v ./...
RUN go install -v ./...
RUN go get github.com/gopherjs/gopherjs && \
  go get honnef.co/go/js/dom && \
  go get github.com/GeertJohan/go.rice && \
  go get github.com/GeertJohan/go.rice/rice 

# installing Go 1.12 for GopherJS
RUN go get golang.org/dl/go1.12.16
RUN go1.12.16 download

# building
RUN export GOPHERJS_GOROOT="$(go1.12.16 env GOROOT)"; go generate
RUN go build -o gobkm

# installing GoBkm
RUN mkdir /var/www-data \
    && cp /go/src/github.com/tbellembois/gobkm/gobkm /var/www-data/ \
    && chown -R www-data /var/www-data \
    && chmod +x /var/www-data/gobkm

# cleanup sources
RUN rm -Rf /go/src/*

# copying entrypoint
COPY docker/entrypoint.sh /
RUN chmod +x /entrypoint.sh

# creating volume directory
RUN mkdir /data

USER www-data
WORKDIR /var/www-data
ENTRYPOINT [ "/entrypoint.sh" ]
VOLUME ["/data"]
EXPOSE 8081
