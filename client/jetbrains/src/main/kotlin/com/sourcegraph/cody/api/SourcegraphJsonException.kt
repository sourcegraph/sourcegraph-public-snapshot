package com.sourcegraph.cody.api

class SourcegraphJsonException(message: String, cause: Throwable) :
    SourcegraphConfusingException(message, cause)
