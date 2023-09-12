package com.sourcegraph.cody.config.notification

import com.intellij.openapi.Disposable
import com.intellij.openapi.diagnostic.Logger
import com.intellij.openapi.project.Project
import com.intellij.util.messages.MessageBusConnection
import com.sourcegraph.cody.CodyAgentProjectListener
import com.sourcegraph.cody.agent.CodyAgent
import com.sourcegraph.cody.config.CodyApplicationSettings.Companion.getInstance
import com.sourcegraph.config.ConfigUtil
import com.sourcegraph.find.browser.JavaToJSBridge
import com.sourcegraph.telemetry.GraphQlLogger

class AccountSettingChangeListener(val project: Project) : Disposable {
    var connection: MessageBusConnection? = null
    var javaToJSBridge: JavaToJSBridge? = null
    val logger = Logger.getInstance(AccountSettingChangeListener::class.java)

    init {
        val bus = project.messageBus
        connection = bus.connect()
        connection?.subscribe(
            AccountSettingChangeActionNotifier.TOPIC,
            object : AccountSettingChangeActionNotifier {
                override fun beforeAction(context: AccountSettingChangeContext) {
                    val codyApplicationSettings = getInstance()
                    if (context.serverUrlChanged) {
                        GraphQlLogger.logUninstallEvent(project)
                        codyApplicationSettings.isInstallEventLogged = false
                    }
                }

                override fun afterAction(context: AccountSettingChangeContext) {
                    val codyApplicationSettings = getInstance()
                    // Notify JCEF about the config changes
                    javaToJSBridge?.callJS("pluginSettingsChanged", ConfigUtil.getConfigAsJson(project))

                    if (ConfigUtil.isCodyEnabled()) {
                        // Starting the agent is idempotent, so it's OK if we call startAgent multiple times.
                        CodyAgentProjectListener.startAgent(project)
                    } else {
                        // Stopping the agent is idempotent, so it's OK if we call stopAgent multiple times.
                        CodyAgentProjectListener.stopAgent(project)
                    }

                    // Notify Cody Agent about config changes.
                    val agentServer = CodyAgent.getServer(project)
                    if (ConfigUtil.isCodyEnabled() && agentServer != null) {
                        agentServer.configurationDidChange(ConfigUtil.getAgentConfiguration(project))
                    }

                    // Log install events
                    if (context.serverUrlChanged) {
                        GraphQlLogger.logInstallEvent(project)
                            .thenAccept { e -> codyApplicationSettings.isInstallEventLogged = e }
                    } else if (context.accessTokenChanged
                        && !codyApplicationSettings.isInstallEventLogged) {
                        GraphQlLogger.logInstallEvent(project)
                            .thenAccept { e -> codyApplicationSettings.isInstallEventLogged = e }
                    }
                }
            })
    }

    override fun dispose() {
        connection!!.disconnect()
    }
}
