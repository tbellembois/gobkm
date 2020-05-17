#!/usr/bin/env bash
proxyurl=""
logfile=""
debug=""

if [ ! -z "$GOBKM_PROXYURL" ]
then
      proxyurl="-proxy $GOBKM_PROXYURL"
fi
if [ ! -z "$GOBKM_DEBUG" ]
then
      debug="-debug"
fi
if [ ! -z "$GOBKM_LOGFILE" ]
then
      logfile="-logfile $GOBKM_LOGFILE"
fi

echo "proxyurl: $GOBKM_PROXYURL"
echo "logfile: $GOBKM_LOGFILE"
echo "debug: $DEBUG"

/var/www-data/gobkm -db /data/bkm.db \
$debug \
$proxyurl \
$logfile