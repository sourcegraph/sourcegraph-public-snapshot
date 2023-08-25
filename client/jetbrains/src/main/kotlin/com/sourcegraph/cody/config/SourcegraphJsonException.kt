package com.sourcegraph.cody.config

class SourcegraphJsonException : SourcegraphConfusingException {
  constructor() : super()
  constructor(message: String) : super(message)
  constructor(message: String, cause: Throwable) : super(message, cause)
  constructor(cause: Throwable) : super(cause)
}
