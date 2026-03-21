#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGE_NAME="ibkr-cpgw"
CONTAINER_NAME="ibkr-cpgw"
HOST_PORT="${IBKR_CPGW_PORT:-5000}"

# Auto-detect container runtime: prefer podman, fall back to docker.
if command -v podman &>/dev/null; then
  RUNTIME=podman
elif command -v docker &>/dev/null; then
  RUNTIME=docker
else
  echo "Error: neither podman nor docker found in PATH." >&2
  exit 1
fi

usage() {
  cat <<EOF
Usage: $(basename "$0") <command>

Commands:
  build   Build the container image
  run     Start the gateway container
  stop    Stop and remove the gateway container
  logs    Follow container logs
  status  Show container status
  alive   Check if the container is running (exit 0/1)
  start   Build image if needed, then run container (idempotent)

Container runtime: ${RUNTIME}
Override with CONTAINER_RUNTIME=docker|podman
EOF
}

# Allow explicit override via env var.
RUNTIME="${CONTAINER_RUNTIME:-${RUNTIME}}"

container_exists() {
  if [ "${RUNTIME}" = "podman" ]; then
    podman container exists "$1" 2>/dev/null
  else
    docker container inspect "$1" &>/dev/null
  fi
}

image_exists() {
  if [ "${RUNTIME}" = "podman" ]; then
    podman image exists "$1" 2>/dev/null
  else
    docker image inspect "$1" &>/dev/null
  fi
}

cmd_build() {
  echo "Building ${IMAGE_NAME} (${RUNTIME})..."
  #shellcheck disable=SC2086 # Allow word splitting for BUILD_A  #shellcheck disable=SC2086 # Allow word splitting for BUILD_AR  #shellcheck disable=SC2086 # Allow word splitting for BUILD_ARGS
  ${RUNTIME} build ${BUILD_ARGS:-} -t "${IMAGE_NAME}" -f "${SCRIPT_DIR}/Containerfile" "${SCRIPT_DIR}"
  echo "Done. Image: ${IMAGE_NAME}"
}

cmd_run() {
  if container_exists "${CONTAINER_NAME}"; then
    echo "Container ${CONTAINER_NAME} already exists. Stop it first."
    exit 1
  fi

  # Podman uses :Z for SELinux relabeling; docker does not.
  MOUNT_OPTS="ro"
  [ "${RUNTIME}" = "podman" ] && MOUNT_OPTS="ro,Z"

  echo "Starting ${CONTAINER_NAME} (https://localhost:${HOST_PORT})..."
  ${RUNTIME} run -d \
    --name "${CONTAINER_NAME}" \
    -p "127.0.0.1:${HOST_PORT}:5000" \
    -v "${SCRIPT_DIR}/conf.yaml:/opt/clientportal.gw/root/conf.yaml:${MOUNT_OPTS}" \
    "${IMAGE_NAME}"

  echo "Gateway started. Authenticate at https://localhost:${HOST_PORT}"
}

cmd_stop() {
  echo "Stopping ${CONTAINER_NAME}..."
  ${RUNTIME} stop "${CONTAINER_NAME}" 2>/dev/null || true
  ${RUNTIME} rm "${CONTAINER_NAME}" 2>/dev/null || true
  echo "Stopped."
}

cmd_logs() {
  ${RUNTIME} logs -f "${CONTAINER_NAME}"
}

cmd_status() {
  if container_exists "${CONTAINER_NAME}"; then
    ${RUNTIME} ps --filter "name=${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
  else
    echo "Container ${CONTAINER_NAME} is not running."
  fi
}

cmd_alive() {
  if container_exists "${CONTAINER_NAME}"; then
    exit 0
  else
    exit 1
  fi
}

cmd_start() {
  if container_exists "${CONTAINER_NAME}"; then
    exit 0
  fi
  if ! image_exists "${IMAGE_NAME}"; then
    cmd_build
  fi
  cmd_run
}

case "${1:-}" in
build) cmd_build ;;
run) cmd_run ;;
stop) cmd_stop ;;
logs) cmd_logs ;;
status) cmd_status ;;
# Exec contract commands (used by midwatch exec adapter)
alive) cmd_alive ;;
start) cmd_start ;;
*)
  usage
  exit 1
  ;;
esac
