package com.sourcegraph.cody.config

class SourcegraphStatusCodeException(message: String?, val statusCode: Int) :
    SourcegraphConfusingException(message)
