package com.sourcegraph.cody.agent;

import com.google.gson.GsonBuilder;
import com.intellij.ide.plugins.IdeaPluginDescriptor;
import com.intellij.ide.plugins.PluginManagerCore;
import com.intellij.openapi.Disposable;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.editor.EditorFactory;
import com.intellij.openapi.editor.event.EditorEventMulticaster;
import com.intellij.openapi.editor.ex.EditorEventMulticasterEx;
import com.intellij.openapi.extensions.PluginId;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.Disposer;
import com.intellij.openapi.util.SystemInfoRt;
import com.intellij.util.system.CpuArch;
import com.sourcegraph.cody.CodyAgentFocusListener;
import com.sourcegraph.cody.agent.protocol.ClientInfo;
import com.sourcegraph.cody.agent.protocol.ServerInfo;
import com.sourcegraph.config.ConfigUtil;
import java.io.File;
import java.io.IOException;
import java.io.PrintWriter;
import java.nio.file.*;
import java.util.Objects;
import java.util.concurrent.*;
import org.eclipse.lsp4j.jsonrpc.Launcher;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/**
 * Orchestrator for the Cody agent, which is a Node.js program that implements the prompt logic for
 * Cody. The agent communicates via a JSON-RPC protocol that is documented in the file
 * "cody/agent/src/protocol.ts".
 *
 * <p>The class {{{@link com.sourcegraph.cody.CodyAgentProjectListener}}} is responsible for
 * initializing and shutting down the agent.
 */
public class CodyAgent implements Disposable {
  public static Logger logger = Logger.getInstance(CodyAgent.class);
  private static final @NotNull PluginId PLUGIN_ID = PluginId.getId("com.sourcegraph.jetbrains");
  public static final ExecutorService executorService = Executors.newCachedThreadPool();

  Disposable disposable = Disposer.newDisposable("CodyAgent");
  private final @NotNull Project project;
  private final CodyAgentClient client = new CodyAgentClient();
  private String initializationErrorMessage = "";
  private final CompletableFuture<CodyAgentServer> initialized = new CompletableFuture<>();
  private Future<Void> listeningToJsonRpc;

  public CodyAgent(@NotNull Project project) {
    this.project = project;
  }

  @NotNull
  public static CodyAgentClient getClient(@NotNull Project project) {
    return project.getService(CodyAgent.class).client;
  }

  @NotNull
  public static CompletableFuture<CodyAgentServer> getInitializedServer(@NotNull Project project) {
    return project.getService(CodyAgent.class).initialized;
  }

  @SuppressWarnings("BooleanMethodIsAlwaysInverted")
  public static boolean isConnected(@NotNull Project project) {
    CodyAgent agent = project.getService(CodyAgent.class);
    return agent != null
        && agent.initializationErrorMessage.isEmpty()
        && agent.listeningToJsonRpc != null
        && !agent.listeningToJsonRpc.isDone()
        && !agent.listeningToJsonRpc.isCancelled()
        && agent.client.server != null;
  }

  @Nullable
  public static CodyAgentServer getServer(@NotNull Project project) {
    if (!isConnected(project)) {
      return null;
    }
    return getClient(project).server;
  }

  public static CodyAgentCodebase getCodebase(@NotNull Project project) {
    if (!isConnected(project)) {
      return null;
    }
    return getClient(project).codebase;
  }

  public void initialize() {
    if (!"true".equals(System.getProperty("cody-agent.enabled", "false"))) {
      logger.info("Cody agent is disabled due to system property '-Dcody-agent.enabled=false'");
      return;
    }
    try {
      startListeningToAgent();
      executorService.submit(
          () -> {
            try {
              final CodyAgentServer server = Objects.requireNonNull(client.server);
              ServerInfo info =
                  server
                      .initialize(
                          new ClientInfo()
                              .setName("JetBrains")
                              .setVersion(ConfigUtil.getPluginVersion())
                              .setWorkspaceRootPath(ConfigUtil.getWorkspaceRoot(project))
                              .setExtensionConfiguration(
                                  ConfigUtil.getAgentConfiguration(this.project)))
                      .get();
              logger.info("connected to Cody agent " + info.name);
              server.initialized();
              this.subscribeToFocusEvents();
              this.initialized.complete(server);
            } catch (Exception e) {
              initializationErrorMessage =
                  "failed to send 'initialize' JSON-RPC request Cody agent";
              logger.warn(initializationErrorMessage, e);
            }
          });
    } catch (Exception e) {
      initializationErrorMessage = "unable to start Cody agent";
      logger.warn(initializationErrorMessage, e);
    }
  }

