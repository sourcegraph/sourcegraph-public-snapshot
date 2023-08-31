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
import com.sourcegraph.config.AccessTokenStorage
import com.sourcegraph.config.CodyApplicationService
import com.sourcegraph.config.CodyProjectService
import com.sourcegraph.config.ConfigUtil
import com.sourcegraph.config.UserLevelConfig
import java.util.UUID
import java.util.concurrent.CompletableFuture

class SettingsMigration : StartupActivity, DumbAware {

  private val accountManager = service<SourcegraphAccountManager>()
  override fun runActivity(project: Project) {
    RunOnceUtil.runOnceForProject(project, UUID.randomUUID().toString()) {
      migrateAccounts(project)
      migrateProjectSettings(project)
    }
    RunOnceUtil.runOnceForApp(UUID.randomUUID().toString()) { migrateApplicationSettings() }
  }

  private fun migrateAccounts(
      project: Project,
  ) {
    val defaultAccountHolder = project.service<SourcegraphProjectDefaultAccountHolder>()
    val requestExecutorFactory = SourcegraphApiRequestExecutor.Factory.getInstance()
    val defaultAccountType = ConfigUtil.getDefaultAccountType(project)
    migrateDotcomAccount(project, defaultAccountType, requestExecutorFactory, defaultAccountHolder)
    migrateEnterpriseAccount(
        project, defaultAccountType, requestExecutorFactory, defaultAccountHolder)
    migrateCodyAccount(defaultAccountType, requestExecutorFactory, defaultAccountHolder)
  }

