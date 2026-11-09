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

FROM golang:1.22-bookworm AS builder
ENV GONOPROXY=
RUN apt-get update && apt-get install -y build-essential git cmake python3 meson ninja-build libssl-dev zsh xxd \
        && git clone https://github.com/dyne/Zenroom.git /zenroom \
        && git -C /zenroom checkout 4388660e7b0b
# The same Zenroom commit as the Go binding in go.mod (master dropped linux-go).
RUN cd /zenroom && make linux-go
ADD . /app
WORKDIR /app
RUN go build -o wallet .

# Devuan 5 (Debian 12 base): OpenSSL 3 and the glibc the builder linked against.
FROM dyne/devuan:daedalus
RUN apt-get update && apt-get install -y --no-install-recommends libssl3 ca-certificates \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /root
ENV HOST=0.0.0.0
ENV PORT=80
ENV GIN_MODE=release
EXPOSE 80
COPY --from=builder /app/wallet /root/
COPY --from=builder /zenroom/meson/libzenroom.so /usr/lib/
CMD ["/root/wallet"]
