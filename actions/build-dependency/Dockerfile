FROM paketobuildpacks/build:full

ARG cnb_uid=0
ARG cnb_gid=0

USER ${cnb_uid}:${cnb_gid}

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
