package com.sourcegraph.cody.config

import com.intellij.openapi.actionSystem.DataKey

interface CodyAccountsHost {
  fun addAccount(server: SourcegraphServerPath, login: String, displayName: String?, token: String)

  fun isAccountUnique(login: String, server: SourcegraphServerPath): Boolean

  companion object {
    val KEY: DataKey<CodyAccountsHost> = DataKey.create("CodyAccountsHots")
  }
}
