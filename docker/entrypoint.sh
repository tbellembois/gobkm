#!/usr/bin/env bash
proxyurl=""
logfile=""
debug=""

if [ ! -z "$GOBKM_PROXYURL" ]
then
      proxyurl="-proxyurl $GOBKM_PROXYURL"
fi
if [ ! -z "$GOBKM_DEBUG" ]
then
      debug="-debug"
fi
if [ ! -z "$GOBKM_LOGFILE" ]
then
      logfile="-logfile $GOBKM_LOGFILE"
fi

/var/www-data/gobkm -db /data/bkm.db \
$listenport \
$proxyurl \
$logfile