# sidecar-java:openjdk-11-jre
FROM ${sidecar:Image}:${sidecar:Tag} AS sidecar

FROM ${openjdk-11-jre:Image}:${openjdk-11-jre:Tag}

COPY --from=sidecar /lib/x86_64-linux-gnu/ /lib/x86_64-linux-gnu/ 
COPY --from=sidecar /usr/local/ /usr/local/

ADD ./startup.sh /usr/local/bin/startup.sh

ENV ACTIVE_PROFILE=dev
RUN mkdir -p /var/www && \
    chmod +x /usr/local/bin/startup.sh

ENTRYPOINT [ "docker-entrypoint.sh" ]
CMD ["startup.sh"]