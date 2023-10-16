# Dockerfile
# zls


# # On Debian GNU/Linux 12 (bookworm) - image size ~= 928MB
# FROM golang:latest
# or
# On Alpine Linux v3.18  - image size ~= 335MB
FROM golang:alpine
#
WORKDIR /app
COPY . .
# Note that GOPATH=/go
RUN go build -ldflags "-s -w" -o /go/bin/zls
CMD ["zls"]


# EXPLORE multistage builds - image size ~= really small is the promise!
# # STEP 1: Build your binary
# FROM golang:alpine AS builder
# RUN apk update
# RUN apk add --no-cache git ca-certificates tzdata && update-ca-certificates
# COPY . .
# #RUN go get -d -v ./...
# #RUN go build -o /bin/my-service
# RUN go build -ldflags "-s -w" -o /bin/zls
# 
# # STEP 2: Use Scratch to build your smallest image
# FROM scratch
# COPY --from=builder /etc/ssl/certs/* /etc/ssl/certs/
# COPY --from=builder /bin/ /bin/
# CMD ["/bin/zls"]
