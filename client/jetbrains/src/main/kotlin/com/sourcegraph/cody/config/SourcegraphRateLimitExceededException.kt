package com.sourcegraph.cody.config

import java.io.IOException

class SourcegraphRateLimitExceededException(message: String): IOException(message)
