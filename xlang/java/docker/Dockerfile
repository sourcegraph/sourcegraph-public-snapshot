FROM mlaccetti/docker-oracle-java8-ubuntu-16.04
MAINTAINER Aaron Leung <the.other.aaron@sourcegraph.com>
EXPOSE 2088

COPY ./artifacts /root/.java-langserver/artifacts

COPY ./android-sdk-jars/platforms/android-25 /root/android-sdk/platforms/android-25
COPY ./android-sdk-jars/extras /root/android-sdk/extras
ENV ANDROID_JAR_PATH /root/android-sdk/platforms/android-25/android.jar
ENV ANDROID_SUPPORT_HOME /root/android-sdk/extras/android/m2repository/com/android/support
ENV ANDROID_GOOGLE_HOME /root/android-sdk/extras/google/m2repository/com/google/android/support/wearable

COPY ./java-language-server.jar /usr/bin

CMD java $JVM_OPT -jar /usr/bin/java-language-server.jar
