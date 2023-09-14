package com.sourcegraph.cody.auth

import java.util.EventListener

/** @param A - account type */
interface AccountsListener<A> : EventListener {
  fun onAccountListChanged(old: Collection<A>, new: Collection<A>) {}

  fun onAccountCredentialsChanged(account: A) {}
}
