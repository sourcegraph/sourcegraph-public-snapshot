FROM node:6.9-slim

RUN npm install -g yarn

RUN apt-get update
RUN apt-get install -y zip
RUN apt-get install -y git

ADD ./bundle.sh /bundle.sh
