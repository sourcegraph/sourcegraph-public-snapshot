package com.sourcegraph.cody.config.notification

import com.intellij.util.messages.Topic

interface CodySettingChangeActionNotifier {
    companion object {
        @JvmStatic
        val TOPIC = Topic.create(
            "Sourcegraph Cody + Code Search: Cody AI settings have changed",
            CodySettingChangeActionNotifier::class.java)
    }

    fun afterAction(context: CodySettingChangeContext)
}
