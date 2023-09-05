package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.ServerAccount
import com.intellij.openapi.util.NlsSafe
import com.intellij.util.xmlb.annotations.Attribute
import com.intellij.util.xmlb.annotations.Property
import com.intellij.util.xmlb.annotations.Tag
import com.intellij.util.xmlb.annotations.Transient
import com.sourcegraph.cody.localapp.LocalAppManager
import com.sourcegraph.config.ConfigUtil
import java.util.UUID

enum class AccountType {
  DOTCOM,
  ENTERPRISE,
  LOCAL_APP
}

@Tag("account")
class CodyAccount(
    @set:Transient @NlsSafe @Attribute("name") override var name: String = "",
    @Property(style = Property.Style.ATTRIBUTE, surroundWithTag = false)
    override val server: SourcegraphServerPath =
        SourcegraphServerPath(LocalAppManager.DEFAULT_LOCAL_APP_URL),
    @Attribute("id") override val id: String,
) : ServerAccount() {

  fun isCodyApp(): Boolean {
    return id == LocalAppManager.LOCAL_APP_ID
  }

  fun getAccountType(): AccountType {
    if (isDotcomAccount()) {
      return AccountType.DOTCOM
    }
    if (id == LocalAppManager.LOCAL_APP_ID) {
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
        id: String = UUID.randomUUID().toString(),
    ): CodyAccount {
      val username =
          if (id == LocalAppManager.LOCAL_APP_ID) {
            LocalAppManager.LOCAL_APP_ID
          } else {
            name
          }
      return CodyAccount(username, server, id)
    }
  }
}

fun Collection<CodyAccount>.getFirstAccountOrNull() =
    this.firstOrNull { it.isDotcomAccount() } ?: this.firstOrNull()
