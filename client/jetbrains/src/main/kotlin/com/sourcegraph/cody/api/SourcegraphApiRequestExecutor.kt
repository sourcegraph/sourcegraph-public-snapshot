package com.sourcegraph.cody.api

import com.intellij.openapi.Disposable
import com.intellij.openapi.components.Service
import com.intellij.openapi.components.service
import com.intellij.openapi.diagnostic.logger
import com.intellij.openapi.progress.EmptyProgressIndicator
import com.intellij.openapi.progress.ProcessCanceledException
import com.intellij.openapi.progress.ProgressIndicator
import com.intellij.util.EventDispatcher
import com.intellij.util.ThrowableConvertor
import com.intellij.util.concurrency.annotations.RequiresBackgroundThread
import com.intellij.util.io.HttpRequests
import com.intellij.util.io.HttpSecurityUtil
import com.intellij.util.io.RequestBuilder
import java.io.IOException
import java.io.InputStream
import java.io.InputStreamReader
import java.io.Reader
import java.net.HttpURLConnection
import java.util.EventListener
import java.util.zip.GZIPInputStream
import org.jetbrains.annotations.CalledInAny
import org.jetbrains.annotations.TestOnly

sealed class SourcegraphApiRequestExecutor {

  protected val authDataChangedEventDispatcher =
      EventDispatcher.create(AuthDataChangeListener::class.java)

  @RequiresBackgroundThread
  @Throws(IOException::class, ProcessCanceledException::class)
  abstract fun <T> execute(indicator: ProgressIndicator, request: SourcegraphApiRequest<T>): T

  @TestOnly
  @RequiresBackgroundThread
  @Throws(IOException::class, ProcessCanceledException::class)
  fun <T> execute(request: SourcegraphApiRequest<T>): T = execute(EmptyProgressIndicator(), request)

  fun addListener(listener: AuthDataChangeListener, disposable: Disposable) =
      authDataChangedEventDispatcher.addListener(listener, disposable)

  fun addListener(disposable: Disposable, listener: () -> Unit) =
      authDataChangedEventDispatcher.addListener(
          object : AuthDataChangeListener {
            override fun authDataChanged() {
              listener()
            }
          },
          disposable)

  class WithTokenAuth internal constructor(token: String, private val useProxy: Boolean) : Base() {
    @Volatile
    internal var token: String = token
      set(value) {
        field = value
        authDataChangedEventDispatcher.multicaster.authDataChanged()
      }

    @Throws(IOException::class, ProcessCanceledException::class)
    override fun <T> execute(indicator: ProgressIndicator, request: SourcegraphApiRequest<T>): T {
      indicator.checkCanceled()
      return createRequestBuilder(request)
          .tuner { connection ->
            request.additionalHeaders.forEach(connection::addRequestProperty)
            connection.addRequestProperty(
                HttpSecurityUtil.AUTHORIZATION_HEADER_NAME, "token $token")
          }
          .useProxy(useProxy)
          .execute(request, indicator)
    }
  }

  abstract class Base() : SourcegraphApiRequestExecutor() {
    protected fun <T> RequestBuilder.execute(
        request: SourcegraphApiRequest<T>,
        indicator: ProgressIndicator
    ): T {
      indicator.checkCanceled()
      try {
        LOG.debug("Request: ${request.url} ${request.operationName} : Connecting")
        return connect {
          val connection = it.connection as HttpURLConnection
          if (request is SourcegraphApiRequest.WithBody) {
            LOG.debug(
                "Request: ${connection.requestMethod} ${connection.url} with body:\n${request.body} : Connected")
            request.body?.let { body -> it.write(body) }
          } else {
            LOG.debug("Request: ${connection.requestMethod} ${connection.url} : Connected")
          }
          checkResponseCode(connection)
          indicator.checkCanceled()
          val result = request.extractResult(createResponse(it, indicator))
          LOG.debug("Request: ${connection.requestMethod} ${connection.url} : Result extracted")
          result
        }
      } catch (e: SourcegraphStatusCodeException) {
        @Suppress("UNCHECKED_CAST")
        if (request is SourcegraphApiRequest.Get.Optional<*> &&
            e.statusCode == HttpURLConnection.HTTP_NOT_FOUND)
            return null as T
        else throw e
      } catch (e: SourcegraphConfusingException) {
        if (request.operationName != null) {
          val errorText = "Can't ${request.operationName}"
          e.setDetails(errorText)
          LOG.debug(errorText, e)
        }
        throw e
      }
    }

    protected fun createRequestBuilder(request: SourcegraphApiRequest<*>): RequestBuilder {
      return when (request) {
            is SourcegraphApiRequest.Get -> HttpRequests.request(request.url)
            is SourcegraphApiRequest.Post -> HttpRequests.post(request.url, request.bodyMimeType)
            else -> throw UnsupportedOperationException("${request.javaClass} is not supported")
          }
          .userAgent("Cody")
          .throwStatusCodeException(false)
          .forceHttps(false)
          .accept(request.acceptMimeType)
    }

    @Throws(IOException::class)
    private fun checkResponseCode(connection: HttpURLConnection) {
      if (connection.responseCode < 400) return
      val statusLine = "${connection.responseCode} ${connection.responseMessage}"
      val errorText = getErrorText(connection)
      LOG.debug(
          "Request: ${connection.requestMethod} ${connection.url} : Error $statusLine body:\n${errorText}")

      throw when (connection.responseCode) {
        HttpURLConnection.HTTP_UNAUTHORIZED,
        HttpURLConnection.HTTP_FORBIDDEN ->
            SourcegraphAuthenticationException("Request response: " + (errorText ?: statusLine))
        else -> SourcegraphStatusCodeException("$statusLine - $errorText", connection.responseCode)
      }
    }

    private fun getErrorText(connection: HttpURLConnection): String? {
      val errorStream = connection.errorStream ?: return null
      val stream =
          if (connection.contentEncoding == "gzip") GZIPInputStream(errorStream) else errorStream
      return InputStreamReader(stream, Charsets.UTF_8).use { it.readText() }
    }

    private fun createResponse(
        request: HttpRequests.Request,
        indicator: ProgressIndicator
    ): SourcegraphApiResponse {
      return object : SourcegraphApiResponse {
        override fun findHeader(headerName: String): String? =
            request.connection.getHeaderField(headerName)

        override fun <T> readBody(converter: ThrowableConvertor<Reader, T, IOException>): T =
            request.getReader(indicator).use { converter.convert(it) }

        override fun <T> handleBody(converter: ThrowableConvertor<InputStream, T, IOException>): T =
            request.inputStream.use { converter.convert(it) }
      }
    }
  }

  @Service
  class Factory {
    @CalledInAny
    fun create(token: String): WithTokenAuth {
      return create(token, true)
    }

    @CalledInAny
    fun create(token: String, useProxy: Boolean = true): WithTokenAuth {
      return WithTokenAuth(token, useProxy)
    }

    companion object {
      @JvmStatic
      val instance: Factory
        get() = service()
    }
  }

  companion object {
    private val LOG = logger<SourcegraphApiRequestExecutor>()
  }

  interface AuthDataChangeListener : EventListener {
    fun authDataChanged()
  }
}
