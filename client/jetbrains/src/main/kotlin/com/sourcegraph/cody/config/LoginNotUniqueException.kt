package com.sourcegraph.cody.config

internal class LoginNotUniqueException(val login: String) : RuntimeException()
