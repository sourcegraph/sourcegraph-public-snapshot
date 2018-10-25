# Installing code intelligence on AWS manually

These instructions walk you through adding [code intelligence](../index.md) to Sourcegraph Server manually **on AWS**.

**Most users will never need to follow these steps, and instead should rely on the [default automatic installation](index.md).**

For more information, see "[Installing code intelligence](index.md)".

## Manual installation

Generally you do not ever need to perform manual installation of code intelligence. Language servers are automatically downloaded, set up, and configured when you add a repository with a language that we support. If, however, you are using a modified `docker run` command rather than the one we recommend (for example, if you have removed the Docker socket pass-through flag, or if you are running Sourcegraph with an older user-data `#cloud-config`), you can use the following steps to configure code intelligence manually on AWS:

1.  SSH into the node running your Sourcegraph instance from the previous step, e.g:

    ```
    ssh -i ~/.ssh/key.pem ec2-user@$PUBLIC_URL
    ```

2.  Stop the running `sourcegraph/server` container:

    ```
    docker ps # get the $CONTAINER_ID of the running sourcegraph/server container
    docker rm -f $CONTAINER_ID
    ```

3.  To run Sourcegraph with language servers for code intelligence, you must first create a Docker user-defined network and run Docker containers on this network.

    ```
    docker network create --driver bridge lsp
    docker run -d --publish 80:7080 --network lsp --name sourcegraph --restart unless-stopped --volume /home/ec2-user/.sourcegraph/config:/etc/sourcegraph --volume /home/ec2-user/.sourcegraph/data:/var/opt/sourcegraph sourcegraph/server:2.12.2
    docker run -d --network=lsp --name=typescript --restart unless-stopped sourcegraph/codeintel-typescript:latest
    ```

    Alternatively, to boot a fresh EC2 instance running Sourcegraph and language servers, simply modify your user data [previously provided here](../../../admin/install/docker/aws.md#deploy-to-ec2) (or cloud-init.txt) to look like the following:

    ```
    #cloud-config
    repo_update: true
    repo_upgrade: all

    packages:
    - docker

    runcmd:
    - mkdir -p /home/ec2-user/.sourcegraph/config
    - mkdir -p /home/ec2-user/.sourcegraph/data
    - sed -i -e 's/1024/10240/g' /etc/sysconfig/docker
    - sed -i -e 's/4096/40960/g' /etc/sysconfig/docker
    - service docker start
    - usermod -a -G docker ec2-user
    - [ sh, -c, 'docker network create --driver bridge lsp' ]
    - [ sh, -c, 'docker run -d --publish 80:7080 --publish 443:7443 --network lsp --name sourcegraph --restart unless-stopped --volume /home/ec2-user/.sourcegraph/config:/etc/sourcegraph --volume /home/ec2-user/.sourcegraph/data:/var/opt/sourcegraph sourcegraph/server:2.12.2' ]
    - [ sh, -c, 'docker run -d --network=lsp --name=typescript --restart unless-stopped sourcegraph/codeintel-typescript:latest' ]
    ```

4.  [update site configuration to point to the language servers](index.md#configure-sourcegraph-to-connect-to-the-language-servers). You can also see a list of all available language servers there.

---

## Next steps

To get code intelligence on your code host and/or code review tool, see the [browser extension documentation](../../../integration/browser_extension.md).
