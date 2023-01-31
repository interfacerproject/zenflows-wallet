FROM golang:1.19-bullseye AS builder
RUN apt update && apt install -y build-essential git cmake vim python3 python3-pip zsh libssl-dev \
        && pip3 install meson ninja \
        && git clone https://github.com/dyne/Zenroom.git /zenroom
RUN cd /zenroom && make linux-go
ADD . /app
WORKDIR /app
RUN go build -o wallet .

FROM dyne/devuan:chimaera
WORKDIR /root
ENV HOST=0.0.0.0
ENV PORT=80
ENV GIN_MODE=release
EXPOSE 80
COPY --from=builder /app/wallet /root/
COPY --from=builder /zenroom/meson/libzenroom.so /usr/lib/
COPY --from=builder /usr/lib/x86_64-linux-gnu/libssl.so.1.1 /lib/
COPY --from=builder /usr/lib/x86_64-linux-gnu/libcrypto.so.1.1 /lib/
CMD ["/root/wallet"]
