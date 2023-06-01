# Testing

## Prepare 

The webhooks can be tested from Gitlab.

```
docker run \
    --name gitlab \
    --publish 9443:443 --publish 9080:80 --publish 9022:22 \
    --shm-size 256m \
    gitlab/gitlab-ee:latest
```