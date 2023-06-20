package com.sourcegraph.agent;

import com.intellij.ide.plugins.IdeaPluginDescriptor;
import com.intellij.ide.plugins.PluginManagerCore;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.extensions.PluginId;
import com.intellij.openapi.project.ProjectManager;
import com.intellij.openapi.util.SystemInfoRt;
import com.intellij.util.system.CpuArch;
import com.sourcegraph.agent.protocol.*;
import com.sourcegraph.config.ConfigUtil;
import java.io.File;
import java.io.IOException;
import java.io.PrintWriter;
import java.nio.file.*;
import java.util.Objects;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import org.eclipse.lsp4j.jsonrpc.Launcher;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/**
 * Orchestrator for the Cody agent, which is a Node.js program that implements the prompt logic for
 * Cody. The agent communicates via a JSON-RPC protocol that is documented in the file
 * "client/cody-agent/src/protocol.ts".
 */
public class CodyAgent {
  // TODO: actually stop the agent based on application lifecycle events
  public static Logger logger = Logger.getInstance(CodyAgent.class);
  private static final @NotNull PluginId PLUGIN_ID = PluginId.getId("com.sourcegraph.jetbrains");

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

  private static String cpuArchitecture() {
    if (System.getProperty("os.arch").toLowerCase().contains("arm")) {
      return "arm";
    } else if (System.getProperty("os.arch").toLowerCase().contains("x86")) {
      return "x86";
    } else if (System.getProperty("os.arch").toLowerCase().contains("amd64")) {
      return "amd64";
    } else {
      throw new UnsupportedOperationException(
          "Cody agent is not supported on your CPU architecture");
    }
  }

  private static String binarySuffix() {
    return SystemInfoRt.isWindows ? ".exe" : "";
  }

  private static String agentBinaryName() {
    String os = SystemInfoRt.isMac ? "macos" : SystemInfoRt.isWindows ? "windows" : "linux";
    @SuppressWarnings("MissingRecentApi")
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
      throw new CodyAgentException("Sourcegraph plugin path not found");
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
        logger.error("unable to trace JSON-RPC debugging information to path " + tracePath, e);
      }
    }
    return null;
  }

  public static Future<Void> run(CodyAgentClient client)
      throws IOException, CodyAgentException, ExecutionException, InterruptedException {
    File binary = agentBinary();
    logger.info("starting Cody agent " + binary.getAbsolutePath());
    Process process =
        new ProcessBuilder(binary.getAbsolutePath())
            .redirectError(ProcessBuilder.Redirect.INHERIT)
            .start();
    Launcher<CodyAgentServer> launcher =
        new Launcher.Builder<CodyAgentServer>()
            .setRemoteInterface(CodyAgentServer.class)
            .traceMessages(traceWriter())
            .setExecutorService(executorService)
            .setInput(process.getInputStream())
            .setOutput(process.getOutputStream())
            .setLocalService(client)
            .create();
    CodyAgentServer server = launcher.getRemoteProxy();
    client.server = server;

    executorService.submit(
        () -> {
          try {
            ServerInfo info =
                server
                    .initialize(
                        new ClientInfo()
                            .setName("JetBrains")
                            .setVersion(ConfigUtil.getPluginVersion())
                            .setWorkspaceRootPath(ConfigUtil.getWorkspaceRoot())
                            .setConnectionConfiguration(
                                ConfigUtil.getAgentConfiguration(
                                    ProjectManager.getInstance().getDefaultProject())))
                    .get();
            logger.info("connected to Cody agent " + info.name);
            server.initialized();
            server.configurationDidChange(
                ConfigUtil.getAgentConfiguration(ProjectManager.getInstance().getDefaultProject()));
          } catch (Exception e) {
            logger.error("failed to send 'initialize' JSON-RPC request Cody agent", e);
          }
        });

    return launcher.startListening();
  }
}
