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
RUN CGO_ENABLED=1 go build -o emailer
ENTRYPOINT ["./emailer"]

# FROM busybox
# WORKDIR /app
# COPY --from=build /src/emailer .
# COPY --from=build /usr/include/tesseract /usr/include/tesseract
# COPY --from=build /usr/include/leptonica /usr/include/leptonica 

# ENTRYPOINT ["/app/emailer"]