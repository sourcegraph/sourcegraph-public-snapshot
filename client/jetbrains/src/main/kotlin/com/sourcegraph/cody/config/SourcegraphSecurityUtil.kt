package com.sourcegraph.cody.config

import com.intellij.openapi.progress.ProgressIndicator

object SourcegraphSecurityUtil {

  @JvmStatic
  fun loadCurrentUserDetails(
      executor: SourcegraphApiRequestExecutor,
      progressIndicator: ProgressIndicator,
      server: SourcegraphServerPath
  ): SourcegraphAccountDetailed {
    return executor
        .execute(progressIndicator, SourcegraphApiRequests.CurrentUser.getDetails(server))
        .currentUser
  }
}
