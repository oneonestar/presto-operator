FROM golang:1.11-alpine3.9 AS build

# Install tools required for project
RUN apk add --no-cache git
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go get github.com/golang/dep/cmd/dep
COPY Gopkg.lock Gopkg.toml /go/src/github.com/oneonestar/presto-controller/
WORKDIR /go/src/github.com/oneonestar/presto-controller/
# Install library dependencies
RUN dep ensure -vendor-only

# Copy the entire project and build it
# This layer is rebuilt when a file changes in the project directory
COPY . /go/src/github.com/oneonestar/presto-controller/
RUN go build -o /bin/project

# This results in a single layer image
#FROM scratch
FROM alpine:3.7
COPY --from=build /bin/project /bin/presto-controller
ENTRYPOINT ["/bin/presto-controller"]

# Install jsonnet
#RUN go get github.com/fatih/color \
#  && go get github.com/google/go-jsonnet \
#  && cd $GOPATH/src/github.com/google/go-jsonnet/jsonnet \
#  && go build \
#  && cp jsonnet /usr/local/bin


## Install git.
## Git is required for fetching the dependencies.
#RUN apt update \
#  && apt install git \
#  && curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
#COPY . $GOPATH/src/mypackage/myapp/
#WORKDIR $GOPATH/src/mypackage/myapp/
## Build the binary.
#RUN go build -o /go/bin/hello
#


# Fetch dependencies.
# Using go get.
#RUN dep ensure -v

############################
# STEP 2 build a small image
############################
#FROM scratch
# Copy our static executable.
#COPY --from=builder /go/bin/hello /go/bin/hello
# Run the hello binary.
#ENTRYPOINT ["/go/bin/hello"]