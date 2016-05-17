#
# Docker image for srclib-java
#

#
# Dependencies
#
FROM maven:3-jdk-8

RUN apt-get update -y
# See https://code.google.com/p/android/issues/detail?id=82711
RUN apt-get install -qq git make -y lib32z1 lib32ncurses5 lib32stdc++6

ENV SRCLIBPATH /srclib/toolchains

# Gradle
ENV GRADLE_VERSION 2.10
WORKDIR /usr/lib
RUN wget -q https://downloads.gradle.org/distributions/gradle-${GRADLE_VERSION}-bin.zip && unzip "gradle-${GRADLE_VERSION}-bin.zip" && ln -s "/usr/lib/gradle-${GRADLE_VERSION}/bin/gradle" /usr/bin/gradle && rm "gradle-${GRADLE_VERSION}-bin.zip"

# Install Android SDK
RUN cd /opt && curl http://dl.google.com/android/android-sdk_r24.3.4-linux.tgz | tar xz
ENV ANDROID_HOME /opt/android-sdk-linux
ENV PATH $PATH:$ANDROID_HOME/tools:$ANDROID_HOME/platform-tools
RUN echo y | android update sdk --filter platform-tools,build-tools-23.0.3,build-tools-23.0.2,build-tools-23.0.1,build-tools-23,build-tools-22.0.1,build-tools-22,build-tools-21.1.2,build-tools-21.1.1,build-tools-21.1,build-tools-21.0.2,build-tools-21.0.1,build-tools-21,build-tools-20,build-tools-19.1,build-tools-19.0.3,build-tools-19.0.2,build-tools-19.0.1,build-tools-19,build-tools-18.1.1,build-tools-18.1,build-tools-18.0.1,build-tools-17,android-23,android-22,android-21,android-20,android-19,android-18,android-17,android-16,android-15,android-14,extra-android-support,extra-android-m2repository,extra-google-m2repository --no-ui --force --all

# Add special JDK
RUN mkdir -p $SRCLIBPATH/sourcegraph.com/sourcegraph/srclib-java
WORKDIR $SRCLIBPATH/sourcegraph.com/sourcegraph/srclib-java
RUN curl https://srclib-support.s3-us-west-2.amazonaws.com/srclib-java/build/bundled-jdk1.8.0_45.tar.gz | tar xz

#
# Install srclib-java executable
#
ARG TOOLCHAIN_URL
ADD $TOOLCHAIN_URL /toolchain/t
RUN (cd /toolchain && tar xfz t && rm t && mv * /toolchain/t) || true
RUN cp -a /toolchain/t/. $SRCLIBPATH/sourcegraph.com/sourcegraph/srclib-java && rm -rf /toolchain/t
# Install
RUN make

#
# Install srclib binary (assumes this has been built on the Docker host)
#
ADD ./srclib /bin/srclib

# Run environment
ENV GOPATH /drone
ENTRYPOINT srclib config && srclib make
