# Installing code intelligence on DigitalOcean manually

These instructions walk you through adding [code intelligence](../index.md) to Sourcegraph Server manually **on DigitalOcean**.

**Most users will never need to follow these steps, and instead should rely on the [default automatic installation](index.md).**

For more information, see "[Installing code intelligence](index.md)".

## Manual installation

Generally you do not ever need to perform manual installation of code intelligence. Language servers are automatically downloaded, set up, and configured when you add a repository with a language that we support. If, however, you are using a modified `docker run` command rather than the one we recommend (for example, if you have removed the Docker socket pass-through flag, or if you are running Sourcegraph with an older user-data `#cloud-config`), you can use the following steps to configure code intelligence manually on DigitalOcean:

1.  SSH into the node running your Sourcegraph droplet from the previous step, e.g:

    ```
    ssh -i ~/.ssh/key.pem root@$IP_ADDRESS
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

    Alternatively, you may boot fresh droplet running Sourcegraph and language servers. Simply modify your user data from the previous section as follows:

    ```
    #cloud-config
    repo_update: true
    repo_upgrade: all

    runcmd:
    - mkdir -p /root/.sourcegraph/config
    - mkdir -p /root/.sourcegraph/data
    - [ sh, -c, 'docker network create --driver bridge lsp' ]
    - [ sh, -c, 'docker run -d --publish 80:7080 --publish 443:7443 --network lsp --name sourcegraph --restart unless-stopped --volume /root/.sourcegraph/config:/etc/sourcegraph --volume /root/.sourcegraph/data:/var/opt/sourcegraph sourcegraph/server:2.12.2' ]
    - [ sh, -c, 'docker run -d --network=lsp --name=typescript --restart unless-stopped sourcegraph/codeintel-typescript:latest' ]
    ```

4.  [Update site configuration to point to the language servers](index.md#configure-sourcegraph-to-connect-to-the-language-servers). You can also see a list of all available language servers there.

---

## Next steps

To get code intelligence on your code host and/or code review tool, see the [browser extension documentation](../../../integration/browser_extension.md).
