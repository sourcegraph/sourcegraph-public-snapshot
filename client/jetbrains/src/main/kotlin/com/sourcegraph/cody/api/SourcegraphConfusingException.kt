package com.sourcegraph.cody.api

import java.io.IOException

open class SourcegraphConfusingException : IOException {
  private var myDetails: String? = null

  constructor(message: String?) : super(message)
  constructor(message: String?, cause: Throwable?) : super(message, cause)

  fun setDetails(details: String?) {
    myDetails = details
  }

  override val message: String
    get() =
        if (myDetails == null) {
          super.message!!
        } else {
          """$myDetails

              ${super.message}"""
              .trimIndent()
        }
}
