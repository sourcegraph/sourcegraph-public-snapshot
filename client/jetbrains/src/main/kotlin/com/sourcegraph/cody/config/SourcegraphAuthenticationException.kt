package com.sourcegraph.cody.config

import java.io.IOException

class SourcegraphAuthenticationException: IOException {

    constructor(): super()
    constructor(message: String): super(message)
    constructor(message: String, cause: Throwable): super(message, cause)
    constructor(cause: Throwable): super(cause)
}
