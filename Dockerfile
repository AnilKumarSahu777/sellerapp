FROM golang:1.12-alpine

WORKDIR /

COPY sellerapp /
CMD ["/sellerapp"]
