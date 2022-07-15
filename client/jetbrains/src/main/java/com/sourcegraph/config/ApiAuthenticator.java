package com.sourcegraph.config;

import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.intellij.openapi.diagnostic.Logger;
import com.sourcegraph.api.GraphQlClient;
import org.apache.http.HttpResponse;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.io.IOException;
import java.util.function.Consumer;

public class ApiAuthenticator {
    private final static Logger logger = Logger.getInstance(ApiAuthenticator.class);

    public static void testConnection(@NotNull String instanceUrl, @Nullable String accessToken, @NotNull Consumer<ConnectionStatus> callback) {
        new Thread(() -> {
            String query = "" +
                "query {" +
                "    currentUser {" +
                "        id" +
                "    }" +
                "}";

            try {
                HttpResponse response = GraphQlClient.callGraphQLService(instanceUrl, accessToken, query, new JsonObject());
                if (GraphQlClient.getStatusCode(response) == 200) {
                    JsonElement id = GraphQlClient.getResponseBodyJson(response).getAsJsonObject().get("currentUser").getAsJsonObject().get("id");
                    callback.accept(id.isJsonNull() ? ConnectionStatus.COULD_CONNECT_BUT_NOT_AUTHENTICATED : ConnectionStatus.AUTHENTICATED);
                } else {
                    callback.accept(ConnectionStatus.COULD_NOT_CONNECT);
                }
            } catch (IOException e) {
                logger.info(e);
            }
        }).start();

    }

    enum ConnectionStatus {
        AUTHENTICATED,
        COULD_NOT_CONNECT,
        COULD_CONNECT_BUT_NOT_AUTHENTICATED
    }
}
