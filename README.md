# Gitlab Listener

The Gitlab listener accepts webhook POST requests from Gitlab and sends them as cloudevents to Direktiv. The webhook content is in the `data` section of the cloud event. The content is based on the event posted from Gitlab. The event type is the lower case Gitlab event. e.g. `tag-push-hook`. It can be installed in plain mode without Knative or as a Knative source. The following YAML files are installing the listener but the value for `ingressClassName` might need to change.

## Plain Mode

[plain.yaml](https://github.com/direktiv-listeners/gitlab-listener/blob/main/kubernetes/plain.yaml)

## Knative Mode

[knative.yaml](https://github.com/direktiv-listeners/gitlab-listener/blob/main/kubernetes/knative.yaml)

## Configuration

| Environment Variable      | Description |
| ----------- | ----------- |
| DIREKTIV_GITLAB_DEBUG      | Enable debug mode      |
| DIREKTIV_GITLAB_SECRET      | Secret value for Gitlab webhook     |
| DIREKTIV_GITLAB_TOKEN      | Direktiv token if /broadcast API is used    |
| DIREKTIV_GITLAB_ENDPOINT      | Direktiv endpoint    |
| DIREKTIV_GITLAB_INSECURE_TLS      | Skip verifying certificates    |
| DIREKTIV_GITLAB_PATH      | Path to serve the webhook destination   |

