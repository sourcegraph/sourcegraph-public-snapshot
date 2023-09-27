package com.sourcegraph.cody.auth

import java.util.concurrent.CompletableFuture

interface AuthService {
  val name: String

  /** Starting the authorization flow */
  fun authorize(request: AuthRequest): CompletableFuture<String>

  /** Processing received access token */
  fun handleServerCallback(path: String, accessToken: String): Boolean
}
