FROM node:12.0.0-stretch

COPY release/bolivar-vlatest-linux-amd64 /usr/bin/bolivar

CMD bolivar --repo-path=$(mktemp -d)
