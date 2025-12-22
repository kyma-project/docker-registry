FROM --platform=$BUILDPLATFORM europe-docker.pkg.dev/kyma-project/prod/external/library/alpine:3.23.2

RUN set -ex \
    && apk upgrade --no-cache \
    && apk add --no-cache ca-certificates apache2-utils
