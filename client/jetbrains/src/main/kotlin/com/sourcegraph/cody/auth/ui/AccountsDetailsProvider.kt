package com.sourcegraph.cody.auth.ui

import com.intellij.util.concurrency.annotations.RequiresEdt
import com.sourcegraph.cody.auth.Account
import com.sourcegraph.cody.auth.AccountDetails
import com.sourcegraph.cody.auth.SingleValueModel
import org.jetbrains.annotations.Nls
import java.awt.Image

interface AccountsDetailsProvider<in A : Account, out D : AccountDetails> {

  @get:RequiresEdt
  val loadingStateModel: SingleValueModel<Boolean>

  @RequiresEdt
  fun getDetails(account: A): D?

  @RequiresEdt
  fun getAvatarImage(account: A): Image?

  @RequiresEdt
  @Nls
  fun getErrorText(account: A): String?

  @RequiresEdt
  fun checkErrorRequiresReLogin(account: A): Boolean

  @RequiresEdt
  fun reset(account: A)

  @RequiresEdt
  fun resetAll()
}
