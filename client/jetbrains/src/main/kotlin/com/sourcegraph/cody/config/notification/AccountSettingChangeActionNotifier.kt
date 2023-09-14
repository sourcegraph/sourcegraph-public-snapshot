package com.sourcegraph.cody.config.notification

import com.intellij.util.messages.Topic

interface AccountSettingChangeActionNotifier {
  companion object {
    @JvmStatic
    val TOPIC =
        Topic.create(
            "Sourcegraph Cody + Code Search plugin settings have changed",
            AccountSettingChangeActionNotifier::class.java)
  }

  fun beforeAction(serverUrlChanged: Boolean)

  fun afterAction(context: AccountSettingChangeContext)
}
