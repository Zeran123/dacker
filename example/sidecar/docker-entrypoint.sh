#!/bin/bash
nohup fluent-bit -c /usr/local/etc/fluent-bit.conf >> /tmp/fluent-bit.log 2>&1 &
nohup envoy-startup.sh >> /tmp/envoy.log 2>&1 &
if [ -n "$1" ]; then
	exec "$@"
fi