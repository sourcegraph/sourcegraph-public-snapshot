package com.sourcegraph.cody.api

import com.fasterxml.jackson.databind.JsonNode
import com.intellij.collaboration.api.dto.GraphQLErrorDTO
import com.intellij.collaboration.api.dto.GraphQLRequestDTO
import com.intellij.collaboration.api.dto.GraphQLResponseDTO
import java.io.IOException

sealed class SourcegraphApiRequest<out T>(val url: String) {
  var operationName: String? = null
  abstract val acceptMimeType: String?

  protected val headers = mutableMapOf<String, String>()
  val additionalHeaders: Map<String, String>
    get() = headers

  @Throws(IOException::class) abstract fun extractResult(response: SourcegraphApiResponse): T

  fun withOperationName(name: String): SourcegraphApiRequest<T> {
    operationName = name
    return this
  }

  abstract class Get<T>
  @JvmOverloads
  constructor(url: String, override val acceptMimeType: String? = null) :
      SourcegraphApiRequest<T>(url) {
    abstract class Optional<T>
    @JvmOverloads
    constructor(url: String, acceptMimeType: String? = null) : Get<T?>(url, acceptMimeType)
  }

  abstract class WithBody<out T>(url: String) : SourcegraphApiRequest<T>(url) {
    abstract val body: String?
    abstract val bodyMimeType: String
  }

  abstract class Post<out T>
  @JvmOverloads
  constructor(
      override val bodyMimeType: String,
      url: String,
      override var acceptMimeType: String? = null
  ) : WithBody<T>(url) {

    abstract class GQLQuery<out T>(
        url: String,
        private val queryName: String,
        private val variablesObject: Any?
    ) : Post<T>(SourcegraphApiContentHelper.JSON_MIME_TYPE, url) {

      override val body: String
        get() {
          val query = SourcegraphGQLQueryLoader.loadQuery(queryName)
          val request = GraphQLRequestDTO(query, variablesObject)
          return SourcegraphApiContentHelper.toJson(request, true)
        }

      protected fun throwException(errors: List<GraphQLErrorDTO>): Nothing {
        if (errors.size == 1) throw SourcegraphConfusingException(errors.single().toString())
        throw SourcegraphConfusingException(errors.toString())
      }

      class Parsed<out T>(
          url: String,
          requestFilePath: String,
          variablesObject: Any?,
          private val clazz: Class<T>
      ) : GQLQuery<T>(url, requestFilePath, variablesObject) {
        override fun extractResult(response: SourcegraphApiResponse): T {
          val result: GraphQLResponseDTO<out T, GraphQLErrorDTO> = parseGQLResponse(response, clazz)
          val data = result.data
          if (data != null) return data

          val errors = result.errors
          if (errors == null) error("Undefined request state - both result and errors are null")
          else throwException(errors)
        }
      }

      internal fun <T> parseResponse(
          response: SourcegraphApiResponse,
          clazz: Class<T>,
          pathFromData: Array<out String>
      ): T? {
        val result: GraphQLResponseDTO<out JsonNode, GraphQLErrorDTO> =
            parseGQLResponse(response, JsonNode::class.java)
        val data = result.data
        if (data != null && !data.isNull) {
          var node: JsonNode = data
          for (path in pathFromData) {
            node = node[path] ?: break
          }
          if (!node.isNull)
              return SourcegraphApiContentHelper.fromJson(node.toString(), clazz, true)
        }
        val errors = result.errors
        if (errors == null) return null else throwException(errors)
      }
    }
  }
  companion object {
    private fun <T> parseGQLResponse(
        response: SourcegraphApiResponse,
        dataClass: Class<out T>
    ): GraphQLResponseDTO<out T, GraphQLErrorDTO> {
      return response.readBody {
        @Suppress("UNCHECKED_CAST")
        SourcegraphApiContentHelper.readJsonObject(
            it,
            GraphQLResponseDTO::class.java,
            dataClass,
            GraphQLErrorDTO::class.java,
            gqlNaming = true) as GraphQLResponseDTO<T, GraphQLErrorDTO>
      }
    }
  }
}
