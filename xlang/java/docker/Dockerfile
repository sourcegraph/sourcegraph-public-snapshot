FROM mlaccetti/docker-oracle-java8-ubuntu-16.04
MAINTAINER Aaron Leung <the.other.aaron@sourcegraph.com>
EXPOSE 2088

COPY ./artifacts /root/.java-langserver/artifacts
COPY ./java-language-server.jar /usr/bin

CMD java $JVM_OPT -jar /usr/bin/java-language-server.jar $LANGSERVER_OPT
