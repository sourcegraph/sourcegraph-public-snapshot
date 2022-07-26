FROM sourcegraph/alpine-3.14:154143_2022-06-13_1eababf8817e@sha256:f1c4ac9ca1a36257c1eb699d0acf489d83dd86e067b1fc3ea4a563231a047e05
RUN apk --no-cache add tzdata
COPY build-tracker /usr/local/bin/
ENTRYPOINT ["build-tracker"]
