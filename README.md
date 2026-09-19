# s3ntry
deploy & monitoring tooling for alexandria homelab. 
## deploy
pipline done - self hosted GH Actions runner on alexandria, CI gate + rollback path in place, push to main reploys with no manual intervention
## monitoring
lets me monitor the projects i have up at a glance - health & alerting both on a convenient dashboard
### health
`health/main.go`, polls targets concurrently by using a goroutine per service, tracks up/down + latency, and checks on the host's connectivity, reboot status, and available updates 
### alerting
`health/alert.go`, sends alerts to Slack + Discord webhook implementations. should add email soon. 

## deploying the health binary itself
unlike 1337, the health checker isn't dockerised. `checkRebootRequired()` and
`checkUpdatesAvailable()` need to see alexandria's *actual* `/var/run/reboot-required`
and `apt` state - containerising it would mean those checks report on the container,
not the host, which defeats the point of a host monitor. so it runs as a plain
systemd **user** service instead, under the `s3ntry` deploy user (which deliberately
has no sudo, see `docs/setup/runner.md`).

one-time setup on alexandria, as the `s3ntry` user:

```ini
# ~/.config/systemd/user/s3ntry-health.service
[Unit]
Description=s3ntry health checker

[Service]
ExecStart=/opt/services/s3ntry/health/s3ntry-health
Restart=on-failure

[Install]
WantedBy=default.target
```

```bash
loginctl enable-linger s3ntry   # so the user service survives logout
systemctl --user enable --now s3ntry-health
```

`deploy.yml`'s job then just builds the binary and restarts that unit -
`systemctl --user restart s3ntry-health` - no sudo required anywhere in the pipeline.
