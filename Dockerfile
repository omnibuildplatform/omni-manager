FROM golang:alpine3.13 as builder
LABEL maintainer="luonancom<luonancom@qq.com>"
WORKDIR /app
COPY . /app
RUN go mod download
RUN CGO_ENABLED=0 go build -o omni-manager

FROM alpine/git:v2.30.2
ARG user=app
ARG group=app
ARG home=/app
# to fix mv unrecoginzed option T
RUN apk update --no-cache && apk add --no-cache coreutils=8.32-r2 \
    && addgroup -S ${group} && adduser -S ${user} -G ${group} -h ${home}

USER ${user}
WORKDIR ${home}
RUN mkdir -p ${home}/logs  $$ -p ${home}/conf  
COPY --chown=${user} --from=builder /app/omni-manager .
COPY --chown=${user} ./conf ./conf/
#to fix the directory permission issue
VOLUME ["${home}/logs","${home}/conf"]

ENV PATH="${home}:${PATH}" 

EXPOSE 8080 8888
ENTRYPOINT ["/app/omni-manager"]