FROM golang:1.12.5-alpine AS gobuilder
RUN apk update && apk add --no-cache git build-base make gcc ca-certificates tzdata && update-ca-certificates

RUN adduser -D -g '' appuser

WORKDIR /goapp/

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo      \
       -ldflags="-w -s -X main.BUILDDATE=`date +%Y-%m-%dT%T%z`"  \    
       -o /goapp/myapp

COPY ./static/hello.html ./hello.html

FROM scratch
WORKDIR /app
EXPOSE 8001

# Import from builder.
COPY --from=gobuilder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=gobuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=gobuilder /etc/passwd /etc/passwd

# Copy our static executable
COPY --from=gobuilder /goapp/myapp /app/myapp
COPY --from=gobuilder /goapp/hello.html /app/static/hello.html

# Use an unprivileged user.
USER appuser

# Run the hello binary.
CMD ["./myapp"]