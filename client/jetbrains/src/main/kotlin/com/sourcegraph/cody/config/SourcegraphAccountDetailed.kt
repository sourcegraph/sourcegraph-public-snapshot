package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.AccountDetails

class SourcegraphAccountDetailed(val username: String, val avatarURL: String?) : AccountDetails {
  override val name: String
    get() = username
}
