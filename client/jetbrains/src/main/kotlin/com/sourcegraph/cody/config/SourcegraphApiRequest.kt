package com.sourcegraph.cody.config

import com.fasterxml.jackson.databind.JsonNode
import com.intellij.collaboration.api.dto.GraphQLErrorDTO
import com.intellij.collaboration.api.dto.GraphQLRequestDTO
import com.intellij.collaboration.api.dto.GraphQLResponseDTO
import org.jetbrains.plugins.github.api.GithubApiContentHelper
import org.jetbrains.plugins.github.exceptions.GithubJsonException
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
    constructor(url: String, acceptMimeType: String? = null) : Get<T?>(url, acceptMimeType) {
      companion object {
        inline fun <reified T> json(url: String, acceptMimeType: String? = null): Optional<T> =
            Json(url, T::class.java, acceptMimeType)
      }

      open class Json<T>(
          url: String,
          private val clazz: Class<T>,
          acceptMimeType: String? = "application/json"
      ) : Optional<T>(url, acceptMimeType) {

        override fun extractResult(response: SourcegraphApiResponse): T =
            parseJsonObject(response, clazz)
      }
    }

    companion object {
      inline fun <reified T> json(url: String, acceptMimeType: String? = null): Get<T> =
          Json(url, T::class.java, acceptMimeType)
    }

    open class Json<T>(
        url: String,
        private val clazz: Class<T>,
        acceptMimeType: String? = SourcegraphApiContentHelper.JSON_MIME_TYPE
    ) : Get<T>(url, acceptMimeType) {

      override fun extractResult(response: SourcegraphApiResponse): T =
          parseJsonObject(response, clazz)
    }
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
    companion object {
      inline fun <reified T> json(url: String, body: Any, acceptMimeType: String? = null): Post<T> =
          Json(url, body, T::class.java, acceptMimeType)
    }

    open class Json<T>(
        url: String,
        private val bodyObject: Any,
        private val clazz: Class<T>,
        acceptMimeType: String? = SourcegraphApiContentHelper.JSON_MIME_TYPE
    ) : Post<T>(SourcegraphApiContentHelper.JSON_MIME_TYPE, url, acceptMimeType) {

      override val body: String
        get() = SourcegraphApiContentHelper.toJson(bodyObject)

      override fun extractResult(response: SourcegraphApiResponse): T =
          parseJsonObject(response, clazz)
    }

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

      class TraversedParsed<out T : Any>(
          url: String,
          requestFilePath: String,
          variablesObject: Any,
          private val clazz: Class<out T>,
          private vararg val pathFromData: String
      ) : GQLQuery<T>(url, requestFilePath, variablesObject) {

        override fun extractResult(response: SourcegraphApiResponse): T {
          return parseResponse(response, clazz, pathFromData)
              ?: throw GithubJsonException("Non-nullable entity is null or entity path is invalid")
        }
      }

      class OptionalTraversedParsed<T>(
          url: String,
          requestFilePath: String,
          variablesObject: Any,
          private val clazz: Class<T>,
          private vararg val pathFromData: String
      ) : GQLQuery<T?>(url, requestFilePath, variablesObject) {
        override fun extractResult(response: SourcegraphApiResponse): T? {
          return parseResponse(response, clazz, pathFromData)
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

  abstract class Put<T>
  @JvmOverloads
  constructor(
      override val bodyMimeType: String,
      url: String,
      override val acceptMimeType: String? = null
  ) : WithBody<T>(url) {
    companion object {
      inline fun <reified T> json(url: String, body: Any? = null): Put<T> =
          Json(url, body, T::class.java)
    }

    open class Json<T>(url: String, private val bodyObject: Any?, private val clazz: Class<T>) :
        Put<T>(
            SourcegraphApiContentHelper.JSON_MIME_TYPE,
            url,
            SourcegraphApiContentHelper.JSON_MIME_TYPE) {
      init {
        if (bodyObject == null) headers["Content-Length"] = "0"
      }

      override val body: String?
        get() = bodyObject?.let { SourcegraphApiContentHelper.toJson(it) }

      override fun extractResult(response: SourcegraphApiResponse): T =
          parseJsonObject(response, clazz)
    }
  }

  abstract class Patch<T>
  @JvmOverloads
  constructor(
      override val bodyMimeType: String,
      url: String,
      override var acceptMimeType: String? = null
  ) : Post<T>(bodyMimeType, url, acceptMimeType) {
    companion object {
      inline fun <reified T> json(url: String, body: Any): Post<T> = Json(url, body, T::class.java)
    }

    open class Json<T>(url: String, bodyObject: Any, clazz: Class<T>) :
        Post.Json<T>(url, bodyObject, clazz)
  }

  companion object {
    private fun <T> parseJsonObject(response: SourcegraphApiResponse, clazz: Class<T>): T {
      return response.readBody { SourcegraphApiContentHelper.readJsonObject(it, clazz) }
    }
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
