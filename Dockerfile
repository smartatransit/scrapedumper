FROM alpine:3.9

COPY scrapedumper scrapedumper

CMD ["/scrapedumper", "--config-path", "/config.yml"]
