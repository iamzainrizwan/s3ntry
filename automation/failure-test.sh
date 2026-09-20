#!/usr/bin/env bash
# randomly kills and recovers a target container N times, timestamping each
# action so the resulting alert times (from journalctl or discord) can be
# cross-referenced afterwards to compute real time-to-detect numbers.
#
# usage: ./failure-test.sh [trials] [container]
set -euo pipefail

TARGET_CONTAINER="${2:-1337}"
TRIALS="${1:-5}"
LOG_FILE="failure-test-$(date +%Y-%m-%d).log"

echo "running $TRIALS trials against container '$TARGET_CONTAINER', logging to $LOG_FILE"

for i in $(seq 1 "$TRIALS"); do
  # random pre-kill wait so kills land at a random phase of the health
  # checker's poll cycle, not synced to it
  pre_wait=$(((RANDOM % 20) + 5)) # 5-24s
  echo "[trial $i] waiting ${pre_wait}s before kill..."
  sleep "$pre_wait"

  kill_time=$(date '+%Y-%m-%d %H:%M:%S.%3N')
  docker stop "$TARGET_CONTAINER" >/dev/null
  echo "[trial $i] killed at $kill_time" | tee -a "$LOG_FILE"

  # comfortably longer than the poll interval so detection + alert
  # definitely fires before recovery muddies the next trial
  down_wait=$(((RANDOM % 15) + 15)) # 15-29s
  sleep "$down_wait"

  recover_time=$(date '+%Y-%m-%d %H:%M:%S.%3N')
  docker start "$TARGET_CONTAINER" >/dev/null
  echo "[trial $i] recovered at $recover_time" | tee -a "$LOG_FILE"
done

echo "done. cross-reference $LOG_FILE against journalctl/discord to compute time-to-detect per trial."
