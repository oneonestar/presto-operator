FROM golang:1.12-alpine AS build

# Install tools required for project
RUN apk add --no-cache git
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go get github.com/golang/dep/cmd/dep
COPY Gopkg.lock Gopkg.toml /go/src/github.com/oneonestar/presto-operator/
WORKDIR /go/src/github.com/oneonestar/presto-operator/
# Install library dependencies
RUN dep ensure -vendor-only

# Copy the entire project and build it
# This layer is rebuilt when a file changes in the project directory
COPY . /go/src/github.com/oneonestar/presto-operator/
RUN go build -o /bin/project

# This results in a single layer image
#FROM scratch
FROM alpine:3.7
COPY --from=build /bin/project /bin/presto-operator
ENTRYPOINT ["/bin/presto-operator"]
