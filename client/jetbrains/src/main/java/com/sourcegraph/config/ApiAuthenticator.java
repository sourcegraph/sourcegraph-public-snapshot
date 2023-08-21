package com.sourcegraph.config;

import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.agent.CodyAgentServer;
import java.util.concurrent.CompletableFuture;
import org.jetbrains.annotations.NotNull;

public class ApiAuthenticator {
  public static CompletableFuture<ConnectionStatus> testConnection(@NotNull Project project) {
    return CodyAgent.withServer(project, CodyAgentServer::currentUserId)
        .thenApply((ignored) -> ConnectionStatus.AUTHENTICATED)
        .exceptionally(ignored -> ConnectionStatus.NOT_AUTHENTICATED);
  }

  public enum ConnectionStatus {
    AUTHENTICATED,
    NOT_AUTHENTICATED
  }
}
