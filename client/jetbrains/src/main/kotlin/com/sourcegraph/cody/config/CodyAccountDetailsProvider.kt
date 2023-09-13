package com.sourcegraph.cody.config

import com.intellij.collaboration.async.CompletableFutureUtil.submitIOTask
import com.intellij.collaboration.async.CompletableFutureUtil.successOnEdt
import com.intellij.collaboration.util.ProgressIndicatorsProvider
import com.intellij.openapi.application.ModalityState
import com.intellij.openapi.components.service
import com.intellij.openapi.progress.EmptyProgressIndicator
import com.intellij.openapi.progress.ProgressIndicator
import com.intellij.openapi.progress.ProgressManager
import com.intellij.util.IconUtil
import com.sourcegraph.cody.api.SourcegraphApiRequestExecutor
import com.sourcegraph.cody.api.SourcegraphSecurityUtil
import com.sourcegraph.cody.auth.ui.LoadingAccountsDetailsProvider
import java.util.concurrent.CompletableFuture

class CodyAccountDetailsProvider(
    progressIndicatorsProvider: ProgressIndicatorsProvider,
    private val accountManager: CodyAccountManager,
    private val accountsModel: CodyAccountListModel
) : LoadingAccountsDetailsProvider<CodyAccount, CodyAccountDetails>(progressIndicatorsProvider) {

  override fun scheduleLoad(
      account: CodyAccount,
      indicator: ProgressIndicator
  ): CompletableFuture<DetailsLoadingResult<CodyAccountDetails>> {
    val token =
        accountsModel.newCredentials.getOrElse(account) { accountManager.findCredentials(account) }
            ?: return CompletableFuture.completedFuture(noToken())
    val executor = service<SourcegraphApiRequestExecutor.Factory>().create(token)
    return ProgressManager.getInstance()
        .submitIOTask(EmptyProgressIndicator()) {
          if (account.isCodyApp()) {
            val details = CodyAccountDetails(account.name, account.name, null)
            DetailsLoadingResult(details, IconUtil.toBufferedImage(defaultIcon), null, false)
          } else {
            val accountDetails =
                SourcegraphSecurityUtil.loadCurrentUserDetails(executor, it, account.server)
            val image =
                accountDetails.avatarURL?.let { url ->
                  CachingCodyUserAvatarLoader.getInstance().requestAvatar(executor, url).join()
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
      DetailsLoadingResult<CodyAccountDetails>(null, null, "Missing access token", true)
}
