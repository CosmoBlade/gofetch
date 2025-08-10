#!/bin/bash

echo "building..."
go build -o gofetch main.go
if [ $? -ne 0 ]; then
    echo "build failed"
    exit 1
fi

echo "setting permissions..."
chmod +x gofetch

echo "installing..."
sudo mv gofetch /usr/local/bin/

echo "Done!"
echo "How to use: type 'gofetch'"
