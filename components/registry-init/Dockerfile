# result container base on alpine:3.19.1
FROM europe-docker.pkg.dev/kyma-project/prod/external/alpine@sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b

RUN set -ex \
    && apk upgrade --no-cache \
    && apk add --no-cache ca-certificates apache2-utils
