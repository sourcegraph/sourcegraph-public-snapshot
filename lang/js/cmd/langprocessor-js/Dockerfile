#
# Docker image for Java Language Processor
#

#
# Dependencies
#
FROM node:5.7.0-slim

RUN apt-get update -y
RUN npm install typescript -g 

ENV SGPATH="/sourcegraph"

ADD js-language-processor/build /
ADD js-language-processor/node_modules /node_modules

ENTRYPOINT node /main.js
