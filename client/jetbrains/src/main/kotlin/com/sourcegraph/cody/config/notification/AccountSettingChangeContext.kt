package com.sourcegraph.cody.config.notification

class AccountSettingChangeContext(
    val serverUrlChanged: Boolean = false,
    val accessTokenChanged: Boolean = false
)
