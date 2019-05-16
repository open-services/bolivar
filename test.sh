#! /usr/bin/env bash

set -ex

docker build -t open-services/bolivar .

CONTAINER_ID=$(docker run -d open-services/bolivar)

function finish {
  echo "Showing you the logs from the container before kill/rm"
  docker logs $CONTAINER_ID
  echo "## END LOGS"
  docker kill $CONTAINER_ID
  docker rm $CONTAINER_ID
}
trap finish EXIT

echo $CONTAINER_ID

docker exec -it $CONTAINER_ID git clone https://github.com/webpack/webpack.git --depth=1 /app

docker exec -it $CONTAINER_ID yarn config set registry http://localhost:8080

# Give the Bolivar HTTP server a chance to start
sleep 5

time docker exec -it $CONTAINER_ID bash -c "cd /app && rm yarn.lock && yarn install --verbose --non-interactive"
