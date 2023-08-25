package com.sourcegraph.cody.config

import java.io.IOException

open class SourcegraphConfusingException : IOException {
  private var myDetails: String? = null

  constructor()
  constructor(message: String?) : super(message)
  constructor(message: String?, cause: Throwable?) : super(message, cause)
  constructor(cause: Throwable?) : super(cause)

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
