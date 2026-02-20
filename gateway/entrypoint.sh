#!/usr/bin/env bash
set -euo pipefail

# Wrapper around IBKR's bin/run.sh that exec's the Java process
# so it receives signals (SIGTERM) as PID 1 in the container.

config_file="${1:-root/conf.yaml}"
config_path="$(dirname "$config_file")"

RUNTIME_PATH="${config_path}:dist/ibgroup.web.core.iblink.router.clientportal.gw.jar:build/lib/runtime/*"

echo "runtime path : ${RUNTIME_PATH}"
echo "config file  : ${config_file}"
echo "Java Version: $(java -version 2>&1 | head -1)"

exec java \
    -server \
    -Dvertx.disableDnsResolver=true \
    -Djava.net.preferIPv4Stack=true \
    -Dvertx.logger-delegate-factory-class-name=io.vertx.core.logging.SLF4JLogDelegateFactory \
    -Dnologback.statusListenerClass=ch.qos.logback.core.status.OnConsoleStatusListener \
    -Dnolog4j.debug=true \
    -Dnolog4j2.debug=true \
    -cp "${RUNTIME_PATH}" \
    ibgroup.web.core.clientportal.gw.GatewayStart \
    --conf "../${config_file}"
