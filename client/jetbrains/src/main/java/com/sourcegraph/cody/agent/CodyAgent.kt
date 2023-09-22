package com.sourcegraph.cody.agent

import com.google.gson.GsonBuilder
import com.intellij.ide.plugins.PluginManagerCore
import com.intellij.openapi.Disposable
import com.intellij.openapi.diagnostic.Logger
import com.intellij.openapi.editor.EditorFactory
import com.intellij.openapi.editor.ex.EditorEventMulticasterEx
import com.intellij.openapi.extensions.PluginId
import com.intellij.openapi.project.Project
import com.intellij.openapi.util.Disposer
import com.intellij.openapi.util.SystemInfoRt
import com.intellij.util.system.CpuArch
import com.sourcegraph.cody.CodyAgentFocusListener
import com.sourcegraph.cody.agent.protocol.ClientInfo
import com.sourcegraph.config.ConfigUtil
import java.io.File
import java.io.IOException
import java.io.PrintWriter
import java.nio.file.*
import java.util.*
import java.util.concurrent.CompletableFuture
import java.util.concurrent.Executors
import java.util.concurrent.Future
import java.util.concurrent.atomic.AtomicBoolean
import java.util.function.Function
import org.eclipse.lsp4j.jsonrpc.Launcher

/**
 * Orchestrator for the Cody agent, which is a Node.js program that implements the prompt logic for
 * Cody. The agent communicates via a JSON-RPC protocol that is documented in the file
 * "cody/agent/src/protocol.ts".
 *
 * The class {{[com.sourcegraph.cody.CodyAgentProjectListener]}} is responsible for initializing and
 * shutting down the agent.
 */
class CodyAgent(private val project: Project) : Disposable {
  var disposable = Disposer.newDisposable("CodyAgent")
  private val client = CodyAgentClient()
  private var agentNotRunningExplanation = ""
  private var initialized = CompletableFuture<CodyAgentServer?>()
  private val firstConnection = AtomicBoolean(true)
  private var listeningToJsonRpc: Future<Void?> = CompletableFuture.completedFuture(null)
  private var agentProcess: Process? = null

  fun initialize() {
    if ("true" != System.getProperty("cody-agent.enabled", "true")) {
      logger.info("Cody agent is disabled due to system property '-Dcody-agent.enabled=false'")
      return
    }
    try {
      val isFirstConnection = firstConnection.getAndSet(false)
      if (!isFirstConnection) {
        // Restart `initialized` future so that new callers can subscribe to the next instance of
        // the Cody agent server.
        initialized = CompletableFuture()
      }
      agentNotRunningExplanation = ""
      startListeningToAgent()
      executorService.submit {
        try {
          val server = client.server
          if (server == null) {
            return@submit
          }
          val info =
              server
                  .initialize(
                      ClientInfo()
                          .setName("JetBrains")
                          .setVersion(ConfigUtil.getPluginVersion())
                          .setWorkspaceRootPath(ConfigUtil.getWorkspaceRoot(project))
                          .setExtensionConfiguration(ConfigUtil.getAgentConfiguration(project)))
                  .get()
          logger.info("connected to Cody agent " + info.name)
          server.initialized()
          subscribeToFocusEvents()
          initialized.complete(server)
        } catch (e: Exception) {
          agentNotRunningExplanation = "failed to send 'initialize' JSON-RPC request Cody agent"
          logger.warn(agentNotRunningExplanation, e)
        }
      }
    } catch (e: Exception) {
      agentNotRunningExplanation = "unable to start Cody agent"
      logger.warn(agentNotRunningExplanation, e)
    }
  }

  fun subscribeToFocusEvents() {
    // Code example taken from
    // https://intellij-support.jetbrains.com/hc/en-us/community/posts/4578776718354/comments/4594838404882
    // This listener is registered programmatically because it was not working via plugin.xml
    // listeners.
    val multicaster = EditorFactory.getInstance().eventMulticaster
    if (multicaster is EditorEventMulticasterEx) {
      try {
        multicaster.addFocusChangeListener(CodyAgentFocusListener(), disposable)
      } catch (ignored: Exception) {
        // Ignore exception https://github.com/sourcegraph/sourcegraph/issues/56032
      }
    }
  }

  fun shutdown() {
    val server = getServer(project) ?: return
    executorService.submit<CompletableFuture<Void>> {
      server.shutdown().thenAccept {
        server.exit()
        agentNotRunningExplanation = "Cody Agent shut down"
        listeningToJsonRpc.cancel(true)
      }
    }
  }

