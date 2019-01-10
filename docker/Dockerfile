############################
# STEP 1 build executable binary
############################
FROM golang:stretch


# Fetch dependencies.
# Using go get.
RUN dep ensure -v

# Install jsonnet
RUN go get github.com/fatih/color \
  && go get github.com/google/go-jsonnet \
  && cd $GOPATH/src/github.com/google/go-jsonnet/jsonnet \
  && go build \
  && cp jsonnet /usr/local/bin


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



############################
# STEP 2 build a small image
############################
#FROM scratch
# Copy our static executable.
#COPY --from=builder /go/bin/hello /go/bin/hello
# Run the hello binary.
#ENTRYPOINT ["/go/bin/hello"]