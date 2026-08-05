#!/usr/bin/env bash

# Measures how much memory the operator spends on Secrets it never reads.
#
# The operator watches Secrets to react to storage credential changes. Every watched type is backed by
# an informer, so an unrestricted Secret watch keeps every Secret in the cluster in memory, including
# the big ones nothing here cares about, for example Helm release Secrets. This eval fills a cluster
# with dummy Secrets and reports how much the operator grows because of them.
#
# Usage: the operator has to be deployed and Ready.
#   ./hack/secret_cache_memory_eval.sh
#   SECRET_COUNT=4000 THRESHOLD_MB=20 ./hack/secret_cache_memory_eval.sh

set -o errexit
set -o nounset
set -o pipefail

CONTEXT="${CONTEXT:-}"
NAMESPACE="${NAMESPACE:-docker-registry}"
EVAL_NAMESPACE="${EVAL_NAMESPACE:-secret-cache-eval}"
OPERATOR_SELECTOR="${OPERATOR_SELECTOR:-app.kubernetes.io/component=dockerregistry-operator.kyma-project.io}"
SECRET_COUNT="${SECRET_COUNT:-2000}"
SECRET_SIZE_KB="${SECRET_SIZE_KB:-16}"
# a scoped cache should stay flat, an unrestricted one grows with the payload of every Secret
THRESHOLD_MB="${THRESHOLD_MB:-20}"
BATCH_SIZE="${BATCH_SIZE:-200}"

kube() {
  if [ -n "${CONTEXT}" ]; then
    kubectl --context "${CONTEXT}" "$@"
  else
    kubectl "$@"
  fi
}

log() {
  echo ""
  echo "### $1"
}

# operatorMemoryMi prints the working set of the operator Pod in MiB.
operatorMemoryMi() {
  local output
  local waited=0
  while [ "${waited}" -lt 120 ]; do
    output=$(kube top pod -n "${NAMESPACE}" -l "${OPERATOR_SELECTOR}" --no-headers 2>/dev/null || true)
    if [ -n "${output}" ]; then
      echo "${output}" | awk '{gsub(/Mi/, "", $3); print $3; exit}'
      return 0
    fi
    sleep 5
    waited=$((waited + 5))
  done
  echo "metrics for the operator pod are not available" >&2
  return 1
}

# peakMemoryMi samples the operator memory a few times and prints the highest value, because the
# metrics pipeline reports with a delay.
peakMemoryMi() {
  local peak=0
  for _ in 1 2 3 4; do
    local sample
    sample=$(operatorMemoryMi)
    if [ "${sample}" -gt "${peak}" ]; then
      peak=${sample}
    fi
    sleep 20
  done
  echo "${peak}"
}

createSecrets() {
  local payload
  payload=$(head -c $((SECRET_SIZE_KB * 1024)) /dev/urandom | base64 | tr -d '\n')

  local created=0
  while [ "${created}" -lt "${SECRET_COUNT}" ]; do
    local manifest=""
    local batchEnd=$((created + BATCH_SIZE))
    if [ "${batchEnd}" -gt "${SECRET_COUNT}" ]; then
      batchEnd=${SECRET_COUNT}
    fi

    local index=${created}
    while [ "${index}" -lt "${batchEnd}" ]; do
      manifest="${manifest}
---
apiVersion: v1
kind: Secret
metadata:
  name: filler-${index}
  namespace: ${EVAL_NAMESPACE}
type: Opaque
data:
  payload: ${payload}"
      index=$((index + 1))
    done

    echo "${manifest}" | kube apply -f - >/dev/null
    created=${batchEnd}
    echo "  created ${created}/${SECRET_COUNT} secrets"
  done
}

cleanup() {
  log "removing the dummy secrets"
  kube delete namespace "${EVAL_NAMESPACE}" --ignore-not-found --wait=false >/dev/null
}
trap cleanup EXIT

log "restarting the operator to measure from a clean baseline"
kube -n "${NAMESPACE}" rollout restart deploy/dockerregistry-operator >/dev/null
kube -n "${NAMESPACE}" rollout status deploy/dockerregistry-operator --timeout=180s

log "measuring the baseline memory"
BASELINE=$(peakMemoryMi)
echo "  baseline: ${BASELINE} MiB"

log "creating ${SECRET_COUNT} secrets of ${SECRET_SIZE_KB} KiB in ${EVAL_NAMESPACE}"
kube create namespace "${EVAL_NAMESPACE}" --dry-run=client -o yaml | kube apply -f - >/dev/null
createSecrets

log "measuring the memory with the dummy secrets in the cluster"
LOADED=$(peakMemoryMi)
echo "  with ${SECRET_COUNT} secrets: ${LOADED} MiB"

GROWTH=$((LOADED - BASELINE))
PAYLOAD_MB=$((SECRET_COUNT * SECRET_SIZE_KB / 1024))

log "result"
echo "  secrets:      ${SECRET_COUNT} x ${SECRET_SIZE_KB} KiB = ${PAYLOAD_MB} MiB of payload"
echo "  baseline:     ${BASELINE} MiB"
echo "  loaded:       ${LOADED} MiB"
echo "  growth:       ${GROWTH} MiB (threshold ${THRESHOLD_MB} MiB)"

if [ "${GROWTH}" -gt "${THRESHOLD_MB}" ]; then
  echo ""
  echo "FAIL: the operator caches Secrets it does not use"
  exit 1
fi

echo ""
echo "PASS: the operator memory does not follow the number of Secrets in the cluster"
