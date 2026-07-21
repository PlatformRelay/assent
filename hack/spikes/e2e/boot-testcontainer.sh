#!/usr/bin/env bash
# Spike B (P2-E2-S01) — boot GitLab CE as a plain Docker container ("testcontainer" profile).
#
# Idempotent: tears down any previous instance before booting. `--teardown` cleans up and exits.
# On success prints exactly one machine-readable line:
#   RESULT boot_seconds=<n> rss_mb=<n>
#
# Measurement notes (see docs/planning/spikes/spike-b-e2e.md):
# - The image pull is excluded from boot_seconds (pre-pulled outside the timer) so both
#   profiles are compared on cached-image cold boot, which is what a warmed CI runner sees.
# - rss_mb is the container's steady-state memory usage (docker stats) after a settle period.
set -euo pipefail

IMAGE="${ASSENT_SPIKE_GITLAB_IMAGE:-gitlab/gitlab-ce:19.2.0-ce.0}"
NAME="assent-spike-gitlab"
HTTP_PORT="${ASSENT_SPIKE_HTTP_PORT:-8980}"
READY_TIMEOUT="${ASSENT_SPIKE_READY_TIMEOUT:-900}"
SETTLE_SECONDS="${ASSENT_SPIKE_SETTLE_SECONDS:-60}"

# Slimmed Omnibus config — this affects the RAM measurement and is recorded in the report:
# prometheus stack, container registry, KAS and outbound mail are disabled; puma runs in
# single mode; sidekiq concurrency is lowered. The monitoring whitelist is opened so the
# host-side readiness probe (which arrives from the docker bridge IP) gets a real answer.
OMNIBUS_CONFIG="
external_url 'http://localhost:${HTTP_PORT}'
prometheus_monitoring['enable'] = false
registry['enable'] = false
gitlab_kas['enable'] = false
gitlab_rails['gitlab_email_enabled'] = false
gitlab_rails['monitoring_whitelist'] = ['0.0.0.0/0']
puma['worker_processes'] = 0
sidekiq['concurrency'] = 5
"

teardown() {
  docker rm -f "${NAME}" >/dev/null 2>&1 || true
}

if [[ "${1:-}" == "--teardown" ]]; then
  teardown
  echo "teardown complete"
  exit 0
fi

teardown # idempotent: always start from a clean slate

docker image inspect "${IMAGE}" >/dev/null 2>&1 || docker pull "${IMAGE}" >/dev/null

start=$(date +%s)
docker run -d --name "${NAME}" \
  -p "${HTTP_PORT}:${HTTP_PORT}" \
  --shm-size 256m \
  --env GITLAB_OMNIBUS_CONFIG="${OMNIBUS_CONFIG}" \
  "${IMAGE}" >/dev/null

deadline=$((start + READY_TIMEOUT))
while true; do
  code=$(curl -s -o /dev/null -w '%{http_code}' "http://localhost:${HTTP_PORT}/-/readiness" || true)
  [[ "${code}" == "200" ]] && break
  if (($(date +%s) > deadline)); then
    echo "ERROR: readiness timeout after ${READY_TIMEOUT}s (last code: ${code})" >&2
    docker logs --tail 30 "${NAME}" >&2 || true
    exit 1
  fi
  sleep 5
done
boot_seconds=$(($(date +%s) - start))

sleep "${SETTLE_SECONDS}"
rss_mb=$(docker stats --no-stream --format '{{.MemUsage}}' "${NAME}" | awk '
  {u=$1}
  u ~ /GiB/ {gsub(/GiB/,"",u); printf "%d", u*1024; next}
  u ~ /MiB/ {gsub(/MiB/,"",u); printf "%d", u; next}
  {print 0}')

echo "RESULT boot_seconds=${boot_seconds} rss_mb=${rss_mb}"
