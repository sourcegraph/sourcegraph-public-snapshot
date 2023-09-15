package com.sourcegraph.cody;

import com.intellij.openapi.project.Project;
import com.sourcegraph.config.AccessTokenStorage;
import io.netty.channel.Channel;
import io.netty.channel.ChannelHandlerContext;
import io.netty.handler.codec.http.FullHttpRequest;
import io.netty.handler.codec.http.HttpResponseStatus;
import io.netty.handler.codec.http.HttpUtil;
import io.netty.handler.codec.http.QueryStringDecoder;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class SaveAccessTokenHttpService extends org.jetbrains.ide.RestService {
  Project project;

  // By saving the project we can make sure that we only handle calls in one IDE window
  public SaveAccessTokenHttpService(Project project) {
    this.project = project;
  }

  @Nullable
  @Override
  public String execute(@NotNull QueryStringDecoder queryStringDecoder,
      @NotNull FullHttpRequest request,
      @NotNull ChannelHandlerContext context) {
    final boolean keepAlive = HttpUtil.isKeepAlive(request);
    final Channel channel = context.channel();

    QueryStringDecoder urlDecoder = new QueryStringDecoder(request.uri());
    if (urlDecoder.path().startsWith("/" + PREFIX + "/" + getServiceName() + "/")) {
      urlDecoder = new QueryStringDecoder(
          urlDecoder.path().substring(1 + PREFIX.length() + 1 + getServiceName().length()));

      String accessToken = urlDecoder.parameters().get("token").get(0);
      String destination = urlDecoder.parameters().get("destination").get(0); // Destination is "/projectLocationHash"
      String projectLocationHash = (destination != null && destination.length() > 1) ? destination.substring(1) : null;
      if (accessToken == null || projectLocationHash == null || !projectLocationHash.equals(
          project.getLocationHash())) {
        sendStatus(HttpResponseStatus.BAD_REQUEST, keepAlive, channel);
        return null;
      }

      AccessTokenStorage.setApplicationDotComAccessToken(accessToken);
      sendStatus(HttpResponseStatus.OK, keepAlive, channel);
    }
    return null;
  }

  @NotNull
  @Override
  protected String getServiceName() {
    return "sourcegraph";
  }
}
