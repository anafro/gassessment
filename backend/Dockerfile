FROM golang:1.23-alpine AS build
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o /server .
FROM alpine:3.21
COPY --from=build /server /server
CMD ["/server"]
