#! /bin/sh
docker build -t client .

docker create --name client-container client

docker cp client-container:/app/assets.tar.gz .

tar -xvf assets.tar.gz .
