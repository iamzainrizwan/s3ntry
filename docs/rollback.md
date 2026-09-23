# how to roll back a project if something goes wrong
we're all prone to mistakes, and sometimes (usually) we will push something that the workflow doesn't catch and breaks something. or maybe we just want to revert a version! how do we do that?

## reverting using Docker
if the application uses Docker containers, you can view the history and use a different image. 
### step 1: looking for ID
navigate to your app's directory, and view your local history.
```bash 
cd /opt/services/repo
docker compose ps
```
to find the older tag or image hash of the image you want, check the local images:
```bash
docker image ls
```
### step 2: force deployment using the older tag
if you tag builds sequentially (eg `v1.2.0`, `v1.1.9`), update `docker-compose.yml` file to point back to point back to the desired tag. if you rely on `latest`, you can temporarily force-run a specific container directly:
```bash
docker run -d --name my-app-rollback -p 80:80 <previous_image_id>
```

## reverting using GitHub
this is the best way for a permanent fix. treats rollback as a brand-new code change, triggering deployment pipeline to overwite broken prod.
### step 1: finding the broken commit
clone repo locally, or go to your GitHub commit history. find the alphanumeric hash of the breaking commit. 
```bash
git log --oneline
```
### step 2: create revert commit:
use the native Git revert command, creating a new commit that applies the **opposite** of the broken changes.  
```bash
git revert <broken_commit_hash> -m "message"
```
