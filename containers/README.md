# Containers

Dockerfiles for TRISA test net application images. All docker containers expect that the context directory is the project root, e.g. they should be built as follows:

```
$ docker build -t trisa/myapp:latest -f containers/myapp/Dockerfile .
```

Additionally you can use the `skaffold build` command to build and push the images to Dockerhub.
