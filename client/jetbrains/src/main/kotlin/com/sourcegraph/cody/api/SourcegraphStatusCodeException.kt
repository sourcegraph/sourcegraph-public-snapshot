package com.sourcegraph.cody.api

class SourcegraphStatusCodeException(message: String?, val statusCode: Int) :
    SourcegraphConfusingException(message)
