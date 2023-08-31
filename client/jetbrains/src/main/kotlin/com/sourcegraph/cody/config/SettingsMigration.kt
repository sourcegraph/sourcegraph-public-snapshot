package com.sourcegraph.cody.config

import com.intellij.collaboration.async.CompletableFutureUtil.submitIOTask
import com.intellij.collaboration.async.CompletableFutureUtil.successOnEdt
import com.intellij.ide.util.RunOnceUtil
import com.intellij.openapi.application.ModalityState
import com.intellij.openapi.components.service
import com.intellij.openapi.diagnostic.logger
import com.intellij.openapi.progress.EmptyProgressIndicator
import com.intellij.openapi.progress.ProgressManager
import com.intellij.openapi.project.DumbAware
import com.intellij.openapi.project.Project
import com.intellij.openapi.startup.StartupActivity
import com.intellij.util.containers.orNull
import com.sourcegraph.cody.localapp.LocalAppManager
import com.sourcegraph.config.ConfigUtil
import com.sourcegraph.config.SettingsComponent
import java.util.UUID
import java.util.concurrent.CompletableFuture

class SettingsMigration : StartupActivity, DumbAware {

  private val accountManager = service<SourcegraphAccountManager>()
  override fun runActivity(project: Project) {
    val settingsModel = SettingsModel.getInstance(project)
    val defaultAccountHolder = project.service<SourcegraphProjectDefaultAccountHolder>()
    RunOnceUtil.runOnceForProject(project, UUID.randomUUID().toString()) {
      migrateAccounts(project, defaultAccountHolder)
      migrateSettings(project, settingsModel)
    }
  }

  private fun migrateAccounts(
      project: Project,
      defaultAccountHolder: SourcegraphProjectDefaultAccountHolder
  ) {
    val requestExecutorFactory = SourcegraphApiRequestExecutor.Factory.getInstance()
    val instanceType = ConfigUtil.getInstanceType(project)
    migrateDotcomAccount(project, instanceType, requestExecutorFactory, defaultAccountHolder)
    migrateEnterpriseAccount(project, instanceType, requestExecutorFactory, defaultAccountHolder)
    migrateCodyAccount(instanceType, requestExecutorFactory, defaultAccountHolder)
  }

  private fun migrateDotcomAccount(
      project: Project,
      instanceType: SettingsComponent.InstanceType,
      requestExecutorFactory: SourcegraphApiRequestExecutor.Factory,
      defaultAccountHolder: SourcegraphProjectDefaultAccountHolder
  ) {
    val dotcomAccessToken = ConfigUtil.getDotComAccessToken(project)
    if (!dotcomAccessToken.isNullOrEmpty()) {
      val server = SourcegraphServerPath(ConfigUtil.DOTCOM_URL)
      val shouldSetAccountAsDefault = instanceType == SettingsComponent.InstanceType.DOTCOM
      if (shouldSetAccountAsDefault) {
        addAsDefaultAccountIfUnique(
            dotcomAccessToken,
            server,
            requestExecutorFactory,
            EmptyProgressIndicator(ModalityState.NON_MODAL),
            defaultAccountHolder)
      } else {
        addAccountIfUnique(
            dotcomAccessToken,
            server,
            requestExecutorFactory,
            EmptyProgressIndicator(ModalityState.NON_MODAL))
      }
    }
  }

  private fun migrateEnterpriseAccount(
      project: Project,
      instanceType: SettingsComponent.InstanceType,
      requestExecutorFactory: SourcegraphApiRequestExecutor.Factory,
      defaultAccountHolder: SourcegraphProjectDefaultAccountHolder
  ) {
    val enterpriseAccessToken = ConfigUtil.getEnterpriseAccessToken(project)
    if (!enterpriseAccessToken.isNullOrEmpty()) {
      val enterpriseUrl = ConfigUtil.getEnterpriseUrl(project)
      runCatching { SourcegraphServerPath.from(enterpriseUrl) }
          .fold({
            val shouldSetAccountAsDefault =
                instanceType == SettingsComponent.InstanceType.ENTERPRISE
            if (shouldSetAccountAsDefault) {
              addAsDefaultAccountIfUnique(
                  enterpriseAccessToken,
                  it,
                  requestExecutorFactory,
                  EmptyProgressIndicator(ModalityState.NON_MODAL),
                  defaultAccountHolder)
            } else {
              addAccountIfUnique(
                  enterpriseAccessToken,
                  it,
                  requestExecutorFactory,
                  EmptyProgressIndicator(ModalityState.NON_MODAL))
            }
          }) {
            LOG.warn("Unable to parse enterprise server url: '$enterpriseUrl'", it)
          }
    }
  }

