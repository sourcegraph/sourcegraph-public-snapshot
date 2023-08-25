package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.ui.AccountsPanelFactory.accountsPanel
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
import com.sourcegraph.cody.Icons
import com.sourcegraph.config.CodyApplicationService
import com.sourcegraph.config.ConfigUtil

class NewSettingsConfigurable(private val project: Project) :
    BoundConfigurable(ConfigUtil.SERVICE_DISPLAY_NAME) {
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
          .resizableRow()
    }
  }
}
