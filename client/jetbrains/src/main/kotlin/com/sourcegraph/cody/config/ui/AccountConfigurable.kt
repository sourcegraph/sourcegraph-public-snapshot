package com.sourcegraph.cody.config.ui

import com.intellij.collaboration.util.ProgressIndicatorsProvider
import com.intellij.ide.DataManager
import com.intellij.openapi.components.service
import com.intellij.openapi.options.BoundConfigurable
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.DialogPanel
import com.intellij.openapi.util.Disposer
import com.intellij.ui.dsl.builder.panel
import com.intellij.ui.dsl.gridLayout.HorizontalAlign
import com.intellij.ui.dsl.gridLayout.VerticalAlign
import com.intellij.util.ui.EmptyIcon
import com.sourcegraph.cody.auth.ui.customAccountsPanel
import com.sourcegraph.cody.config.CodyAccountDetailsProvider
import com.sourcegraph.cody.config.CodyAccountListModel
import com.sourcegraph.cody.config.CodyAccountManager
import com.sourcegraph.cody.config.CodyAccountsHost
import com.sourcegraph.cody.config.CodyAuthenticationManager
import com.sourcegraph.cody.config.CodyProjectActiveAccountHolder
import com.sourcegraph.cody.config.getFirstAccountOrNull
import com.sourcegraph.cody.config.notification.AccountSettingChangeActionNotifier
import com.sourcegraph.cody.config.notification.AccountSettingChangeContext
import com.sourcegraph.config.ConfigUtil
import java.awt.Dimension

class AccountConfigurable(val project: Project) :
    BoundConfigurable(ConfigUtil.SOURCEGRAPH_DISPLAY_NAME) {
  private val accountManager = service<CodyAccountManager>()
  private val accountsModel = CodyAccountListModel(project)
  private val activeAccountHolder = project.service<CodyProjectActiveAccountHolder>()
  private lateinit var dialogPanel: DialogPanel

  override fun createPanel(): DialogPanel {
    dialogPanel = panel {
      group("Authentication") {
        row {
          customAccountsPanel(
                  accountManager,
                  activeAccountHolder,
                  accountsModel,
                  CodyAccountDetailsProvider(
                      ProgressIndicatorsProvider().also { Disposer.register(disposable!!, it) },
                      accountManager,
                      accountsModel),
                  disposable!!,
                  true,
                  EmptyIcon.ICON_16) {
                    it.copy(server = it.server.copy())
                  }
              .horizontalAlign(HorizontalAlign.FILL)
              .verticalAlign(VerticalAlign.FILL)
              .applyToComponent { this.preferredSize = Dimension(Int.MAX_VALUE, 200) }
              .also {
                DataManager.registerDataProvider(it.component) { key ->
                  if (CodyAccountsHost.KEY.`is`(key)) accountsModel else null
                }
              }
        }
      }
    }
    return dialogPanel
  }

  override fun reset() {
    dialogPanel.reset()
  }

  override fun apply() {
    val bus = project.messageBus
    val publisher = bus.syncPublisher(AccountSettingChangeActionNotifier.TOPIC)

    val oldActiveAccount = activeAccountHolder.account
    val oldUrl = oldActiveAccount?.server?.url ?: ""

    var activeAccount = accountsModel.activeAccount
    val activeAccountRemoved = !accountsModel.accounts.contains(activeAccount)
    if (activeAccountRemoved || activeAccount == null) {
      activeAccount = accountsModel.accounts.getFirstAccountOrNull()
    }
    val newAccessToken = activeAccount?.let { accountsModel.newCredentials[it] }

    val activeAccountChanged = oldActiveAccount != activeAccount
    val accessTokenChanged = newAccessToken != null || activeAccountRemoved || activeAccountChanged

    val newUrl = activeAccount?.server?.url ?: ""

    val serverUrlChanged = oldUrl != newUrl

    publisher.beforeAction(serverUrlChanged)
    super.apply()
    val context = AccountSettingChangeContext(serverUrlChanged, accessTokenChanged)
    CodyAuthenticationManager.getInstance().setActiveAccount(project, activeAccount)
    accountsModel.activeAccount = activeAccount
    publisher.afterAction(context)
  }
}