  private fun migrateCodyAccount(
      instanceType: SettingsComponent.InstanceType,
      requestExecutorFactory: SourcegraphApiRequestExecutor.Factory,
      defaultAccountHolder: SourcegraphProjectDefaultAccountHolder
  ) {
    val localAppAccessToken = LocalAppManager.getLocalAppAccessToken().orNull()
    if (LocalAppManager.isLocalAppInstalled() && localAppAccessToken != null) {
      val codyUrl = LocalAppManager.getLocalAppUrl()
      runCatching { SourcegraphServerPath.from(codyUrl) }
          .fold({
            val shouldSetAccountAsDefault = instanceType == SettingsComponent.InstanceType.LOCAL_APP
            if (shouldSetAccountAsDefault) {
              addAsDefaultAccountIfUnique(
                  localAppAccessToken,
                  it,
                  requestExecutorFactory,
                  EmptyProgressIndicator(ModalityState.NON_MODAL),
                  defaultAccountHolder,
                  LocalAppManager.LOCAL_APP_ID)
            } else {
              addAccountIfUnique(
                  localAppAccessToken,
                  it,
                  requestExecutorFactory,
                  EmptyProgressIndicator(ModalityState.NON_MODAL),
                  LocalAppManager.LOCAL_APP_ID)
            }
          }) {
            LOG.warn("Unable to parse local app server url: '$localAppAccessToken'", it)
          }
    }
  }

  private fun addAccountIfUnique(
      accessToken: String,
      server: SourcegraphServerPath,
      requestExecutorFactory: SourcegraphApiRequestExecutor.Factory,
      progressIndicator: EmptyProgressIndicator,
      id: String = UUID.randomUUID().toString(),
  ) {
    loadUserDetails(requestExecutorFactory, accessToken, progressIndicator, server) {
      addAccount(SourcegraphAccount.create(it.name, server, id), accessToken)
    }
  }

  private fun addAsDefaultAccountIfUnique(
      accessToken: String,
      server: SourcegraphServerPath,
      requestExecutorFactory: SourcegraphApiRequestExecutor.Factory,
      progressIndicator: EmptyProgressIndicator,
      projectDefaultAccountHolder: SourcegraphProjectDefaultAccountHolder,
      id: String = UUID.randomUUID().toString(),
  ) {
    loadUserDetails(requestExecutorFactory, accessToken, progressIndicator, server) {
      val sourcegraphAccount = SourcegraphAccount.create(it.name, server, id)
      addAccount(sourcegraphAccount, accessToken)
      projectDefaultAccountHolder.account = sourcegraphAccount
    }
  }

  private fun loadUserDetails(
      requestExecutorFactory: SourcegraphApiRequestExecutor.Factory,
      accessToken: String,
      progressIndicator: EmptyProgressIndicator,
      server: SourcegraphServerPath,
      onSuccess: (SourcegraphAccountDetailed) -> Unit
  ): CompletableFuture<Unit> =
      service<ProgressManager>()
          .submitIOTask(progressIndicator) {
            runCatching {
              SourcegraphSecurityUtil.loadCurrentUserDetails(
                  requestExecutorFactory.create(accessToken), progressIndicator, server)
            }
          }
          .successOnEdt(progressIndicator.modalityState) {
            it.fold(onSuccess) {
              LOG.warn("Unable to load user details for '${server.url}' account", it)
            }
          }

  private fun addAccount(sourcegraphAccount: SourcegraphAccount, token: String) {
    if (isAccountUnique(sourcegraphAccount)) {
      accountManager.updateAccount(sourcegraphAccount, token)
    }
  }

  private fun isAccountUnique(sourcegraphAccount: SourcegraphAccount): Boolean {
    return accountManager.accounts.none {
      it.name == sourcegraphAccount.name && it.server == sourcegraphAccount.server
    }
  }

  private fun migrateSettings(project: Project, settingsModel: SettingsModel) {
    settingsModel.isCodyEnabled = ConfigUtil.isCodyEnabled()
    settingsModel.isCodyAutocompleteEnabled = ConfigUtil.isCodyAutocompleteEnabled()
    settingsModel.isCodyDebugEnabled = ConfigUtil.isCodyDebugEnabled()
    settingsModel.isCodyVerboseDebugEnabled = ConfigUtil.isCodyVerboseDebugEnabled()
    settingsModel.defaultBranchName = ConfigUtil.getDefaultBranchName(project)
    settingsModel.remoteUrlReplacements = ConfigUtil.getRemoteUrlReplacements(project)
    settingsModel.customRequestHeaders = ConfigUtil.getCustomRequestHeaders(project)
    settingsModel.isUrlNotificationDismissed = ConfigUtil.isUrlNotificationDismissed()
  }

  companion object {
    private val LOG = logger<SettingsMigration>()
  }
}
