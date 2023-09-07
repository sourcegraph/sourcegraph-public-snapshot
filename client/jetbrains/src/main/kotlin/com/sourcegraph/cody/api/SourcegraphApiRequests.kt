package com.sourcegraph.cody.api

import com.sourcegraph.cody.config.CodyAccountDetails
import com.sourcegraph.cody.config.SourcegraphServerPath
import java.awt.Image

object SourcegraphApiRequests {
  object CurrentUser {
    fun getDetails(
        server: SourcegraphServerPath
    ): SourcegraphApiRequest.Post.GQLQuery<CurrentUserWrapper> {

      return SourcegraphApiRequest.Post.GQLQuery.Parsed(
          server.toGraphQLUrl(),
          SourcegraphGQLQueries.getUserDetails,
          null,
          CurrentUserWrapper::class.java)
    }

    data class CurrentUserWrapper(val currentUser: CodyAccountDetails)

    @JvmStatic
    fun getAvatar(url: String) =
        object : SourcegraphApiRequest.Get<Image>(url) {
              override fun extractResult(response: SourcegraphApiResponse): Image {
                return response.handleBody { SourcegraphApiContentHelper.loadImage(it) }
              }
            }
            .withOperationName("get profile avatar")
  }
}
