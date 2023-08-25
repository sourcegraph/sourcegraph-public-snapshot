package com.sourcegraph.cody.config

import com.intellij.collaboration.async.CompletableFutureUtil.submitIOTask
import com.intellij.collaboration.async.CompletableFutureUtil.successOnEdt
import com.intellij.collaboration.auth.ui.LoadingAccountsDetailsProvider
import com.intellij.collaboration.util.ProgressIndicatorsProvider
import com.intellij.openapi.application.ModalityState
import com.intellij.openapi.components.service
import com.intellij.openapi.progress.EmptyProgressIndicator
import com.intellij.openapi.progress.ProgressIndicator
import com.intellij.openapi.progress.ProgressManager
import com.intellij.util.IconUtil
import com.sourcegraph.cody.Icons
import java.util.concurrent.CompletableFuture

internal class SourcegraphAccounDetailsProvider(
    progressIndicatorsProvider: ProgressIndicatorsProvider,
    private val accountManager: SourcegraphAccountManager,
    private val accountsModel: SourcegraphAccountListModel
) :
    LoadingAccountsDetailsProvider<SourcegraphAccount, SourcegraphAccountDetailed>(
        progressIndicatorsProvider) {

  override val defaultIcon = IconUtil.resizeSquared(Icons.CodyLogo, 40)

  override fun scheduleLoad(
      account: SourcegraphAccount,
      indicator: ProgressIndicator
  ): CompletableFuture<DetailsLoadingResult<SourcegraphAccountDetailed>> {
    val token =
        accountsModel.newCredentials.getOrElse(account) { accountManager.findCredentials(account) }
            ?: return CompletableFuture.completedFuture(noToken())
    val executor = service<SourcegraphApiRequestExecutor.Factory>().create(token)
    return ProgressManager.getInstance()
        .submitIOTask(EmptyProgressIndicator()) {
          if (account.isCodyApp()) {
            val details = SourcegraphAccountDetailed(account.name, null)
            DetailsLoadingResult(details, IconUtil.toBufferedImage(defaultIcon), null, false)
          } else {
            val accountDetails =
                SourcegraphSecurityUtil.loadCurrentUserDetails(executor, it, account.server)
            val image =
                accountDetails.avatarURL?.let { url ->
                  CachingSourcegraphUserAvatarLoader.getInstance()
                      .requestAvatar(executor, url)
                      .join()
                }
            DetailsLoadingResult(accountDetails, image, null, false)
          }
        }
        .successOnEdt(ModalityState.any()) {
          accountsModel.accountsListModel.contentsChanged(account)
          it
        }
  }

  private fun noToken() =
      DetailsLoadingResult<SourcegraphAccountDetailed>(null, null, "Missing access token", true)
}
