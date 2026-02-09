docker buildx create --use
docker buildx build --platform linux/amd64 -t us-east1-docker.pkg.dev/tone-486901/tone-repo/tone-registry:latest --push .
