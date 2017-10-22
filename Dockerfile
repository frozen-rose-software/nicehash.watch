FROM golang:1.7

ADD . /go/src/frozenrosesoftware.com/nicehash.watch

RUN cd /go/src/frozenrosesoftware.com/nicehash.watch; make

ENTRYPOINT cd /go/src/frozenrosesoftware.com/nicehash.watch; ./nicehash.watch

EXPOSE 8080
