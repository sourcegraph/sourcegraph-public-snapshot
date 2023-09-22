package com.sourcegraph.cody.config.notification

import com.intellij.openapi.project.Project
import com.sourcegraph.cody.CodyToolWindowContent
import com.sourcegraph.cody.agent.CodyAgent
import com.sourcegraph.cody.agent.CodyAgentManager
import com.sourcegraph.cody.config.CodyApplicationSettings.Companion.getInstance
import com.sourcegraph.config.ConfigUtil
import com.sourcegraph.telemetry.GraphQlLogger

class AccountSettingChangeListener(project: Project) : ChangeListener(project) {
  init {
    connection.subscribe(
        AccountSettingChangeActionNotifier.TOPIC,
        object : AccountSettingChangeActionNotifier {
          override fun beforeAction(serverUrlChanged: Boolean) {
            val codyApplicationSettings = getInstance()
            if (serverUrlChanged) {
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
              CodyAgentManager.startAgent(project)
            } else {
              // Stopping the agent is idempotent, so it's OK if we call stopAgent multiple times.
              CodyAgentManager.stopAgent(project)
            }

            // Notify Cody Agent about config changes.
            val agentServer = CodyAgent.getServer(project)
            if (ConfigUtil.isCodyEnabled() && agentServer != null) {
              agentServer.configurationDidChange(ConfigUtil.getAgentConfiguration(project))
            }

            // Refresh onboarding panels
            if (ConfigUtil.isCodyEnabled()) {
              val codyToolWindowContent = CodyToolWindowContent.getInstance(project)
              codyToolWindowContent.refreshPanelsVisibility()
            }

            // Log install events
            if (context.serverUrlChanged) {
              GraphQlLogger.logInstallEvent(project).thenAccept { e ->
                codyApplicationSettings.isInstallEventLogged = e
              }
            } else if (context.accessTokenChanged &&
                !codyApplicationSettings.isInstallEventLogged) {
              GraphQlLogger.logInstallEvent(project).thenAccept { e ->
                codyApplicationSettings.isInstallEventLogged = e
              }
            }
          }
        })
  }
}
