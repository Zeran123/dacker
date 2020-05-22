#!/bin/bash

if [ -n "$1" ]; then
  set -- "$@"
else
  set -- -DXms=256m -DXmx=256m "$@"
fi

java -Dspring.profiles.active=$ACTIVE_PROFILE $@ -jar /var/www/*.jar