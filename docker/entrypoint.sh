#!/usr/bin/env bash
proxy=""
logfile=""
debug=""
history=3
username=""

if [ ! -z "$GOBKM_PROXYURL" ]
then
      proxy="-proxy $GOBKM_PROXYURL"
fi
if [ ! -z "$GOBKM_HISTORY" ]
then
      history="-history $GOBKM_HISTORY"
fi
if [ ! -z "$GOBKM_USERNAME" ]
then
      username="-username $GOBKM_USERNAME"
fi
if [ ! -z "$GOBKM_DEBUG" ]
then
      debug="-debug"
fi
if [ ! -z "$GOBKM_LOGFILE" ]
then
      logfile="-logfile $GOBKM_LOGFILE"
fi

echo "proxy: $GOBKM_PROXYURL"
echo "logfile: $GOBKM_LOGFILE"
echo "debug: $GOBKM_DEBUG"
echo "history: $GOBKM_HISTORY"
echo "username: $GOBKM_USERNAME"

/var/www-data/gobkm -db /data/bkm.db \
$debug \
$proxy \
$logfile \
$history \
$username