FROM golang:1.24 AS build
WORKDIR /usr/src/sss
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY ./cmd/api/ ./cmd/api
COPY ./internal/data/ ./internal/data/
RUN CGO_ENABLED=0 go build -v -o /usr/local/bin/ ./...

FROM golang:1.24 AS user
RUN adduser --uid 10001 --shell /bin/false appuser \
    && cat /etc/passwd | grep appuser > /etc/passwd_appuser

FROM scratch
COPY --from=user /etc/passwd_appuser /etc/passwd
COPY --from=build /usr/local/bin/api /app
USER appuser
CMD ["/app"]