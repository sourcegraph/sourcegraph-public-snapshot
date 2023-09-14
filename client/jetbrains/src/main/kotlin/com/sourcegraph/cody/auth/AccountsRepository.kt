package com.sourcegraph.cody.auth

/**
 * In most cases should be an instance of [com.intellij.openapi.components.PersistentStateComponent]
 */
interface AccountsRepository<A : Account> {
  var accounts: Set<A>
}
