FROM alpine:3.10@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a AS builder

RUN apk add --no-cache nodejs-current=12.4.0-r0 nodejs-npm=10.16.3-r0
RUN npm install -g yarn@1.17.3

COPY package.json tsconfig.json yarn.lock /lsif/
COPY src /lsif/src

RUN yarn --cwd /lsif
RUN yarn --cwd /lsif run build
