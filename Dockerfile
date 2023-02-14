# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2023 Dyne.org foundation <foundation@dyne.org>.
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as
# published by the Free Software Foundation, either version 3 of the
# License, or (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