  @Throws(IOException::class, CodyAgentException::class)
  private fun startListeningToAgent() {
    val binary = agentBinary()
    logger.info("starting Cody agent " + binary.absolutePath)
    val processBuilder = ProcessBuilder(binary.absolutePath)
    if (java.lang.Boolean.getBoolean("cody.accept-non-trusted-certificates-automatically")) {
      processBuilder.environment()["NODE_TLS_REJECT_UNAUTHORIZED"] = "0"
    }

    val process =
        processBuilder
            .redirectErrorStream(false)
            .redirectError(ProcessBuilder.Redirect.PIPE)
            .start()

    // Redirect agent stderr into idea.log by buffering line by line into `logger.warn()`
    // statements. Without this logic, the stderr output of the agent process is lost if the process
    // fails to start for some reason. We use `logger.warn()` because the agent shouldn't print much
    // normally (excluding a few noisy messages during initialization), it's mostly used to report
    // unexpected errors.
    Thread { process.errorStream.bufferedReader().forEachLine { line -> logger.warn(line) } }
        .start()

    agentProcess = process
    val launcher =
        Launcher.Builder<
                CodyAgentServer>() // emit `null` instead of leaving fields undefined because Cody
            // in VSC has
            // many `=== null` checks that return false for undefined fields.
            .configureGson { obj: GsonBuilder -> obj.serializeNulls() }
            .setRemoteInterface(CodyAgentServer::class.java)
            .traceMessages(traceWriter())
            .setExecutorService(executorService)
            .setInput(process.inputStream)
            .setOutput(process.outputStream)
            .setLocalService(client)
            .create()
    val server = launcher.remoteProxy
    client.server = server
    client.documents = CodyAgentDocuments(server)
    client.codebase = CodyAgentCodebase(server, project)
    listeningToJsonRpc = launcher.startListening()
  }

  override fun dispose() {
    disposable.dispose()
  }

  companion object {
    var logger = Logger.getInstance(CodyAgent::class.java)
    private val PLUGIN_ID = PluginId.getId("com.sourcegraph.jetbrains")
    @JvmField val executorService = Executors.newCachedThreadPool()

    @JvmStatic
    fun getClient(project: Project): CodyAgentClient {
      return project.getService(CodyAgent::class.java).client
    }

    @JvmStatic
    fun getInitializedServer(project: Project): CompletableFuture<CodyAgentServer?> {
      return project.getService(CodyAgent::class.java).initialized
    }

    @JvmStatic
    fun isConnected(project: Project): Boolean {
      val agent = project.getService(CodyAgent::class.java)
      // NOTE(olafurpg): there are probably too many conditions below. We test multiple conditions
      // because we don't know 100% yet what exactly constitutes a "connected" state. Out of
      // abundance
      // of caution, we check everything we can think of.
      return (agent?.agentProcess != null &&
          agent.agentProcess!!.isAlive &&
          !agent.listeningToJsonRpc.isDone &&
          !agent.listeningToJsonRpc.isCancelled &&
          agent.client.server != null)
    }

    @JvmStatic
    fun <T> withServer(
        project: Project,
        callback: Function<CodyAgentServer?, CompletableFuture<T>?>?
    ): CompletableFuture<T> {
      return getInitializedServer(project).thenCompose(callback)
    }

    @JvmStatic
    fun getServer(project: Project): CodyAgentServer? {
      return if (!isConnected(project)) {
        null
      } else getClient(project).server
    }

    private fun binarySuffix(): String {
      return if (SystemInfoRt.isWindows) ".exe" else ""
    }

    private fun agentBinaryName(): String {
      val os = if (SystemInfoRt.isMac) "macos" else if (SystemInfoRt.isWindows) "win" else "linux"
      val arch = if (CpuArch.isArm64()) "arm64" else "x64"
      return "agent-" + os + "-" + arch + binarySuffix()
    }

    private fun agentDirectory(): Path? {
      val fromProperty = System.getProperty("cody-agent.directory", "")
      if (!fromProperty.isEmpty()) {
        return Paths.get(fromProperty)
      }
      val plugin = PluginManagerCore.getPlugin(PLUGIN_ID) ?: return null
      return plugin.pluginPath
    }

    @Throws(CodyAgentException::class)
    private fun agentBinary(): File {
      val pluginPath =
          agentDirectory()
              ?: throw CodyAgentException("Sourcegraph Cody + Code Search plugin path not found")
      val binarySource = pluginPath.resolve("agent").resolve(agentBinaryName())
      if (!Files.isRegularFile(binarySource)) {
        throw CodyAgentException(
            "Cody agent binary not found at path " + binarySource.toAbsolutePath())
      }
      return try {
        val binaryTarget = Files.createTempFile("cody-agent", binarySuffix())
        logger.info("extracting Cody agent binary to " + binaryTarget.toAbsolutePath())
        Files.copy(binarySource, binaryTarget, StandardCopyOption.REPLACE_EXISTING)
        val binary = binaryTarget.toFile()
        if (binary.setExecutable(true)) {
          binary.deleteOnExit()
          binary
        } else {
          throw CodyAgentException("failed to make executable " + binary.absolutePath)
        }
      } catch (e: IOException) {
        throw CodyAgentException("failed to create agent binary", e)
      }
    }

    private fun traceWriter(): PrintWriter? {
      val tracePath = System.getProperty("cody-agent.trace-path", "")
      if (!tracePath.isEmpty()) {
        val trace = Paths.get(tracePath)
        try {
          Files.createDirectories(trace.parent)
          return PrintWriter(
              Files.newOutputStream(
                  trace, StandardOpenOption.CREATE, StandardOpenOption.TRUNCATE_EXISTING))
        } catch (e: IOException) {
          logger.warn("unable to trace JSON-RPC debugging information to path $tracePath", e)
        }
      }
      return null
    }
  }
}
