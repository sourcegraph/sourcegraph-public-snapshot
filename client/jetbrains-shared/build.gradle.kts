import org.jetbrains.kotlin.gradle.tasks.KotlinCompile

fun properties(key: String) = project.findProperty(key).toString()

plugins {
    id("java")
    // Dependencies are locked at this version to work with JDK 11 on CI.
    id("org.jetbrains.intellij") version "1.13.3"
}

repositories {
    mavenCentral()
}


intellij {
    version.set(properties("platformVersion"))
}

tasks {
    // Set the JVM compatibility versions
    properties("javaVersion").let {
        withType<JavaCompile> {
            sourceCompatibility = it
            targetCompatibility = it
        }
        withType<KotlinCompile> {
            kotlinOptions.jvmTarget = it
        }
    }
}

