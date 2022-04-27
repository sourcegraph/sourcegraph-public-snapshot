FROM golang:1.17-alpine AS builder

WORKDIR /go/src/tracking-issue
COPY . .
RUN go mod init tracking-issue
RUN go get ./...
RUN CGO_ENABLED=0 go install .

FROM sourcegraph/alpine-3.12:142406_2022-04-14_8836ac3499f4@sha256:4681a48d1fb9a73fef1b540c08b3411f797351bbeda749f5dca21213a1e71526
COPY --from=builder /go/bin/* /usr/local/bin/
ENTRYPOINT ["tracking-issue"]
