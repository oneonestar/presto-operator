############################
# STEP 1 build executable binary
############################
FROM golang:stretch

# Install git.
# Git is required for fetching the dependencies.
RUN apt update \
  && apt install git \
  && curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
COPY . $GOPATH/src/mypackage/myapp/
WORKDIR $GOPATH/src/mypackage/myapp/

# Fetch dependencies.
# Using go get.
RUN dep ensure -v

# Build the binary.
RUN go build -o /go/bin/hello

############################
# STEP 2 build a small image
############################
#FROM scratch
# Copy our static executable.
#COPY --from=builder /go/bin/hello /go/bin/hello
# Run the hello binary.
#ENTRYPOINT ["/go/bin/hello"]