FROM mlaccetti/docker-oracle-java8-ubuntu-16.04
MAINTAINER Aaron Leung <the.other.aaron@sourcegraph.com>
EXPOSE 2088

RUN apt-get -qy update
RUN apt-get -qy install unzip git maven

COPY ./artifacts /root/.java-langserver/artifacts

ENV ANDROID_HOME /android

ADD https://dl.google.com/android/repository/tools_r25.2.3-linux.zip /tmp/android-tools.zip
RUN unzip /tmp/android-tools.zip -d /tmp/android-tools
RUN yes | /tmp/android-tools/tools/bin/sdkmanager --sdk_root=$ANDROID_HOME "platforms;android-25"
RUN /tmp/android-tools/tools/bin/sdkmanager --sdk_root=$ANDROID_HOME "platforms;android-24"
RUN /tmp/android-tools/tools/bin/sdkmanager --sdk_root=$ANDROID_HOME "platforms;android-23"
RUN /tmp/android-tools/tools/bin/sdkmanager --sdk_root=$ANDROID_HOME "extras;android;m2repository"
RUN /tmp/android-tools/tools/bin/sdkmanager --sdk_root=$ANDROID_HOME "extras;google;m2repository"
RUN /tmp/android-tools/tools/bin/sdkmanager --sdk_root=$ANDROID_HOME "extras;google;google_play_services"

COPY ./add-android-support-libs.sh /add-android-support-libs.sh
RUN ./add-android-support-libs.sh

ENV ANDROID_JAR_PATH /android/platforms/android-25/android.jar
COPY ./java-language-server.jar /usr/bin

## Leave maven in the image so that we can test in the prod env
# RUN apt-get -qy remove git maven

CMD java $JVM_OPT -jar /usr/bin/java-language-server.jar $LS_OPT
