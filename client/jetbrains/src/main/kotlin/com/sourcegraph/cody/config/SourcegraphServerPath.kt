package com.sourcegraph.cody.config

import com.intellij.collaboration.api.ServerPath
import com.intellij.util.xmlb.annotations.Attribute
import com.intellij.util.xmlb.annotations.Tag
import com.sourcegraph.config.ConfigUtil
import java.util.regex.Pattern

@Tag("server")
data class SourcegraphServerPath(
    @Attribute("url") var url: String = "",
    @Attribute("customRequestHeaders") var customRequestHeaders: String = ""
) : ServerPath {

  private val GRAPHQL_API_SUFFIX = ".api/graphql"

  override fun toString(): String {
    return url
  }

  fun toGraphQLUrl(): String {
    return url + GRAPHQL_API_SUFFIX
  }

  companion object {
    const val DEFAULT_HOST = ConfigUtil.DOTCOM_URL

    // 1 - schema, 2 - host, 4 - port, 5 - path
    private val URL_REGEX =
        Pattern.compile(
            "^(https?://)?([^/?:]+)(:(\\d+))?((/[^/?#]+)*)?/?", Pattern.CASE_INSENSITIVE)

    @Throws(SourcegraphParseException::class)
    @JvmStatic
    fun from(uri: String, customRequestHeaders: String): SourcegraphServerPath {
      val matcher = URL_REGEX.matcher(uri)
      if (!matcher.matches()) throw SourcegraphParseException("Not a valid URL")
      return SourcegraphServerPath(uri, customRequestHeaders)
    }
  }
}
