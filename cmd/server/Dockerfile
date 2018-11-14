FROM alpine:3.8

#RUN apk add --no-cache --update ca-certificates

ADD ./ht /usr/local/bin/

RUN mkdir -p /ht/static
WORKDIR /ht

EXPOSE 10080

CMD ["/usr/local/bin/ht"]