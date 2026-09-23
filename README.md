# s3ntry
deploy & monitoring tooling for my alexandria homelab.

## deploy
pipeilne done - self-hosted GH Actions runner on alexandriam CI gate + rollback path in place, push to main redeploys with no manual intervention.

## monitoring
a `/status` endpoint gives me health + alerting state for everything i have running + host machine

### health
`health/main.go` - one goroutine per target, polling concurrently, tracking up/down & latency also checks the host itself - connectivity, reboot-required, and available updates.

### alerting
`health/alert.go` - fires on state transitions via a Discord or Slack webhook

## deploying the /health binary itself 
the health checker here isnt dockerised like the applications it polls may be - `checkRebootRequired()` and `checkUpdatesAvailable()` need to see alexandria's *actual* `var/run/reboot-required` and `apt` state, and dockerising it would mean those checks  report on the container rather than the host. thus, it runs as a plain systemd **user** service under the `s3ntry` deploy user. 

one-time setup on alexandria, as the `s3ntry` user:
```
  # ~/.config/systemd/user/s3ntry-health.service
  [Unit]
  Description=s3ntry health checker

  [Service]
  ExecStart=/opt/services/s3ntry/health/s3ntry-health
  Restart=on-failure
  EnvironmentFile=/home/s3ntry/.env-s3ntry-health

  [Install]
  WantedBy=default.target
```
`~/.env-s3ntry-health` holds the secrets the binary reads via `os.Getenv` (currently just `DISCORD_WEBHOOK_URL`) - kept out of unit file for best practice. after writing the unit file, run:

```bash
loginctl enable-linger s3ntry
systemctl --user enable --now s3ntry-health
```

`deploy.yml` then builds the binary and runs `systemctl --user restart s3ntry-health`

## known limitation: the poll interval is a blind spot
everything here is a 5s/10s poll rather than an event stream, so a flap event inside a single poll window never gets flagged. this came up in my failure test that used real fault injection rather than just assertion - check `automation/failure-test.sh` and `writeup`, where 8/10 transitions were caught, and the 2 misses both landed in gaps shorter than the poll window between a kill and its recovery. reboot-required and updates-available don't have this problem, since they don't flap on sub-10s timescales. the real fix would be to swap to an event-driven stream rather than a poll, but for now a flap <10s is fine to miss.
