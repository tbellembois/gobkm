# Version: 0.0.1
FROM golang:1.14.3-buster
LABEL author="Thomas Bellembois"

# copying sources
WORKDIR /go/src/github.com/tbellembois/gobkm/
COPY . .

# installing dependencies
RUN go get -v ./...

# installing GoBkm
RUN mkdir /var/www-data \
    && cp /go/bin/gobkm /var/www-data/ \
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
