#!/bin/bash
# does a parallel build of all go files in the clients directory
for gofile in clients/*.go ;
do
    echo "Building $gofile"
    # extract the base name: "clients/test.go" -> "test.go"
    name=$(basename "$gofile")
    # start the build in the background. The name of the binary will be the name of the file w/o ".go"
    ( go build -o ${name%%.*} $gofile ) &
done
# wait for all started processes
wait
