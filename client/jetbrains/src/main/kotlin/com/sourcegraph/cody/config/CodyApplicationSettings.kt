package com.sourcegraph.cody.config

import com.intellij.openapi.components.PersistentStateComponent
import com.intellij.openapi.components.State
import com.intellij.openapi.components.Storage
import com.intellij.openapi.components.service

@State(name = "CodyApplicationSettings", storages = [Storage("cody_application_settings.xml")])
data class CodyApplicationSettings(
    var isCodyEnabled: Boolean = false,
    var isCodyAutocompleteEnabled: Boolean = false,
    var isCodyDebugEnabled: Boolean = false,
    var isCodyVerboseDebugEnabled: Boolean = false,
    var isDefaultDotcomAccountNotificationDismissed: Boolean = false,
    var anonymousUserId: String? = null,
    var isInstallEventLogged: Boolean = false,
    var lastUpdateNotificationPluginVersion: String? = null,
    var isCustomAutocompleteColorEnabled: Boolean = false,
    var customAutocompleteColor: Int? = null,
) : PersistentStateComponent<CodyApplicationSettings> {
  override fun getState(): CodyApplicationSettings = this

  override fun loadState(state: CodyApplicationSettings) {
    this.isCodyEnabled = state.isCodyEnabled
    this.isCodyAutocompleteEnabled = state.isCodyAutocompleteEnabled
    this.isCodyDebugEnabled = state.isCodyDebugEnabled
    this.isCodyVerboseDebugEnabled = state.isCodyVerboseDebugEnabled
    this.isDefaultDotcomAccountNotificationDismissed =
        state.isDefaultDotcomAccountNotificationDismissed
    this.anonymousUserId = state.anonymousUserId
    this.isInstallEventLogged = state.isInstallEventLogged
    this.lastUpdateNotificationPluginVersion = state.lastUpdateNotificationPluginVersion
    this.isCustomAutocompleteColorEnabled = state.isCustomAutocompleteColorEnabled
    this.customAutocompleteColor = state.customAutocompleteColor
  }

  companion object {
    @JvmStatic
    fun getInstance(): CodyApplicationSettings {
      return service<CodyApplicationSettings>()
    }
  }
}
