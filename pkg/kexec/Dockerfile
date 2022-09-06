ARG EVE_BUILDER_IMAGE=lfedge/eve-alpine:145f062a40639b6c65efa36bed1c5614b873be52
FROM ${EVE_BUILDER_IMAGE} AS build

ENV BUILD_PKGS patch curl make gcc perl util-linux-dev git mtools linux-headers musl-dev xz-dev elfutils-dev libbz2
ENV PKGS xz-libs elfutils-dev libbz2

RUN eve-alpine-deploy.sh

# Build kexec-tools
WORKDIR /tmp/kexec/kexec-tools-2.0.23
RUN curl -L https://mirrors.edge.kernel.org/pub/linux/utils/kernel/kexec/kexec-tools-2.0.23.tar.gz | tar -C .. -xzvf -
RUN ./configure --prefix /usr && \
    make && \
    make DESTDIR=/out install

FROM scratch
COPY --from=build /out /
