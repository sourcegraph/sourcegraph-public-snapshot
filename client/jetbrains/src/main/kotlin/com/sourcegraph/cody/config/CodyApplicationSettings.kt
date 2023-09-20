package com.sourcegraph.cody.config

import com.intellij.openapi.components.PersistentStateComponent
import com.intellij.openapi.components.State
import com.intellij.openapi.components.Storage
import com.intellij.openapi.components.service

@State(name = "CodyApplicationSettings", storages = [Storage("cody_application_settings.xml")])
data class CodyApplicationSettings(
    var isCodyEnabled: Boolean = true,
    var isCodyAutocompleteEnabled: Boolean = true,
    var isCodyDebugEnabled: Boolean = false,
    var isCodyVerboseDebugEnabled: Boolean = false,
    var isGetStartedNotificationDismissed: Boolean = false,
    var anonymousUserId: String? = null,
    var isInstallEventLogged: Boolean = false,
    var isCustomAutocompleteColorEnabled: Boolean = false,
    var customAutocompleteColor: Int? = null,
    var blacklistedLanguageIds: List<String> = listOf(),
) : PersistentStateComponent<CodyApplicationSettings> {
  override fun getState(): CodyApplicationSettings = this

  override fun loadState(state: CodyApplicationSettings) {
    this.isCodyEnabled = state.isCodyEnabled
    this.isCodyAutocompleteEnabled = state.isCodyAutocompleteEnabled
    this.isCodyDebugEnabled = state.isCodyDebugEnabled
    this.isCodyVerboseDebugEnabled = state.isCodyVerboseDebugEnabled
    this.isGetStartedNotificationDismissed = state.isGetStartedNotificationDismissed
    this.anonymousUserId = state.anonymousUserId
    this.isInstallEventLogged = state.isInstallEventLogged
    this.isCustomAutocompleteColorEnabled = state.isCustomAutocompleteColorEnabled
    this.customAutocompleteColor = state.customAutocompleteColor
    this.blacklistedLanguageIds = state.blacklistedLanguageIds
  }

  companion object {
    @JvmStatic
    fun getInstance(): CodyApplicationSettings {
      return service<CodyApplicationSettings>()
    }
  }
}
