package com.sourcegraph.cody.auth

import com.intellij.util.Url

interface AuthRequest {
  val serviceName: String

  /** Url that is usually opened in browser where user can accept authorization */
  val authUrlWithParameters: Url
}
