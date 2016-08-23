# Java Language Processor Proxy Server

## Prerequisites

 - JDK 1.8.x (Oracle JDK or OpenJDK)
 - Set `$JAVA_HOME` to your JDK directory:
   - OS X: ```export JAVA_HOME=`/usr/libexec/java_home```

## Building and pushing Docker image

```bash
./build.sh
```

## Running

Start `langprocessor-java` (cd to this directory first):

```bash
rego .
```

Clone the Java server:

```bash
git clone https://github.com/alexsaveliev/java-language-processor && cd java-language-processor
```

Build:

```bash
./gradlew assemble
```

Run:

```bash
java -cp $JAVA_HOME/lib/tools.jar:./build/libs/java-language-processor-0.0.1-SNAPSHOT.jar org.springframework.boot.loader.JarLauncher --server.port=4143
```
