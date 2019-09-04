FROM golang:1.12

COPY scrapedumper scrapedumper

CMD ["./scrapedumper"]
