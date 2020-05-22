# # <registry>/sidecar-base:latest
FROM fluent/fluent-bit:1.2.2 AS fluent
FROM envoyproxy/envoy:v1.13.1 AS envoy

FROM ${ubuntu:Image}:${ubuntu:Tag}

ENV LOG_COLLECTION_TLS=Off
ENV LOG_COLLECTION_TYPE=sidecar
ENV DEPLOY_TYPE=production
ENV DYNAMIC_SIDECAR_CONFIG=On

COPY --from=fluent /usr/lib/x86_64-linux-gnu/*sasl* \
									 /usr/lib/x86_64-linux-gnu/libz* \
									 /lib/x86_64-linux-gnu/libz* \
									 /usr/lib/x86_64-linux-gnu/libssl.so* \
									 /usr/lib/x86_64-linux-gnu/libcrypto.so* \
									 /lib/x86_64-linux-gnu/libsystemd* \
									 /lib/x86_64-linux-gnu/libselinux.so* \
									 /lib/x86_64-linux-gnu/liblzma.so* \
									 /usr/lib/x86_64-linux-gnu/liblz4.so* \
									 /lib/x86_64-linux-gnu/libgcrypt.so* \
									 /lib/x86_64-linux-gnu/libpcre.so* \
									 /lib/x86_64-linux-gnu/libgpg-error.so* \
									 /lib/x86_64-linux-gnu/

COPY --from=fluent /fluent-bit /usr/local/fluent-bit

COPY --from=envoy /usr/local/bin/envoy /usr/local/bin/envoy
COPY --from=envoy /etc/envoy/envoy.yaml /usr/local/etc/envoy.yaml
ADD ./envoy-default.yaml /usr/local/etc/envoy-default.yaml

ADD ./docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
ADD ./envoy-startup.sh /usr/local/bin/envoy-startup.sh

RUN chmod +x /usr/local/bin/docker-entrypoint.sh && \
		chmod +x /usr/local/bin/envoy-startup.sh  && \
		mv /usr/local/fluent-bit/bin/* /usr/local/bin/. && \
		mv /usr/local/fluent-bit/etc/* /usr/local/etc/. && \
		rm -rf /usr/local/fluent-bit

ADD ./fluent-bit.conf /usr/local/etc/fluent-bit.conf

ENTRYPOINT [ "docker-entrypoint.sh" ]