package com.sourcegraph.cody

import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.application.ModalityState
import com.intellij.openapi.progress.EmptyProgressIndicator
import com.intellij.openapi.ui.Messages
import com.sourcegraph.cody.agent.CodyAgent
import com.sourcegraph.cody.api.SourcegraphApiRequestExecutor
import com.sourcegraph.cody.config.CodyPersistentAccountsHost
import com.sourcegraph.cody.config.CodyTokenCredentialsUi
import com.sourcegraph.cody.config.SourcegraphServerPath
import com.sourcegraph.config.ConfigUtil
import io.netty.buffer.Unpooled
import io.netty.channel.ChannelHandlerContext
import io.netty.handler.codec.http.DefaultFullHttpResponse
import io.netty.handler.codec.http.FullHttpRequest
import io.netty.handler.codec.http.FullHttpResponse
import io.netty.handler.codec.http.HttpHeaderNames
import io.netty.handler.codec.http.HttpResponseStatus
import io.netty.handler.codec.http.HttpUtil
import io.netty.handler.codec.http.HttpVersion
import io.netty.handler.codec.http.QueryStringDecoder
import io.netty.util.CharsetUtil
import org.jetbrains.ide.RestService

class SaveAccessTokenHttpService : RestService() {
  override fun execute(
      urlDecoder: QueryStringDecoder,
      request: FullHttpRequest,
      context: ChannelHandlerContext
  ): String? {
    val keepAlive = HttpUtil.isKeepAlive(request)
    val channel = context.channel()
    val requestUriDecoder = QueryStringDecoder(request.uri())
    if (requestUriDecoder.path().startsWith("/" + PREFIX + "/" + getServiceName() + "/")) {
      val accessToken = requestUriDecoder.parameters()["token"]!![0]
      if (accessToken == null) {
        sendStatus(HttpResponseStatus.BAD_REQUEST, keepAlive, channel)
      }

      addAccount(accessToken)

      // Send response
      val htmlContent =
          "<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"utf-8\"><title>Token saved correctly.</title></head><body><p>Token saved correctly.</p><p>You can close this window.</p></body></html>"
      val response: FullHttpResponse =
          DefaultFullHttpResponse(
              HttpVersion.HTTP_1_1,
              HttpResponseStatus.OK,
              Unpooled.copiedBuffer(htmlContent, CharsetUtil.UTF_8))
      response.headers()[HttpHeaderNames.CONTENT_TYPE] = "text/html; charset=UTF-8"
      response.headers()[HttpHeaderNames.CONTENT_LENGTH] = response.content().readableBytes()
      sendResponse(request, context, response)
    }
    return null
  }

  private fun addAccount(accessToken: String) {
    val sourcegraphServerPath = SourcegraphServerPath(ConfigUtil.DOTCOM_URL, "")
    val executor = SourcegraphApiRequestExecutor.Factory.getInstance().create(accessToken)
    val modalityState = ModalityState.NON_MODAL
    val emptyProgressIndicator = EmptyProgressIndicator(modalityState)

    val project = getLastFocusedOrOpenedProject() ?: return
    val accountsHost = CodyPersistentAccountsHost(project)
    runCatching {
          CodyTokenCredentialsUi.acquireLogin(
              sourcegraphServerPath,
              executor,
              emptyProgressIndicator,
              { login, server -> accountsHost.isAccountUnique(login, server) },
              null)
        }
        .fold({ login ->
          // Save account with login and token
          accountsHost.addAccount(sourcegraphServerPath, login, accessToken)
          CodyAgent.getServer(project)
              ?.configurationDidChange(ConfigUtil.getAgentConfiguration(project))
        }) {
          val validationInfo = CodyTokenCredentialsUi.handleError(it)
          ApplicationManager.getApplication().invokeLater {
            Messages.showErrorDialog(project, validationInfo.message, "Failed to sign in")
          }
        }
  }

  override fun getServiceName(): String {
    return "sourcegraph"
  }
}
