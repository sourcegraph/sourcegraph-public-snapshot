# Package com.sourcegraph.cody.auth

This package contains API to manage user accounts.
Most of the code was copied here from package `com.intellij.collaboration.auth`, because it looks like it is a public API,
but it was changing without any sign of deprecation, and it broke our code. That's why we decided to copy it to our sources.
The code can be updated to the same state as in the original package in the new versions of IntelliJ IDEA
