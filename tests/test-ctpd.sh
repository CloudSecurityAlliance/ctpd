#!/bin/bash

baseUri="http://localhost:8080/api/1.0"
token="token:0000"

. httptools.sh

result=`http_post "$baseUri/views" <<_EOB
{
    "name": "test-view",
    "annotation": "this is a test view"
}
_EOB
`
echo $result
view=`echo "$result" | jq -r .self`
view_id=`lastarg "$view"` 

result=`http_post "$baseUri/tokens" <<_EOB
{
    "name": "test-user",
    "accessTags": [ 
        "access:user", 
        "access:anybody", 
        "view:'$view_id'" 
    ]
}
_EOB
`
echo $result
