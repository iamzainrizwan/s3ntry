# s3ntry
deploy & monitoring tooling for alexandria homelab. 
## deploy
pipline done - self hosted GH Actions runner on alexandria, CI gate + rollback path in place, push to main reploys with no manual intervention
## monitoring
lets me monitor the projects i have up at a glance - health & alerting both on a convenient dashboard
### health
`health/main.go`, polls targets concurrently by using a goroutine per service, tracks up/down + latency
### alerting
`health/alert.go`, sends alerts to Slack + Discord webhook implementations. should add email soon. 
