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
          "<!DOCTYPE html><html lang=\"en\"> <head> <meta charset=\"utf-8\"> <title>Cody: Authentication successful</title> </head> <body style=\"font-family: system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', Roboto, Oxygen, Ubuntu, Cantarell, \'Open Sans\', \'Helvetica Neue\', sans-serif; background: #f9fafb;\"> <div style=\"margin: 40px auto; text-align: center; max-width: 300px; border: 1px solid #e6ebf2; padding: 40px 20px; border-radius: 8px; background: white; box-shadow: 0px 5px 20px 1px rgba(0, 0, 0, 0.1); \"> <svg width=\"32\" height=\"32\" viewBox=\"0 0 195 176\" fill=\"none\" xmlns=\"http://www.w3.org/2000/svg\"> <path fill-rule=\"evenodd\" clip-rule=\"evenodd\" d=\"M141.819 -8.93872e-07C152.834 -4.002e-07 161.763 9.02087 161.763 20.1487L161.763 55.9685C161.763 67.0964 152.834 76.1172 141.819 76.1172C130.805 76.1172 121.876 67.0963 121.876 55.9685L121.876 20.1487C121.876 9.02087 130.805 -1.38754e-06 141.819 -8.93872e-07Z\" fill=\"#FF5543\"/> <path fill-rule=\"evenodd\" clip-rule=\"evenodd\" d=\"M15.5111 47.0133C15.5111 35.8855 24.44 26.8646 35.4543 26.8646H70.9088C81.9231 26.8646 90.8519 35.8855 90.8519 47.0133C90.8519 58.1411 81.9231 67.162 70.9088 67.162H35.4543C24.44 67.162 15.5111 58.1411 15.5111 47.0133Z\" fill=\"#A112FF\"/> <path fill-rule=\"evenodd\" clip-rule=\"evenodd\" d=\"M189.482 105.669C196.58 112.482 196.868 123.818 190.125 130.989L183.85 137.662C134.75 189.88 51.971 188.579 4.50166 134.844C-2.01751 127.464 -1.38097 116.142 5.92343 109.556C13.2278 102.97 24.434 103.613 30.9532 110.993C64.6181 149.101 123.324 150.024 158.146 112.991L164.42 106.318C171.164 99.1472 182.384 98.8565 189.482 105.669Z\" fill=\"#00CBEC\"/> </svg> <h4>Authentication successful</h4> <p style=\"font-size: 12px;\">You may close this tab and return to your editor</p> </body></html>"
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
          CodyTokenCredentialsUi.acquireDetails(
              sourcegraphServerPath,
              executor,
              emptyProgressIndicator,
              { login, server -> accountsHost.isAccountUnique(login, server) },
              null)
        }
        .fold({ accountDetails ->
          // Save account with login and token
          accountsHost.addAccount(
              sourcegraphServerPath,
              accountDetails.username,
              accountDetails.displayName,
              accessToken)
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
