#!/bin/bash
SHOW_DETAILS=$1
TMP=$(mktemp)
LIMIT=80
go test ./... -coverprofile=$TMP || exit 1
TOTAL=$(go tool cover -func=$TMP | grep "total:" | awk '{print $3}' | sed -e 's/%//' | awk -F. '{print $1}') 
if [[ $TOTAL -lt $LIMIT ]];then
    echo FAIL
    echo "Coverage: $TOTAL Limit: $LIMIT"
    exit 1
fi
echo "Coverage: $TOTAL Limit: $LIMIT"
echo $SHOW_DETAILS
if [[ $SHOW_DETAILS == "details" ]];then
    go tool cover -func=$TMP
fi
rm "$TMP"