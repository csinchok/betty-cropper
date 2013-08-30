#!/bin/bash

NAME=betty-cropper
SCRIPTNAME=/etc/init.d/$NAME
PIDFILE=/var/run/$NAME.pid

if [ -f $PIDFILE ]; then
    PID=`cat $PIDFILE`
    if [ -z "`ps axf | grep ${PID} | grep -v grep`" ]; then
        $SCRIPTNAME start
    else
        $SCRIPTNAME restart
    fi
else
    $SCRIPTNAME start
fi