FROM golang:1.25.0 AS builder

WORKDIR /go/bin

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

ARG GIT_COMMIT=dev
ARG BUILD_DATE=unknown


RUN go build -ldflags "\
  -X 'users/version.GitCommit=${GIT_COMMIT}' \
  -X 'users/version.BuildDate=${BUILD_DATE}' \
" -o /go/bin/users .




FROM ubuntu

COPY --from=builder /go/bin/users /go/bin/users
COPY --from=builder  /go/bin/config/settings.example.toml /go/bin/config/settings.toml
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/


WORKDIR /go/bin/

EXPOSE 5300

ENTRYPOINT ["/go/bin/users"]
