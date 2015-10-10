#!/bin/bash

set -e

function echoerr {
    echo "$@" 1>&2
}

function process_http_result {
    result=$1
    ecode=$2
    if [ "$ecode" != "0" ]
    then
        echoerr "## ERROR, GET failed."
        if [ ! -z $result ]
        then
            error=`echo "$result" | jq -r '.error'`
            if [ "$error" = "null" ]
            then
                echoerr "## No further information."
            else
                echoerr "## ERROR $error"
            fi
        fi
        exit 1
    fi
    echoerr "## SUCCESS"
    echo "$result"
}

function http_get {
    echoerr "# GET $1"
    result=$(curl -s -S -f -u "$token" -X GET $1)
    process_http_result "$result" "$?"
}

function http_post {
    echoerr "# POST $1"
    result=$(curl -s -S -f -u "$token" -X POST -d @- $1)
    process_http_result "$result" "$?"
}

function http_put {
    echoerr "# POST $1"
    result=$(curl -s -S -f -u "$token" -X PUT -d @- $1)
    process_http_result "$result" "$?"
}

function http_delete {
    echoerr "# POST $1"
    result=$(curl -s -S -f -u "$token" -X DELETE $1)
    process_http_result "$result" "$?"
}

function lastarg {
    local expression="$@"
    local x="${expression%?*}"
    local y="${x##*/}"
    echoerr "# OBJECTID=$y"
    echo $y
}

if ! jq --version 2>> /dev/null; then
    echoerr "This tool requires jq, the JSON command line parser. See https://stedolan.github.io/jq/."
    exit 2
fi
