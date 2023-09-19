package com.sourcegraph.cody.auth

import com.intellij.openapi.util.NlsSafe

interface ServerPath {
  @NlsSafe override fun toString(): String
}
