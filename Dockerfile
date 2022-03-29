FROM golang:alpine3.13 as builder
LABEL maintainer="luonancom<luonancom@qq.com>"
WORKDIR /app
COPY . /app
RUN go mod download
RUN CGO_ENABLED=0 go build -o omni-manager
#since git-sync doesn't have a binary release, we copy binary from official docker image
FROM k8s.gcr.io/git-sync/git-sync:v3.3.1 as gitsync
RUN echo "git-sync prepared"
FROM alpine/git:v2.30.2
ARG user=app
ARG group=app
ARG home=/app
# to fix mv unrecoginzed option T
RUN apk update --no-cache && apk add --no-cache coreutils=8.32-r2 \
    && addgroup -S ${group} && adduser -S ${user} -G ${group} -h ${home}

USER ${user}
WORKDIR ${home}
RUN mkdir -p ${home}/logs  $$ -p ${home}/conf $$ -p ${home}/docs
COPY --chown=${user} --from=builder /app/omni-manager .
COPY --chown=${user} --from=gitsync /git-sync .
COPY --chown=${user} ./conf ./conf/
COPY --chown=${user} ./docs ./docs/
#to fix the directory permission issue
VOLUME ["${home}/logs","${home}/conf","${home}/docs"]

ENV PATH="${home}:${PATH}"
ENV APP_ENV="release"
ENV APP_PORT="8080"
ENV DB_USER="root"
ENV DB_PSWD="rootpswd"
ENV DB_HOST="192.168.1.193"
ENV DB_NAME="obs_meta"
ENV REDIS_ADDR="192.168.1.193"
ENV REDIS_DB="0"
ENV REDIS_PSWD=""
ENV WS_HOST="192.168.1.193"
ENV WS_PORT="8888"

EXPOSE 8080 8888
ENTRYPOINT ["/app/omni-manager"]