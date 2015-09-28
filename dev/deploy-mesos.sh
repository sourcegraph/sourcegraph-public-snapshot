#!/bin/bash

curl -XDELETE http://10.132.24.14:8080/v2/apps/sgx && \
curl -i -H 'Content-Type: application/json' -XPOST -d '{
"id": "sgx",
"cmd": "ORIG=`pwd`; mkdir -p /.sourcegraph/repos/my/repo2 && cd /.sourcegraph/repos/my/repo2 && git init && echo \"hello, world - my name is Sourcegraph\" > file1 && git add file1 && git config --global user.name MyName && git config --global user.email you@example.com && git commit --allow-empty -am MyFirstCommit && cd .. && rm -rf repo && git clone --bare repo2 repo && rm -rf repo2 && cd $ORIG && wget -q https://sourcegraph-release.s3-us-west-2.amazonaws.com/sgx/0.4/linux-amd64/sgx.gz && gunzip -f sgx.gz && chmod +x ./sgx && ./sgx selfupdate && DEBUG=t ./sgx --grpc-endpoint=http://localhost:$PORT2/ serve --addr=:$PORT1 --http-addr=:$PORT0 --grpc-addr=:$PORT2 --vcsstore.addr=:$PORT4 --appdash.http-addr=:$PORT3 --appdash.url=http://10.132.24.14:7800 --appdash.collector-addr=:$PORT5 --app-url=http://10.132.24.14:3000 --app.motd \"Welcome to Sourcegraph on Mesos!\" --app.custom-feedback-form \" \" ",
"mem": 512,
"cpus": 0.25,
"instances": 1,
"disk": 100.0,
"ports": [3000, 3001, 3100, 7800, 9091, 7801]
}' http://10.132.24.14:8080/v2/apps && \
sleep 1 && \
echo 'cd /run && /usr/local/bin/haproxy-marathon-bridge localhost:8080 > haproxy.cfg && haproxy -f haproxy.cfg -p haproxy.pid -sf $(cat /run/haproxy.pid) && cat haproxy.cfg' | ssh -i ~/.ssh/sg-dev.pem root@10.132.24.14 bash -
