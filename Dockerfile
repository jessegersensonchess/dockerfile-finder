FROM golang:1.19-alpine3.16 as builder
RUN mkdir /build && \
	cd /build
WORKDIR "/build"
COPY . ./
RUN  CGO_ENABLED=0 go build -o dockerfile-finder main.go

FROM alpine:3.16
WORKDIR "/app"
COPY --from=builder /build/dockerfile-finder .
ARG USER=nonrootuser
ARG home=/home/${USER}
RUN addgroup -S app && \
    adduser \
    --disabled-password \
    --gecos "" \
    --home $home \
    ${USER} && \
	chown ${USER}:${USER} /app/dockerfile-finder

USER ${USER}
CMD [ "/app/dockerfile-finder" ]
