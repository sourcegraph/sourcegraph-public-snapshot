package com.sourcegraph.cody.config

import com.sourcegraph.cody.auth.AccountDetails

class CodyAccountDetails(val username: String, val displayName: String?, val avatarURL: String?) :
    AccountDetails {
  override val name: String
    get() = displayName ?: username
}
