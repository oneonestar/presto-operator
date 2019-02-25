#!/usr/bin/env bash
# This file is copy from presto-server-rpm/dist/etc/init.d/presto
./pre-start.sh

SERVICE_NAME='presto'
SERVICE_USER='presto'
# Launcher Config.
# Use data-dir from node.properties file (assumes it to be at /etc/presto). For other args use defaults from rpm install
NODE_PROPERTIES=/etc/presto/node.properties
DATA_DIR=$(grep -Po "(?<=^node.data-dir=).*" $NODE_PROPERTIES | tail -n 1)
SERVER_LOG_FILE=$(grep -Po "(?<=^node.server-log-file=).*" $NODE_PROPERTIES  | tail -n 1)
LAUNCHER_LOG_FILE=$(grep -Po "(?<=^node.launcher-log-file=).*" $NODE_PROPERTIES  | tail -n 1)
CONFIGURATION=(--launcher-config /usr/lib/presto/bin/launcher.properties --data-dir "$DATA_DIR" --node-config "$NODE_PROPERTIES" --jvm-config /etc/presto/jvm.config --config /etc/presto/config.properties --launcher-log-file "${LAUNCHER_LOG_FILE:-/var/log/presto/launcher.log}" --server-log-file "${SERVER_LOG_FILE:-/var/log/presto/server.log}")

if [ -f "/etc/presto/env.sh" ]; then
       source /etc/presto/env.sh
fi

echo "Starting ${SERVICE_NAME} "
if [ -z "$JAVA8_HOME" ]
then
    echo "Warning: No value found for JAVA8_HOME. Default Java will be used."
    sudo -u $SERVICE_USER /usr/lib/presto/bin/launcher run "${CONFIGURATION[@]}"
else
    sudo -u $SERVICE_USER PATH=${JAVA8_HOME}/bin:$PATH /usr/lib/presto/bin/launcher run "${CONFIGURATION[@]}"
fi