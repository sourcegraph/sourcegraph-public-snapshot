package com.sourcegraph.cody;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.project.ProjectManager;
import com.sourcegraph.cody.config.CodyPersistentAccountsHost;
import com.sourcegraph.cody.config.SourcegraphServerPath;
import com.sourcegraph.config.ConfigUtil;
import io.netty.buffer.Unpooled;
import io.netty.channel.Channel;
import io.netty.channel.ChannelHandlerContext;
import io.netty.handler.codec.http.DefaultFullHttpResponse;
import io.netty.handler.codec.http.FullHttpRequest;
import io.netty.handler.codec.http.FullHttpResponse;
import io.netty.handler.codec.http.HttpHeaderNames;
import io.netty.handler.codec.http.HttpResponseStatus;
import io.netty.handler.codec.http.HttpUtil;
import io.netty.handler.codec.http.HttpVersion;
import io.netty.handler.codec.http.QueryStringDecoder;
import io.netty.util.CharsetUtil;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class SaveAccessTokenHttpService extends org.jetbrains.ide.RestService {
  @Nullable
  @Override
  public String execute(
      @NotNull QueryStringDecoder queryStringDecoder,
      @NotNull FullHttpRequest request,
      @NotNull ChannelHandlerContext context) {
    final boolean keepAlive = HttpUtil.isKeepAlive(request);
    final Channel channel = context.channel();

    QueryStringDecoder urlDecoder = new QueryStringDecoder(request.uri());
    if (urlDecoder.path().startsWith("/" + PREFIX + "/" + getServiceName() + "/")) {
      String accessToken = urlDecoder.parameters().get("token").get(0);
      if (accessToken == null) {
        sendStatus(HttpResponseStatus.BAD_REQUEST, keepAlive, channel);
        return null;
      }

      // Save token
      Project project = ProjectManager.getInstance().getOpenProjects()[0];
      var accountsHost = new CodyPersistentAccountsHost(project);
      accountsHost.addAccount(
          new SourcegraphServerPath(ConfigUtil.DOTCOM_URL, ""), "", accessToken);

      // Send response
      String htmlContent =
          "<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"utf-8\"><title>Token saved correctly.</title></head><body><p>Token saved correctly.</p><p>You can close this window.</p></body></html>";
      FullHttpResponse response =
          new DefaultFullHttpResponse(
              HttpVersion.HTTP_1_1,
              HttpResponseStatus.OK,
              Unpooled.copiedBuffer(htmlContent, CharsetUtil.UTF_8));
      response.headers().set(HttpHeaderNames.CONTENT_TYPE, "text/html; charset=UTF-8");
      response.headers().set(HttpHeaderNames.CONTENT_LENGTH, response.content().readableBytes());
      sendResponse(request, context, response);
    }

    return null;
  }

  @NotNull
  @Override
  protected String getServiceName() {
    return "sourcegraph";
  }
}
