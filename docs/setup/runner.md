# setting up a self-hosted runner

this only needs to be done once per server. the server in this example will be `alexandria`, my actual home server.

## 1. create a deployment user

use a dedicated user for deployments rather than a personal or admin one.

```bash
sudo adduser s3ntry
```

give it docker perms (if necessary):

```bash
sudo usermod -aG docker s3ntry
```

## 2. download the runner

switch to the deployment user:

```bash
sudo -i -u s3ntry 
```

create a directory for the runner:

```bash
mkdir ~/actions-runner
cd ~/actions-runner
```

in gh, go to `repo -> settings -> actions -> runners -> new self-hosted`
download the runner archive and extract it into the new directory you just made `actions-runner`

## 3. configure the runner

run the configuration command from gh, usually something like:

```bash
./config.sh \
  --url https://github.com/username/repo \ 
  --token your_token_here
```

## 4. install as a service so it persists on reboots and runs in the bg

exit back to your work user, then install as a systemd service:

```bash
sudo bash -c 'cd /home/s3ntry/actions-runner && ./svc.sh install s3ntry' 
sudo bash -c 'cd /home/s3ntry/actions-runner && ./svc.sh start'
```

> [!NOTE]
> this is a kinda convoluted way to do it that retains s3ntry not having sudo. if you'd rather, you can temporarily add your deployment user to sudoers and remove it after the runner is working

check that its running:

```bash
sudo systemctl status actions-runner*
```

1. check on github
go to `repo -> settings -> actions -> runners`, where the runner should show up as idle.

create a workflow with `runs-on: self-hosted` (see [project.md](/docs/setup/project.md)), push a commit, and verify the runner picks up the job on its own.
