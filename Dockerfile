ARG GO_VERSION=1.20.5

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY ./main.go ./main.go
ARG TARGETOS TARGETARCH

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build main.go

#runner
FROM scratch
EXPOSE 3000
WORKDIR /
COPY --from=build /app/main /main

ENTRYPOINT [ "/main" ]