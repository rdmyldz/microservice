FROM golang:1.17.9-buster AS build

WORKDIR /src
COPY go.* ./
RUN go mod download

COPY *.go ./
COPY *.html /app/
RUN CGO_ENABLED=0 go build -o /app/website

FROM scratch
WORKDIR /app
COPY --from=build /app .
ENTRYPOINT ["/app/website"]