  public void subscribeToFocusEvents() {
    // Code example taken from
    // https://intellij-support.jetbrains.com/hc/en-us/community/posts/4578776718354/comments/4594838404882
    // This listener is registered programmatically because it was not working via plugin.xml
    // listeners.
    EditorEventMulticaster multicaster = EditorFactory.getInstance().getEventMulticaster();
    if (multicaster instanceof EditorEventMulticasterEx) {
      EditorEventMulticasterEx ex = (EditorEventMulticasterEx) multicaster;
      ex.addFocusChangeListener(new CodyAgentFocusListener(), this.disposable);
    }
  }

  public void shutdown() {
    final CodyAgentServer server = CodyAgent.getServer(project);
    if (server == null) {
      return;
    }
    executorService.submit(() -> server.shutdown().thenAccept((Void) -> server.exit()));
  }

  private static String binarySuffix() {
    return SystemInfoRt.isWindows ? ".exe" : "";
  }

  private static String agentBinaryName() {
    String os = SystemInfoRt.isMac ? "macos" : SystemInfoRt.isWindows ? "windows" : "linux";
    String arch = CpuArch.isArm64() ? "arm64" : "x64";
    return "agent-" + os + "-" + arch + binarySuffix();
  }

  @Nullable
  private static Path agentDirectory() {
    String fromProperty = System.getProperty("cody-agent.directory", "");
    if (!fromProperty.isEmpty()) {
      return Paths.get(fromProperty);
    }
    IdeaPluginDescriptor plugin = PluginManagerCore.getPlugin(PLUGIN_ID);
    if (plugin == null) {
      return null;
    }
    return plugin.getPluginPath();
  }

  @NotNull
  private static File agentBinary() throws CodyAgentException {
    Path pluginPath = agentDirectory();
    if (pluginPath == null) {
      throw new CodyAgentException("Cody AI by Sourcegraph plugin path not found");
    }
    Path binarySource = pluginPath.resolve("agent").resolve(agentBinaryName());
    if (!Files.isRegularFile(binarySource)) {
      throw new CodyAgentException(
          "Cody agent binary not found at path " + binarySource.toAbsolutePath());
    }
    try {
      Path binaryTarget = Files.createTempFile("cody-agent", binarySuffix());
      logger.info("extracting Cody agent binary to " + binaryTarget.toAbsolutePath());
      Files.copy(binarySource, binaryTarget, StandardCopyOption.REPLACE_EXISTING);
      File binary = binaryTarget.toFile();
      if (binary.setExecutable(true)) {
        binary.deleteOnExit();
        return binary;
      } else {
        throw new CodyAgentException("failed to make executable " + binary.getAbsolutePath());
      }
    } catch (IOException e) {
      throw new CodyAgentException("failed to create agent binary", e);
    }
  }

  @Nullable
  private static PrintWriter traceWriter() {
    String tracePath = System.getProperty("cody-agent.trace-path", "");
    if (!tracePath.isEmpty()) {
      Path trace = Paths.get(tracePath);
      try {
        Files.createDirectories(trace.getParent());
        return new PrintWriter(
            Files.newOutputStream(
                trace, StandardOpenOption.CREATE, StandardOpenOption.TRUNCATE_EXISTING));
      } catch (IOException e) {
        logger.warn("unable to trace JSON-RPC debugging information to path " + tracePath, e);
      }
    }
    return null;
  }

  private void startListeningToAgent() throws IOException, CodyAgentException {
    File binary = agentBinary();
    logger.info("starting Cody agent " + binary.getAbsolutePath());
    Process process =
        new ProcessBuilder(binary.getAbsolutePath())
            .redirectError(ProcessBuilder.Redirect.INHERIT)
            .start();
    Launcher<CodyAgentServer> launcher =
        new Launcher.Builder<CodyAgentServer>()
            // emit `null` instead of leaving fields undefined because Cody in VSC has
            // many `=== null` checks that return false for undefined fields.
            .configureGson(GsonBuilder::serializeNulls)
            .setRemoteInterface(CodyAgentServer.class)
            .traceMessages(traceWriter())
            .setExecutorService(executorService)
            .setInput(process.getInputStream())
            .setOutput(process.getOutputStream())
            .setLocalService(client)
            .create();
    client.server = launcher.getRemoteProxy();
    client.documents = new CodyAgentDocuments(client.server);
    client.codebase = new CodyAgentCodebase(client.server);
    this.listeningToJsonRpc = launcher.startListening();
  }

  @Override
  public void dispose() {
    this.disposable.dispose();
  }
}
