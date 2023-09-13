package com.sourcegraph.cody.config

import com.intellij.util.xmlb.annotations.Attribute
import com.intellij.util.xmlb.annotations.Tag
import com.sourcegraph.cody.auth.ServerPath
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
      val extractedSchema = matcher.group(1)
      val schema = if (extractedSchema.isNullOrEmpty()) "https://" else extractedSchema
      val host = matcher.group(2) ?: throw SourcegraphParseException("Empty host")
      val port: Int?
      val extractedPort = matcher.group(4)
      port =
          if (extractedPort == null) {
            null
          } else {
            try {
              extractedPort.toInt()
            } catch (e: NumberFormatException) {
              throw SourcegraphParseException("Invalid port format")
            }
          }

      val extractedPath = matcher.group(5)
      val path = if (!extractedPath.endsWith("/")) "$extractedPath/" else extractedPath
      val fullUri = schema + host + (port?.let { ":$it" } ?: "") + path
      return SourcegraphServerPath(fullUri, customRequestHeaders)
    }
  }
}
