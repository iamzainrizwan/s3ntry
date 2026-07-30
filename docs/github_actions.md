# configuring a project to use the self-hosted runner

## 1. prepare the deployment dir

clone the repo into its permanent deployment location, using the **ssh** install method, NOT https.

```bash
  cd /opt/services
  git clone git@github.com@username/repo.git
```

## 2. config repo auth

the deployment user (`s3ntry`) should auth to gh using an **ssh deploy key**, rather than a personal gh account.
generate a key as the deployment user:

```bash
ssh-keygen -t ed25519 -C "alexandria-repo-reploy"
```

copy the pubkey:

```bash
cat ~/.ssh/id_ed25519.pub
```

in gh, go to `repos -> settings -> deploy keys` and add it there.
if you cloned using https, update to use ssh instead:

```bash
git remote set-url origin git@github.com:username/repo.git
```

verify:

```bash
ssh -T git@github.com
```

## 3. create deployment workflow

in the repo, create `.github/workflows/deploy.yml`
e.g.:

```YAML
name: deploy repo
on:
  push:
    branches:
      - main
jobs:
  deploy:
    runs-on: self-hosted

    steps:
    - name: deploy app
        run: |
        cd /opt/services/repo
        git pull
        docker compose up -d --build
```

the important line is `runs-on: self-hosted`, as we self-host a runner on alexandria instead of having github ssh in.

## 4. test

commit & push a change, then watch the runner on alexandria to make sure it pulls the changes and redeploys
