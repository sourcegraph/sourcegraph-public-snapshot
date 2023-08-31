package com.sourcegraph.cody.config

import com.intellij.openapi.components.PersistentStateComponent
import com.intellij.openapi.components.State
import com.intellij.openapi.components.Storage
import com.intellij.openapi.components.service
import com.intellij.openapi.project.Project

@State(name = "Settings", storages = [Storage("sourcegraph_settings.xml")])
data class SettingsModel(
    var isCodyEnabled: Boolean = false,
    var isCodyAutocompleteEnabled: Boolean = false,
    var isCodyDebugEnabled: Boolean = false,
    var isCodyVerboseDebugEnabled: Boolean = false,
    var defaultBranchName: String = "main",
    var remoteUrlReplacements: String = "",
    var isUrlNotificationDismissed: Boolean = false,
    var customRequestHeaders: String = ""
) : PersistentStateComponent<SettingsModel> {
  override fun getState(): SettingsModel = this

  override fun loadState(state: SettingsModel) {
    this.isCodyEnabled = state.isCodyEnabled
    this.isCodyAutocompleteEnabled = state.isCodyAutocompleteEnabled
    this.isCodyDebugEnabled = state.isCodyDebugEnabled
    this.isCodyVerboseDebugEnabled = state.isCodyVerboseDebugEnabled
    this.defaultBranchName = state.defaultBranchName
    this.remoteUrlReplacements = state.remoteUrlReplacements
    this.isUrlNotificationDismissed = state.isUrlNotificationDismissed
    this.customRequestHeaders = state.customRequestHeaders
  }

  companion object {
    fun getInstance(project: Project): SettingsModel {
      return project.service<SettingsModel>()
    }
  }
}
