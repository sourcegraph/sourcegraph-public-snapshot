package com.sourcegraph.cody.config

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

    data class CurrentUserWrapper(val currentUser: SourcegraphAccountDetailed)

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
