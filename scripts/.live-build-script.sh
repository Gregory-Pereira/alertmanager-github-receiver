docker build . -t quay.io/grpereir/alertmanager-github-receiver:latest
docker run -it --mount type=bind,source="$(pwd)/labelmap",target=/home/alertmanager-github-reciever-config/,readonly -p 9393:9393 quay.io/grpereir/alertmanager-github-receiver:latest -authtoken=$(GITHUBTOKEN) -org=Gregory-Pereira -repo=alerts"
