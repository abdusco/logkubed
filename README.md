# log<sup>3</sup> (logcubed)

**logcubed** is a mini app that helps you stream Kubernetes pod logs in realtime.

## Features

- Connects to K8S cluster and streams container logs in realtime.
- Reuses the connections if multiple clients monitor logs of the same container.
- Closes connection if no clients are listening.

## Configuration

| Environment variable | Description                                                        |
|----------------------|--------------------------------------------------------------------|
| `KUBECONFIG_PATH`    | Path to a kubeconfig file. If not given, it uses in-cluster config |
| `PORT`               | Port to listen to. Defaults to `8080`                              |

## Usage

Logs are served via websocket from `/logs` endpoint. 
Specify the `namespace`, `pod` and `container` names as query parameters.
```text
http://localhost:8080/logs?namespace=default&pod=my-nginx&container=nginx
```
Logs will be streamed as soon as they are available in plain text format.
You can connect to multiple containers at the same time.

## TODO
- [ ] Set up CI and automatic releases
- [ ] Add tests
- [ ] Add Docker support

