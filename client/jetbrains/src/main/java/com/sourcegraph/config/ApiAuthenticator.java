package com.sourcegraph.config;

import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonSyntaxException;
import com.intellij.openapi.diagnostic.Logger;
import com.sourcegraph.api.GraphQlClient;
import com.sourcegraph.api.GraphQlResponse;
import java.io.IOException;
import java.util.function.Consumer;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class ApiAuthenticator {
  private static final Logger logger = Logger.getInstance(ApiAuthenticator.class);

  public static void testConnection(
      @NotNull String instanceUrl,
      @Nullable String accessToken,
      @Nullable String customRequestHeaders,
      @NotNull Consumer<ConnectionStatus> callback) {
    new Thread(
            () -> {
              String query = "" + "query {" + "    currentUser {" + "        id" + "    }" + "}";

              try {
                GraphQlResponse response =
                    GraphQlClient.callGraphQLService(
                        instanceUrl, accessToken, customRequestHeaders, query, new JsonObject());
                if (response.getStatusCode() == 200) {
                  JsonElement id =
                      response
                          .getBodyAsJson()
                          .getAsJsonObject("data")
                          .getAsJsonObject("currentUser")
                          .get("id");
                  callback.accept(
                      id.isJsonNull()
                          ? ConnectionStatus.COULD_CONNECT_BUT_NOT_AUTHENTICATED
                          : ConnectionStatus.AUTHENTICATED);
                } else {
                  callback.accept(ConnectionStatus.COULD_NOT_CONNECT);
                }
              } catch (IOException | JsonSyntaxException e) {
                callback.accept(ConnectionStatus.COULD_NOT_CONNECT);
                logger.info(e);
              }
            })
        .start();
  }

  public enum ConnectionStatus {
    AUTHENTICATED,
    COULD_NOT_CONNECT,
    COULD_CONNECT_BUT_NOT_AUTHENTICATED
  }
}
