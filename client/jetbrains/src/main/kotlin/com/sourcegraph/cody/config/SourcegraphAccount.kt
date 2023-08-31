package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.ServerAccount
import com.intellij.openapi.util.NlsSafe
import com.intellij.util.xmlb.annotations.Attribute
import com.intellij.util.xmlb.annotations.Property
import com.intellij.util.xmlb.annotations.Tag
import com.intellij.util.xmlb.annotations.Transient
import com.sourcegraph.cody.localapp.LocalAppManager
import com.sourcegraph.config.ConfigUtil

enum class AccountType {
  DOTCOM,
  ENTERPRISE,
  LOCAL_APP
}

@Tag("account")
class SourcegraphAccount(
    @set:Transient @NlsSafe @Attribute("name") override var name: String = "",
    @Property(style = Property.Style.ATTRIBUTE, surroundWithTag = false)
    override val server: SourcegraphServerPath =
        SourcegraphServerPath(LocalAppManager.DEFAULT_LOCAL_APP_URL),
    @Attribute("id") override val id: String = LocalAppManager.LOCAL_APP_ID,
) : ServerAccount() {

  fun isCodyApp(): Boolean {
    return id == LocalAppManager.LOCAL_APP_ID
  }

  fun getAccountType(): AccountType {
    if (server.url.startsWith(ConfigUtil.DOTCOM_URL)) {
      return AccountType.DOTCOM
    }
    if (id == LocalAppManager.LOCAL_APP_ID) {
      return AccountType.LOCAL_APP
    }
    return AccountType.ENTERPRISE
  }

  override fun toString(): String = "$server/$name"

  companion object {
    fun create(
        name: String,
        server: SourcegraphServerPath,
        id: String,
    ): SourcegraphAccount {
      val username =
          if (id == LocalAppManager.LOCAL_APP_ID) {
            LocalAppManager.LOCAL_APP_ID
          } else {
            name
          }
      return SourcegraphAccount(username, server, id)
    }
  }
}
