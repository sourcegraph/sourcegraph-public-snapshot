package com.sourcegraph.cody.config.ui

import com.intellij.openapi.components.service
import com.intellij.openapi.options.BoundConfigurable
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.DialogPanel
import com.intellij.openapi.ui.setEmptyState
import com.intellij.ui.dsl.builder.MAX_LINE_LENGTH_NO_WRAP
import com.intellij.ui.dsl.builder.bindText
import com.intellij.ui.dsl.builder.panel
import com.intellij.ui.dsl.gridLayout.HorizontalAlign
import com.sourcegraph.cody.config.CodyProjectSettings
import com.sourcegraph.cody.config.SettingsModel
import com.sourcegraph.config.ConfigUtil

class CodeSearchConfigurable(val project: Project) :
    BoundConfigurable(ConfigUtil.CODE_SEARCH_DISPLAY_NAME) {
  private lateinit var dialogPanel: DialogPanel
  private val settingsModel = SettingsModel()
  private val codyProjectSettings = project.service<CodyProjectSettings>()
  override fun createPanel(): DialogPanel {
    dialogPanel = panel {
      group("Code search") {
        row {
          textField()
              .label("Default branch name:")
              .comment("The branch to use if the current branch is not yet pushed")
              .horizontalAlign(HorizontalAlign.FILL)
              .bindText(settingsModel::defaultBranchName)
              .applyToComponent {
                this.setEmptyState("main")
                toolTipText = "Usually \"main\" or \"master\", but can be any name"
              }
        }
        row {
          textField()
              .label("Remote URL replacements:")
              .comment(
                  """You can replace specified strings in your repo's remote URL. <br>
                      |Use any number of pairs: "search1, replacement1, search2, replacement2, ...". <br>
                      |Pairs are replaced from left to right. Whitespace around commas doesn't matter.
                  """
                      .trimMargin(),
                  MAX_LINE_LENGTH_NO_WRAP)
              .horizontalAlign(HorizontalAlign.FILL)
              .bindText(settingsModel::remoteUrlReplacements)
              .applyToComponent {
                this.setEmptyState("search1, replacement1, search2, replacement2, ...")
              }
        }
      }
    }
    return dialogPanel
  }

  override fun reset() {
    settingsModel.defaultBranchName = codyProjectSettings.defaultBranchName
    settingsModel.remoteUrlReplacements = codyProjectSettings.remoteUrlReplacements
    dialogPanel.reset()
  }

  override fun apply() {
    super.apply()
    codyProjectSettings.defaultBranchName = settingsModel.defaultBranchName
    codyProjectSettings.remoteUrlReplacements = settingsModel.remoteUrlReplacements
  }
}
