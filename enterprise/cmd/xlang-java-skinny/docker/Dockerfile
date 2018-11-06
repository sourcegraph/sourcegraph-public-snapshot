# Pin to a hash because the only tag is "latest".
FROM mlaccetti/docker-oracle-java8-ubuntu-16.04@sha256:2858f483343958c68d83c817c3db123cf33d0392f1793e10225a51a6b4d398ab
EXPOSE 2088

RUN apt-get -q -y update
# --no-install-recommends breaks the image.
# hadolint ignore=DL3015
RUN apt-get -q -y install unzip=6.0-20ubuntu1 git=1:2.7.4-0ubuntu1.5 maven=3.3.9-3

# Install Gradle binary
# --no-install-recommends breaks the image.
# hadolint ignore=DL3015
RUN apt-get -q -y install wget=1.17.1-1ubuntu1.4
RUN mkdir /gradle
WORKDIR /gradle
ADD https://services.gradle.org/distributions/gradle-4.1-bin.zip ./gradle.zip
RUN unzip gradle.zip
ENV PATH $PATH:/gradle/gradle-4.1/bin
WORKDIR /

# COPY ./artifacts /root/.java-langserver/artifacts

# ENV ANDROID_HOME /android

# ADD https://dl.google.com/android/repository/tools_r25.2.3-linux.zip /tmp/android-tools.zip
# RUN unzip /tmp/android-tools.zip -d /tmp/android-tools
# RUN yes | /tmp/android-tools/tools/bin/sdkmanager --sdk_root="$ANDROID_HOME" "platforms;android-25"
# RUN /tmp/android-tools/tools/bin/sdkmanager --sdk_root="$ANDROID_HOME" "platforms;android-24"
# RUN /tmp/android-tools/tools/bin/sdkmanager --sdk_root="$ANDROID_HOME" "platforms;android-23"
# RUN /tmp/android-tools/tools/bin/sdkmanager --sdk_root="$ANDROID_HOME" "extras;android;m2repository"
# RUN /tmp/android-tools/tools/bin/sdkmanager --sdk_root="$ANDROID_HOME" "extras;google;m2repository"
# RUN /tmp/android-tools/tools/bin/sdkmanager --sdk_root="$ANDROID_HOME" "extras;google;google_play_services"

# COPY ./add-android-support-libs.sh /add-android-support-libs.sh
# RUN ./add-android-support-libs.sh

# ENV ANDROID_JAR_PATH /android/platforms/android-25/android.jar
COPY ./java-language-server.jar /usr/bin

## Leave maven in the image so that we can test in the prod env
# RUN apt-get -qy remove git maven

ENV TINI_VERSION v0.18.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini
ENTRYPOINT ["/tini", "--"]

CMD java $JVM_OPT -jar /usr/bin/java-language-server.jar
