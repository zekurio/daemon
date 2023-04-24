#!/bin/sh

function populate {
    FILE=$1
    DATA=$2

    echo "---"
    echo "Populate: $FILE"
    printf "$DATA" | tee $FILE
    echo ""
    echo "CHECK: $(cat $FILE)"
}

FILE_LOCATION="./internal/util/embedded"

VERSION=$(git describe --tags --abbrev=0)
COMMIT=$(git rev-parse HEAD)

[ "$VERSION" == "" ] && {
    VERSION="c${COMMIT:0:8}"
}

populate "$FILE_LOCATION/Version.txt" $VERSION
populate "$FILE_LOCATION/Commit.txt" $COMMIT
populate "$FILE_LOCATION/Release.txt" "true"