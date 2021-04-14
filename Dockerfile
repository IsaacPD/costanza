FROM golang:latest as builder

WORKDIR /
COPY . .

RUN make

FROM frolvlad/alpine-glibc:glibc-2.31

RUN apk add --update --no-cache python3 && ln -sf python3 /usr/bin/python
RUN python -m ensurepip
RUN pip3 install --upgrade youtube-dl

RUN apk add --no-cache ffmpeg

COPY --from=builder /out/costanza /usr/local/bin

CMD ["costanza", "-v", "debug"]
