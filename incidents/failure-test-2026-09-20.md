# failure test - 2026-09-20

deliberate failure test, verifying the wired Discord alerter actually fires on a real up/down transition, not just in theory.

## setup

target: `1337` container on alexandria (`container_name: 1337`, `restart: unless-stopped`, so a manual `docker stop` doesn't get auto-recovered). health checker polls every 10s (`monitor(targets, 10*time.Second, ...)`).

ran `automation/failure-test.sh 5` - 5 trials, each stopping and restarting the container with a randomized wait beforehand so kills land at a random phase of the poll cycle instead of syncing to it.

## raw data

| trial | killed at | recovered at |
|---|---|---|
| 1 | 01:56:50.753 | 01:57:10.032 |
| 2 | 01:57:15.156 | 01:57:39.428 |
| 3 | 01:57:59.551 | 01:58:18.847 |
| 4 | 01:58:36.988 | 01:58:55.287 |
| 5 | 01:59:12.420 | 01:59:32.704 |

alert times, cross-referenced from each Discord message's snowflake ID (encodes creation time to the millisecond):

| event | time | matched alert | time-to-detect |
|---|---|---|---|
| trial 1 down | 01:56:50.753 | 01:56:56.548 | 5.795s |
| trial 1 up | 01:57:10.032 | **missed** | - |
| trial 2 down | 01:57:15.156 | **missed** | - |
| trial 2 up | 01:57:39.428 | 01:57:46.668 | 7.240s |
| trial 3 down | 01:57:59.551 | 01:58:06.383 | 6.832s |
| trial 3 up | 01:58:18.847 | 01:58:26.431 | 7.584s |
| trial 4 down | 01:58:36.988 | 01:58:46.352 | 9.364s |
| trial 4 up | 01:58:55.287 | 01:58:56.389 | 1.102s |
| trial 5 down | 01:59:12.420 | 01:59:16.401 | 3.981s |
| trial 5 up | 01:59:32.704 | 01:59:36.361 | 3.657s |

**n=8 detected transitions** (2 of the expected 10 were missed - see finding below). mean time-to-detect: **5.694s**, min 1.102s, max 9.364s. this lands almost exactly on the ~5s predicted by a uniform-random phase over a 10s poll interval, which is a good sanity check that the system's actually behaving the way the design implies rather than just producing plausible-looking numbers.

## finding: a flap faster than the poll interval is invisible

trial 1's recovery and trial 2's failure were never alerted on at all. the gap between trial 1's recovery (01:57:10.032) and trial 2's kill (01:57:15.156) was only **5.124 seconds** - shorter than the 10s poll interval. the container went up, then back down, entirely between two polls: the checker polled while down (before the recovery), then polled again while down (after the next kill), saw no change either time, and never fired anything for either transition.

this isn't a bug in the alerting code - it's an inherent limit of periodic polling. anything that flaps faster than the poll interval is invisible to this system. worth stating plainly in the README's "what could go wrong" section rather than glossing over it: **this monitors "is it up right now, as of the last poll," not "did it ever go down."** lowering the poll interval narrows the blind spot but never closes it - only event-driven monitoring (e.g. reacting to container lifecycle events directly instead of polling HTTP) would.

## conclusion

alerting is real and works: 8/10 genuine transitions were detected and alerted on, averaging 5.7s, bounded by the 10s poll interval as designed. the 2 misses are explained, not mysterious, and point at a specific, honest limitation rather than a flaky test.