  private fun migrateDotcomAccount(
      project: Project,
      accountType: AccountType,
      requestExecutorFactory: SourcegraphApiRequestExecutor.Factory,
      defaultAccountHolder: SourcegraphProjectDefaultAccountHolder
  ) {
    val dotcomAccessToken = extractDotcomAccessToken(project)
    if (!dotcomAccessToken.isNullOrEmpty()) {
      val server = SourcegraphServerPath(ConfigUtil.DOTCOM_URL)
      val shouldSetAccountAsDefault = accountType == AccountType.DOTCOM
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
      accountType: AccountType,
      requestExecutorFactory: SourcegraphApiRequestExecutor.Factory,
      defaultAccountHolder: SourcegraphProjectDefaultAccountHolder
  ) {
    val enterpriseAccessToken = extractEnterpriseAccessToken(project)
    if (!enterpriseAccessToken.isNullOrEmpty()) {
      val enterpriseUrl = extractEnterpriseUrl(project)
      runCatching { SourcegraphServerPath.from(enterpriseUrl) }
          .fold({
            val shouldSetAccountAsDefault = accountType == AccountType.ENTERPRISE
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
      accountType: AccountType,
      requestExecutorFactory: SourcegraphApiRequestExecutor.Factory,
      defaultAccountHolder: SourcegraphProjectDefaultAccountHolder
  ) {
    val localAppAccessToken = LocalAppManager.getLocalAppAccessToken().orNull()
    if (LocalAppManager.isLocalAppInstalled() && localAppAccessToken != null) {
      val codyUrl = LocalAppManager.getLocalAppUrl()
      runCatching { SourcegraphServerPath.from(codyUrl) }
          .fold({
            val shouldSetAccountAsDefault = accountType == AccountType.LOCAL_APP
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

  private fun extractEnterpriseUrl(project: Project): String {
    // Project level
    val projectLevelUrl = project.service<CodyProjectService>().sourcegraphUrl
    if (!projectLevelUrl.isNullOrEmpty()) {
      return addSlashIfNeeded(projectLevelUrl)
    }

    // Application level
    val applicationLevelUrl = service<CodyApplicationService>().sourcegraphUrl
    if (!applicationLevelUrl.isNullOrEmpty()) {
      return addSlashIfNeeded(applicationLevelUrl)
    }

    // User level or default
    val userLevelUrl = UserLevelConfig.getSourcegraphUrl()
    return if (userLevelUrl.isNotEmpty()) addSlashIfNeeded(userLevelUrl) else ""
  }

  private fun extractCustomRequestHeaders(project: Project): String {
    // Project level
    val projectLevelCustomRequestHeaders =
        project.service<CodyProjectService>().getCustomRequestHeaders()
    if (!projectLevelCustomRequestHeaders.isNullOrEmpty()) {
      return projectLevelCustomRequestHeaders
    }

    // Application level
    val applicationLevelCustomRequestHeaders =
        service<CodyApplicationService>().getCustomRequestHeaders()
    return if (!applicationLevelCustomRequestHeaders.isNullOrEmpty()) {
      applicationLevelCustomRequestHeaders
    } else ""
  }

  private fun extractDefaultBranchName(project: Project): String {
    // Project level
    val projectLevelDefaultBranchName = project.service<CodyProjectService>().defaultBranchName
    if (!projectLevelDefaultBranchName.isNullOrEmpty()) {
      return projectLevelDefaultBranchName
    }

    // Application level
    val applicationLevelDefaultBranchName = service<CodyApplicationService>().defaultBranchName
    if (!applicationLevelDefaultBranchName.isNullOrEmpty()) {
      return applicationLevelDefaultBranchName
    }

    // User level or default
    val userLevelDefaultBranchName = UserLevelConfig.getDefaultBranchName()
    return userLevelDefaultBranchName ?: "main"
  }

  private fun extractRemoteUrlReplacements(project: Project): String {
    // Project level
    val projectLevelReplacements = project.service<CodyProjectService>().getRemoteUrlReplacements()
    if (!projectLevelReplacements.isNullOrEmpty()) {
      return projectLevelReplacements
    }

    // Application level
    val applicationLevelReplacements = service<CodyApplicationService>().getRemoteUrlReplacements()
    if (!applicationLevelReplacements.isNullOrEmpty()) {
      return applicationLevelReplacements
    }

    // User level or default
    val userLevelRemoteUrlReplacements = UserLevelConfig.getRemoteUrlReplacements()
    return userLevelRemoteUrlReplacements ?: ""
  }

  private fun extractDotcomAccessToken(project: Project): String? {
    // Project level overrides secure storage
    val projectLevelAccessToken = project.service<CodyProjectService>().getDotComAccessToken()
    if (projectLevelAccessToken != null) {
      return projectLevelAccessToken
    }

    // Get token from secure storage
    val securelyStoredAccessToken = AccessTokenStorage.getDotComAccessToken()
    if (securelyStoredAccessToken.isEmpty) {
      return null // Uer denied access to token storage
    }
    if (securelyStoredAccessToken.get().isNotEmpty()) {
      return securelyStoredAccessToken.get()
    }

    // No secure token found, so use app-level token.
    val codyApplicationService = service<CodyApplicationService>()
    val unsafeApplicationLevelAccessToken = codyApplicationService.dotComAccessToken
    return unsafeApplicationLevelAccessToken ?: ""
  }

  private fun extractEnterpriseAccessToken(project: Project): String? {
    // Project level overrides secure storage
    val unsafeProjectLevelAccessToken =
        project.service<CodyProjectService>().getEnterpriseAccessToken()
    if (unsafeProjectLevelAccessToken != null) {
      return unsafeProjectLevelAccessToken
    }

    // Get token from secure storage
    val securelyStoredAccessToken = AccessTokenStorage.getEnterpriseAccessToken()
    if (securelyStoredAccessToken.isEmpty) {
      return null // Uer denied access to token storage
    }
    if (securelyStoredAccessToken.get().isNotEmpty()) {
      return securelyStoredAccessToken.get()
    }

    // No secure token found, so use app-level token.
    val service = service<CodyApplicationService>()
    val unsafeApplicationLevelAccessToken = service.enterpriseAccessToken
    return unsafeApplicationLevelAccessToken ?: ""
  }

  private fun migrateProjectSettings(project: Project) {
    val codyProjectSettings = project.service<CodyProjectSettings>()
    codyProjectSettings.defaultBranchName = extractDefaultBranchName(project)
    codyProjectSettings.remoteUrlReplacements = extractRemoteUrlReplacements(project)
    codyProjectSettings.customRequestHeaders = extractCustomRequestHeaders(project)
  }

  private fun addSlashIfNeeded(url: String): String {
    return if (url.endsWith("/")) url else "$url/"
  }

  private fun migrateApplicationSettings() {
    val codyApplicationSettings = service<CodyApplicationSettings>()
    val codyApplicationService = service<CodyApplicationService>()
    codyApplicationSettings.isCodyEnabled = codyApplicationService.isCodyEnabled
    codyApplicationSettings.isCodyAutocompleteEnabled =
        if (codyApplicationService.isCodyAutocompleteEnabled == true) true
        else codyApplicationService.areCodyCompletionsEnabled ?: false
    codyApplicationSettings.isCodyDebugEnabled =
        codyApplicationService.isCodyDebugEnabled ?: false
    codyApplicationSettings.isCodyVerboseDebugEnabled =
        codyApplicationService.isCodyVerboseDebugEnabled ?: false
    codyApplicationSettings.isUrlNotificationDismissed =
        codyApplicationService.isUrlNotificationDismissed
    codyApplicationSettings.anonymousUserId = codyApplicationService.anonymousUserId
    codyApplicationSettings.isInstallEventLogged =
        codyApplicationService.isInstallEventLogged
    codyApplicationSettings.lastUpdateNotificationPluginVersion =
        codyApplicationService.lastUpdateNotificationPluginVersion
  }

  companion object {
    private val LOG = logger<SettingsMigration>()
  }
}
