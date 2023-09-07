package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.ui.AccountsPanelFactory.accountsPanel
import com.intellij.collaboration.util.ProgressIndicatorsProvider
import com.intellij.ide.DataManager
import com.intellij.openapi.components.service
import com.intellij.openapi.options.BoundConfigurable
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.DialogPanel
import com.intellij.openapi.ui.setEmptyState
import com.intellij.openapi.util.Disposer
import com.intellij.ui.ColorPanel
import com.intellij.ui.components.JBCheckBox
import com.intellij.ui.dsl.builder.Cell
import com.intellij.ui.dsl.builder.MAX_LINE_LENGTH_NO_WRAP
import com.intellij.ui.dsl.builder.Row
import com.intellij.ui.dsl.builder.bindSelected
import com.intellij.ui.dsl.builder.bindText
import com.intellij.ui.dsl.builder.panel
import com.intellij.ui.dsl.builder.selected
import com.intellij.ui.dsl.builder.toMutableProperty
import com.intellij.ui.dsl.gridLayout.HorizontalAlign
import com.intellij.ui.dsl.gridLayout.VerticalAlign
import com.intellij.util.ui.EmptyIcon
import com.sourcegraph.config.ConfigUtil
import com.sourcegraph.config.PluginSettingChangeActionNotifier
import com.sourcegraph.config.PluginSettingChangeContext
import java.awt.Color

