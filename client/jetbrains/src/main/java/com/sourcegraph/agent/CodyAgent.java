package com.sourcegraph.agent;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.diagnostic.Logger;
import com.sourcegraph.agent.protocol.*;
import java.io.IOException;
import java.io.PrintWriter;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardOpenOption;
import java.util.Objects;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import org.eclipse.lsp4j.jsonrpc.Launcher;
import org.jetbrains.annotations.NotNull;

/**
 * Orchestrator for the Cody agent, which is a Node.js program that implements the prompt logic for
 * Cody. The agent communicates via a JSON-RPC protocol that is documented in the file
 * "client/cody-agent/src/protocol.ts".
 */
public class CodyAgent {
  // TODO: actually stop the agent based on application lifecycle events
  public static Logger logger = Logger.getInstance(CodyAgent.class);

  private CodyAgentClient client;
  public static final ExecutorService executorService = Executors.newCachedThreadPool();

  public static boolean isConnected() {
    CodyAgent agent = ApplicationManager.getApplication().getService(CodyAgent.class);
    return agent != null && agent.client != null && agent.client.server != null;
  }

  @NotNull
  public static CodyAgentClient getClient() {
    return ApplicationManager.getApplication().getService(CodyAgent.class).client;
  }

  @NotNull
  public static CodyAgentServer getServer() {
    return Objects.requireNonNull(Objects.requireNonNull(getClient()).server);
  }

  public static synchronized void run() {
    if (CodyAgent.isConnected()) {
      return;
    }
    CodyAgentClient client = new CodyAgentClient();
    try {
      CodyAgent.run(client);
    } catch (Exception e) {
      logger.error("unable to start Cody agent", e);
    }
    ApplicationManager.getApplication().getService(CodyAgent.class).client = client;
  }

  public static Future<Void> run(CodyAgentClient client)
      throws IOException, ExecutionException, InterruptedException {
    Process process =
        new ProcessBuilder(
                "/Users/olafurpg/.asdf/shims/node",
                "/Users/olafurpg/dev/sourcegraph/sourcegraph/client/cody-agent/dist/agent.js")
            .redirectError(ProcessBuilder.Redirect.INHERIT)
            .start();
    PrintWriter traceWriter = null;
    String tracePath = System.getProperty("cody-agent.trace-path", "");
    if (!tracePath.isEmpty()) {
      Path trace = Paths.get(tracePath);
      try {
        Files.createDirectories(trace.getParent());
        traceWriter =
            new PrintWriter(
                Files.newOutputStream(
                    trace, StandardOpenOption.CREATE, StandardOpenOption.TRUNCATE_EXISTING));
      } catch (IOException e) {
        logger.error("unable to trace JSON-RPC debugging information to path " + tracePath, e);
      }
    }
    Launcher<CodyAgentServer> launcher =
        new Launcher.Builder<CodyAgentServer>()
            .setRemoteInterface(CodyAgentServer.class)
            .traceMessages(traceWriter)
            .setExecutorService(executorService)
            .setInput(process.getInputStream())
            .setOutput(process.getOutputStream())
            .setLocalService(client)
            .create();
    CodyAgentServer server = launcher.getRemoteProxy();
    client.server = server;

    // Very ugly, but sorta works, for now...
    executorService.submit(
        () -> {
          try {
            ServerInfo info = server.initialize(new ClientInfo("JetBrains")).get();
            logger.info("connected to Cody agent " + info.name);
            server.initialized();
          } catch (Exception e) {
            logger.error("failed to send 'initialize' JSON-RPC request Cody agent", e);
          }
        });

    return launcher.startListening();
  }
}
