repo -> settings -> actions -> runners -> new self-hosted runner

run commands, and use github's:

```bash
./config.sh \
  --url https://github.com/YOURNAME/YOURREPO \
  --token YOUR_TOKEN

then in required repo, create `.github/workflows/deploy.yml`, usually looks like:

```yaml
  name: deploy [repo]
  on:
    push:
      branches:
        - main
  jobs:
    deploy:
      runs-on:self-hosted # IMPORTANT! SINCE GIT IS NOT SSHING IN
      steps:
        - name: deploy app
        run : |
            # whatever bash cmds you want here 
        
```