class SettingsConfigurable(private val project: Project) :
    BoundConfigurable(ConfigUtil.SERVICE_DISPLAY_NAME) {
  private val codyProjectSettings = project.service<CodyProjectSettings>()
  private val codyApplicationSettings = service<CodyApplicationSettings>()
  private val settingsModel = SettingsModel()
  private val accountManager = service<CodyAccountManager>()
  private val defaultAccountHolder = project.service<CodyProjectDefaultAccountHolder>()
  private lateinit var accountsModel: CodyAccountListModel
  private lateinit var dialogPanel: DialogPanel

  override fun createPanel(): DialogPanel {
    accountsModel = CodyAccountListModel(project)
    val indicatorsProvider =
        ProgressIndicatorsProvider().also { Disposer.register(disposable!!, it) }
    val detailsProvider =
        CodyAccounDetailsProvider(indicatorsProvider, accountManager, accountsModel)
    dialogPanel = panel {
      group("Authentication") {
        row {
          accountsPanel(
                  accountManager,
                  defaultAccountHolder,
                  accountsModel,
                  detailsProvider,
                  disposable!!,
                  true,
                  EmptyIcon.ICON_16)
              .horizontalAlign(HorizontalAlign.FILL)
              .verticalAlign(VerticalAlign.FILL)
              .also {
                DataManager.registerDataProvider(it.component) { key ->
                  if (CodyAccountsHost.KEY.`is`(key)) accountsModel else null
                }
              }
        }
      }
      group("Cody AI") {
        lateinit var enableCodyCheckbox: Cell<JBCheckBox>
        row {
          enableCodyCheckbox =
              checkBox("Enable Cody")
                  .comment(
                      "Disable this to turn off all AI-based functionality of the plugin, including the Cody chat sidebar and autocomplete",
                      MAX_LINE_LENGTH_NO_WRAP)
                  .bindSelected(settingsModel::isCodyEnabled)
        }
        row {
          checkBox("Enable Cody autocomplete")
              .enabledIf(enableCodyCheckbox.selected)
              .bindSelected(settingsModel::isCodyAutocompleteEnabled)
        }
        row {
          checkBox("Enable debug")
              .comment("Enables debug output visible in the idea.log")
              .enabledIf(enableCodyCheckbox.selected)
              .bindSelected(settingsModel::isCodyDebugEnabled)
        }
        row {
          checkBox("Verbose debug")
              .enabledIf(enableCodyCheckbox.selected)
              .bindSelected(settingsModel::isCodyVerboseDebugEnabled)
        }
        row {
          val enableCustomAutocompleteColor =
              checkBox("Enable custom autocomplete color")
                  .enabledIf(enableCodyCheckbox.selected)
                  .bindSelected(settingsModel::isCustomAutocompleteColorEnabled)
          colorPanel()
              .bind(
                  ColorPanel::getSelectedColor,
                  ColorPanel::setSelectedColor,
                  settingsModel::customAutocompleteColor.toMutableProperty())
              .visibleIf(enableCustomAutocompleteColor.selected)
        }
      }
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
        row {
          checkBox("Do not show the \"No Sourcegraph URL set\" notification for this project")
              .bindSelected(settingsModel::isUrlNotificationDismissed)
        }
      }
    }
    return dialogPanel
  }

  override fun reset() {
    settingsModel.isCodyEnabled = codyApplicationSettings.isCodyEnabled
    settingsModel.isCodyAutocompleteEnabled = codyApplicationSettings.isCodyAutocompleteEnabled
    settingsModel.isCodyDebugEnabled = codyApplicationSettings.isCodyDebugEnabled
    settingsModel.isCodyVerboseDebugEnabled = codyApplicationSettings.isCodyVerboseDebugEnabled
    settingsModel.isUrlNotificationDismissed =
        codyApplicationSettings.isDefaultDotcomAccountNotificationDismissed
    settingsModel.defaultBranchName = codyProjectSettings.defaultBranchName
    settingsModel.remoteUrlReplacements = codyProjectSettings.remoteUrlReplacements
    settingsModel.isCustomAutocompleteColorEnabled =
        codyApplicationSettings.isCustomAutocompleteColorEnabled
    settingsModel.customAutocompleteColor =
        codyApplicationSettings.customAutocompleteColor?.let { Color(it) }
    dialogPanel.reset()
  }

  override fun apply() {
    val bus = project.messageBus
    val publisher = bus.syncPublisher(PluginSettingChangeActionNotifier.TOPIC)

    val oldDefaultAccount = defaultAccountHolder.account
    val oldUrl = oldDefaultAccount?.server?.url ?: ""

    var defaultAccount = accountsModel.defaultAccount
    val newAccessToken = defaultAccount?.let { accountsModel.newCredentials[it] }
    val defaultAccountRemoved = !accountsModel.accounts.contains(defaultAccount)
    if (defaultAccountRemoved) {
      defaultAccount = accountsModel.accounts.getFirstAccountOrNull()
    }
    val defaultAccountChanged = oldDefaultAccount != defaultAccount
    val accessTokenChanged =
        newAccessToken != null || defaultAccountRemoved || defaultAccountChanged

    val newUrl = defaultAccount?.server?.url ?: ""

    val serverUrlChanged = oldUrl != newUrl
    publisher.beforeAction(serverUrlChanged)

    super.apply()
    val context =
        PluginSettingChangeContext(
            serverUrlChanged,
            accessTokenChanged,
            codyApplicationSettings.isCodyEnabled,
            settingsModel.isCodyEnabled,
            codyApplicationSettings.isCodyAutocompleteEnabled,
            settingsModel.isCodyAutocompleteEnabled,
            codyApplicationSettings.isCustomAutocompleteColorEnabled,
            settingsModel.isCustomAutocompleteColorEnabled,
            codyApplicationSettings.customAutocompleteColor,
            settingsModel.customAutocompleteColor?.rgb)
    CodyAuthenticationManager.getInstance().setDefaultAccount(project, defaultAccount)
    accountsModel.defaultAccount = defaultAccount
    codyProjectSettings.defaultBranchName = settingsModel.defaultBranchName
    codyProjectSettings.remoteUrlReplacements = settingsModel.remoteUrlReplacements
    codyApplicationSettings.isCodyEnabled = settingsModel.isCodyEnabled
    codyApplicationSettings.isCodyAutocompleteEnabled = settingsModel.isCodyAutocompleteEnabled
    codyApplicationSettings.isCodyDebugEnabled = settingsModel.isCodyDebugEnabled
    codyApplicationSettings.isCodyVerboseDebugEnabled = settingsModel.isCodyVerboseDebugEnabled
    codyApplicationSettings.isDefaultDotcomAccountNotificationDismissed =
        settingsModel.isUrlNotificationDismissed
    codyApplicationSettings.isCustomAutocompleteColorEnabled =
        settingsModel.isCustomAutocompleteColorEnabled
    codyApplicationSettings.customAutocompleteColor = settingsModel.customAutocompleteColor?.rgb

    publisher.afterAction(context)
  }
}

fun Row.colorPanel() = cell(ColorPanel())
