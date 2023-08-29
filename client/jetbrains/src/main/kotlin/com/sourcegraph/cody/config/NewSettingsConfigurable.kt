package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.ui.AccountsPanelFactory.accountsPanel
import com.intellij.collaboration.util.ProgressIndicatorsProvider
import com.intellij.ide.DataManager
import com.intellij.openapi.components.service
import com.intellij.openapi.options.BoundConfigurable
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.DialogPanel
import com.intellij.openapi.ui.ValidationInfo
import com.intellij.openapi.ui.setEmptyState
import com.intellij.openapi.util.Disposer
import com.intellij.ui.dsl.builder.MAX_LINE_LENGTH_NO_WRAP
import com.intellij.ui.dsl.builder.bindSelected
import com.intellij.ui.dsl.builder.bindText
import com.intellij.ui.dsl.builder.panel
import com.intellij.ui.dsl.gridLayout.HorizontalAlign
import com.intellij.ui.dsl.gridLayout.VerticalAlign
import com.sourcegraph.cody.Icons
import com.sourcegraph.config.CodyApplicationService
import com.sourcegraph.config.ConfigUtil

class NewSettingsConfigurable(private val project: Project) :
    BoundConfigurable(ConfigUtil.SERVICE_DISPLAY_NAME) {
  private var isCodyEnabledCheckBox: Boolean = false
  private var isCodyAutocompleteEnabled: Boolean = false
  private var isCodyDebugEnabled: Boolean = false
  private var isCodyVerboseDebugEnabled: Boolean = false
  private var defaultBranchName: String = "main"
  private var remoteUrlReplacements: String = ""
  private var isUrlNotificationDismissed: Boolean = false
  private var customRequestHeaders: String = ""

  override fun createPanel(): DialogPanel {
    val defaultAccountHolder = project.service<SourcegraphProjectDefaultAccountHolder>()
    val accountManager = service<SourcegraphAccountManager>()
    val settings = CodyApplicationService.getInstance()
    val indicatorsProvider =
        ProgressIndicatorsProvider().also { Disposer.register(disposable!!, it) }
    val accountsModel = SourcegraphAccountListModel(project)
    val detailsProvider =
        SourcegraphAccounDetailsProvider(indicatorsProvider, accountManager, accountsModel)
    return panel {
      group("Authentication") {
        row {
          accountsPanel(
                  accountManager,
                  defaultAccountHolder,
                  accountsModel,
                  detailsProvider,
                  disposable!!,
                  true,
                  Icons.CodyLogo)
              .horizontalAlign(HorizontalAlign.FILL)
              .verticalAlign(VerticalAlign.FILL)
              .also {
                DataManager.registerDataProvider(it.component) { key ->
                  if (SourcegraphAccountsHost.KEY.`is`(key)) accountsModel else null
                }
              }
        }
        row {
          textField()
              .label("Custom request headers:")
              .comment(
                  """Any custom headers to send with every request to Sourcegraph.<br>
                  |Use any number of pairs: "header1, value1, header2, value2, ...".<br>
                  |Whitespace around commas doesn't matter.
              """
                      .trimMargin(),
                  MAX_LINE_LENGTH_NO_WRAP)
              .horizontalAlign(HorizontalAlign.FILL)
              .bindText(::customRequestHeaders)
              .applyToComponent {
                this.setEmptyState("Client-ID, client-one, X-Extra, some metadata")
              }
              .validation {
                if (it.getText().isEmpty()) {
                  return@validation null
                }
                val pairs: Array<String> =
                    it.getText().split(",".toRegex()).dropLastWhile { it.isEmpty() }.toTypedArray()
                if (pairs.size % 2 != 0) {
                  return@validation ValidationInfo(
                      "Must be a comma-separated list of string pairs", it)
                }
                var i = 0
                while (i < pairs.size) {
                  val headerName = pairs[i].trim { it <= ' ' }
                  if (!headerName.matches("[\\w-]+".toRegex())) {
                    return@validation ValidationInfo("Invalid HTTP header name: $headerName", it)
                  }
                  i += 2
                }
                return@validation null
              }
        }
      }
      group("Cody AI") {
        row {
          checkBox("Enable Cody")
              .comment(
                  "Disable this to turn off all AI-based functionality of the plugin, including the Cody chat sidebar and autocomplete",
                  MAX_LINE_LENGTH_NO_WRAP)
              .bindSelected(::isCodyEnabledCheckBox)
        }
        row { checkBox("Enable Cody autocomplete").bindSelected(::isCodyAutocompleteEnabled) }
        row {
          checkBox("Enable debug")
              .comment("Enables debug output visible in the idea.log")
              .bindSelected(::isCodyDebugEnabled)
        }
        row { checkBox("Verbose debug").bindSelected(::isCodyVerboseDebugEnabled) }
      }
      group("Code search") {
        row {
          textField()
              .label("Default branch name:")
              .comment("The branch to use if the current branch is not yet pushed")
              .horizontalAlign(HorizontalAlign.FILL)
              .bindText(::defaultBranchName)
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
              .bindText(::remoteUrlReplacements)
              .applyToComponent {
                this.setEmptyState("search1, replacement1, search2, replacement2, ...")
              }
        }
        row {
          checkBox("Do not show the \"No Sourcegraph URL set\" notification for this project")
              .bindSelected(::isUrlNotificationDismissed)
        }
      }
    }
  }
}
