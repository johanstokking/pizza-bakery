FROM golang:1.22 as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o /bin/fivestars ./cmd/fivestars

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=build /bin/fivestars /bin/fivestars
EXPOSE 8080
CMD ["/bin/fivestars"]