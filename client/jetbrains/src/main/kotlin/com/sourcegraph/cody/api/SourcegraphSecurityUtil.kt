package com.sourcegraph.cody.api

import com.intellij.openapi.progress.ProgressIndicator
import com.sourcegraph.cody.config.CodyAccountDetails
import com.sourcegraph.cody.config.SourcegraphServerPath

object SourcegraphSecurityUtil {

  @JvmStatic
  fun loadCurrentUserDetails(
      executor: SourcegraphApiRequestExecutor,
      progressIndicator: ProgressIndicator,
      server: SourcegraphServerPath
  ): CodyAccountDetails {
    return executor
        .execute(progressIndicator, SourcegraphApiRequests.CurrentUser.getDetails(server))
        .currentUser
  }
}
