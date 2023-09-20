package com.sourcegraph.cody.config

import com.intellij.openapi.util.NlsSafe
import com.intellij.util.xmlb.annotations.Attribute
import com.intellij.util.xmlb.annotations.Property
import com.intellij.util.xmlb.annotations.Tag
import com.sourcegraph.cody.auth.ServerAccount
import com.sourcegraph.cody.localapp.LocalAppManager
import com.sourcegraph.config.ConfigUtil

enum class AccountType {
  DOTCOM,
  ENTERPRISE,
  LOCAL_APP
}

@Tag("account")
data class CodyAccount(
    @NlsSafe @Attribute("name") override var name: String = "",
    @Property(style = Property.Style.ATTRIBUTE, surroundWithTag = false)
    override val server: SourcegraphServerPath =
        SourcegraphServerPath.from(ConfigUtil.DOTCOM_URL, ""),
    @Attribute("id") override val id: String = generateId(),
) : ServerAccount() {

  fun isCodyApp(): Boolean {
    return Companion.isCodyApp(server)
  }

  fun getAccountType(): AccountType {
    if (isDotcomAccount()) {
      return AccountType.DOTCOM
    }
    if (isCodyApp()) {
      return AccountType.LOCAL_APP
    }
    return AccountType.ENTERPRISE
  }

  fun isDotcomAccount() = server.url.startsWith(ConfigUtil.DOTCOM_URL)

  override fun toString(): String = "$server/$name"

  companion object {
    fun create(
        name: String,
        server: SourcegraphServerPath,
        id: String = generateId(),
    ): CodyAccount {
      val username =
          if (isCodyApp(server)) {
            LocalAppManager.LOCAL_APP_ID
          } else {
            name
          }
      return CodyAccount(username, server, id)
    }

    fun isCodyApp(server: SourcegraphServerPath): Boolean {
      return server.url.startsWith(LocalAppManager.DEFAULT_LOCAL_APP_URL)
    }
  }
}

fun Collection<CodyAccount>.getFirstAccountOrNull() =
    this.firstOrNull { it.isDotcomAccount() } ?: this.firstOrNull()
