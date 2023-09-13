package com.sourcegraph.cody.auth

import com.intellij.openapi.util.NlsSafe

interface AccountDetails {
  @get:NlsSafe
  val name: String
}
