package com.sourcegraph.cody

import com.intellij.openapi.project.ProjectManager
import com.sourcegraph.cody.agent.CodyAgent
import com.sourcegraph.cody.config.CodyPersistentAccountsHost
import com.sourcegraph.cody.config.SourcegraphServerPath
import com.sourcegraph.config.ConfigUtil
import io.netty.buffer.Unpooled
import io.netty.channel.ChannelHandlerContext
import io.netty.handler.codec.http.*
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

      // Save token
      val project = ProjectManager.getInstance().openProjects[0]
      val accountsHost = CodyPersistentAccountsHost(project)
      accountsHost.addAccount(SourcegraphServerPath(ConfigUtil.DOTCOM_URL, ""), "", accessToken)
      CodyAgent.getServer(project)
          ?.configurationDidChange(ConfigUtil.getAgentConfiguration(project))

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

  override fun getServiceName(): String {
    return "sourcegraph"
  }
}
