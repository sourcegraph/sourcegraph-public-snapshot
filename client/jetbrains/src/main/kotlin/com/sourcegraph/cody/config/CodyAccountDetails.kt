package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.AccountDetails

class CodyAccountDetails(val username: String, val displayName: String?, val avatarURL: String?) : AccountDetails {
  override val name: String
    get() = displayName ?: username
}
