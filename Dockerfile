FROM golang:alpine

ADD . /go/src/github.com/ValeriyKnyazhev/translator
RUN go install github.com/ValeriyKnyazhev/translator

ENTRYPOINT sleep 10 && /go/bin/translator
# EXPOSE 2345 //unknown port