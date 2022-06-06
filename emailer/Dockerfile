FROM golang:1.17.9-buster AS build

RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends \
	libtesseract-dev \
	tesseract-ocr-tur \
	tesseract-ocr-eng \
	libleptonica-dev \
	; \
	rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.* ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=1 go build -o /app/emailer
ENTRYPOINT ["./emailer"]

FROM debian:buster
RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends \
	libtesseract-dev \
	libleptonica-dev \
	tesseract-ocr-tur \
	tesseract-ocr-eng \
	; \
	rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /app/emailer .

ENTRYPOINT ["/app/emailer"]