#!/bin/bash

temp_download_file="/tmp/envoy.yaml"
sidecar_default_config_file="/usr/local/etc/envoy-default.yaml"
sidecar_config_file="/usr/local/etc/envoy.yaml"

run_sidecar_background() {
	echo "[info] running sidecar background"
	envoy -c "$sidecar_config_file" &
}

validate_sidecar_config() {
	local f="$1"; shift;
	if envoy --mode validate --config-path "$f" | grep -q "OK"; then
		echo 1
	else
		echo 0
	fi
}

dynamic_sidecar_config() {
	local success=0;
	local api_site="https://example.com";
	if [ "$DEPLOY_TYPE" == "test" ]; then
		echo "[info] try to query api manager ${DEPLOY_TYPE} environment"
		api_site="https://example.com";
	else
		echo "[info] try to query api manager ${DEPLOY_TYPE} environment"
	fi

	if wget -O- "${api_site}/config" | jq -r ".data.content" > "$temp_download_file" && [ -s "$temp_download_file" ]; then
		success=1;
	fi;
	if [ "$success" -eq 1 ] && [ $(validate_sidecar_config "$temp_download_file") -eq 1 ]; then
		echo "[info] validate sidecar config success"
		cp -f "$temp_download_file" "$sidecar_config_file"
		echo "[info] try to running sidecar dynamic config"
	else
		echo "[info] validate sidecar config fail"
		echo "[info] try to running sidecar default config"
	fi
}

if [ -z "$LOG_COLLECTION_NAME" ]
then
    echo "[warn] LOG_COLLECTION_NAME is empty"
else
    echo "[info] LOG_COLLECTION_NAME is not empty"
	if [ "$DYNAMIC_SIDECAR_CONFIG" == "On" ]; then
		echo "[info] try to dynamic config sidecar"
		dynamic_sidecar_config
	else
		# use default config file
		cp -f "$sidecar_default_config_file" "sidecar_config_file"
	fi
fi

run_sidecar_background
