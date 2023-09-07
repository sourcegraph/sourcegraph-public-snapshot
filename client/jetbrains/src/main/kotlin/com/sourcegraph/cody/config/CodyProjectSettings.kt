package com.sourcegraph.cody.config

import com.intellij.openapi.components.PersistentStateComponent
import com.intellij.openapi.components.State
import com.intellij.openapi.components.Storage
import com.intellij.openapi.components.service
import com.intellij.openapi.project.Project
import com.intellij.util.xmlb.annotations.Transient
import com.sourcegraph.find.Search

@State(name = "CodyProjectSettings", storages = [Storage("cody_project_settings.xml")])
data class CodyProjectSettings(
    var defaultBranchName: String = "main",
    var remoteUrlReplacements: String = "",
    var lastSearchQuery: String? = null,
    var lastSearchCaseSensitive: Boolean = false,
    var lastSearchPatternType: String? = null,
    var lastSearchContextSpec: String? = null,
) : PersistentStateComponent<CodyProjectSettings> {
  override fun getState(): CodyProjectSettings = this

  override fun loadState(state: CodyProjectSettings) {
    this.defaultBranchName = state.defaultBranchName
    this.remoteUrlReplacements = state.remoteUrlReplacements
    this.lastSearchQuery = state.lastSearchQuery
    this.lastSearchCaseSensitive = state.lastSearchCaseSensitive
    this.lastSearchPatternType = state.lastSearchPatternType
    this.lastSearchContextSpec = state.lastSearchContextSpec
  }

  @Transient
  fun getLastSearch(): Search? {
    return if (lastSearchQuery == null) {
      null
    } else {
      Search(lastSearchQuery, lastSearchCaseSensitive, lastSearchPatternType, lastSearchContextSpec)
    }
  }

  @Transient
  fun setLastSearch(lastSearch: Search) {
    this.lastSearchQuery = if (lastSearch.query != null) lastSearch.query else ""
    this.lastSearchCaseSensitive = lastSearch.isCaseSensitive
    this.lastSearchPatternType =
        if (lastSearch.patternType != null) lastSearch.patternType else "literal"
    this.lastSearchContextSpec =
        if (lastSearch.selectedSearchContextSpec != null) lastSearch.selectedSearchContextSpec
        else "global"
  }

  companion object {
    @JvmStatic
    fun getInstance(project: Project): CodyProjectSettings {
      return project.service<CodyProjectSettings>()
    }
  }
}
