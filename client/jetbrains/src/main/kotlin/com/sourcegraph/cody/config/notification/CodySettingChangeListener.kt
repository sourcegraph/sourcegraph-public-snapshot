package com.sourcegraph.cody.config.notification

import com.intellij.openapi.editor.Editor
import com.intellij.openapi.project.Project
import com.intellij.openapi.wm.ToolWindowManager
import com.sourcegraph.cody.CodyToolWindowContent
import com.sourcegraph.cody.CodyToolWindowFactory
import com.sourcegraph.cody.agent.CodyAgent
import com.sourcegraph.cody.agent.CodyAgentManager
import com.sourcegraph.cody.autocomplete.CodyAutocompleteManager
import com.sourcegraph.cody.autocomplete.render.AutocompleteRenderUtils
import com.sourcegraph.cody.statusbar.CodyAutocompleteStatusService
import com.sourcegraph.config.ConfigUtil
import com.sourcegraph.utils.CollectionUtil.Companion.diff
import java.util.function.Consumer

class CodySettingChangeListener(project: Project) : ChangeListener(project) {
  init {
    connection.subscribe(
        CodySettingChangeActionNotifier.TOPIC,
        object : CodySettingChangeActionNotifier {
          override fun afterAction(context: CodySettingChangeContext) {
            // Notify JCEF about the config changes
            javaToJSBridge?.callJS("pluginSettingsChanged", ConfigUtil.getConfigAsJson(project))

            // Notify Cody Agent about config changes.
            val agentServer = CodyAgent.getServer(project)
            if (ConfigUtil.isCodyEnabled() && agentServer != null) {
              agentServer.configurationDidChange(ConfigUtil.getAgentConfiguration(project))
            }

            if (context.newCodyEnabled) {
              // Starting the agent is idempotent, so it's OK if we call startAgent multiple times.
              CodyAgentManager.startAgent(project)
            } else {
              // Stopping the agent is idempotent, so it's OK if we call stopAgent multiple times.
              CodyAgentManager.stopAgent(project)
            }

            // clear autocomplete suggestions if freshly disabled
            if (context.oldCodyAutocompleteEnabled && !context.newCodyAutocompleteEnabled) {
              CodyAutocompleteManager.getInstance().clearAutocompleteSuggestionsForAllProjects()
            }

            // Disable/enable the Cody tool window depending on the setting
            if (!context.newCodyEnabled && context.oldCodyEnabled) {
              val toolWindowManager = ToolWindowManager.getInstance(project)
              val toolWindow = toolWindowManager.getToolWindow(CodyToolWindowFactory.TOOL_WINDOW_ID)
              toolWindow?.setAvailable(false, null)
            } else if (context.newCodyEnabled && !context.oldCodyEnabled) {
              val toolWindowManager = ToolWindowManager.getInstance(project)
              val toolWindow = toolWindowManager.getToolWindow(CodyToolWindowFactory.TOOL_WINDOW_ID)
              toolWindow?.setAvailable(true, null)
              val codyToolWindow = CodyToolWindowContent.getInstance(project)
              codyToolWindow.refreshPanelsVisibility()
            }

            CodyAutocompleteStatusService.resetApplication(project)

            // Rerender autocompletions when custom autocomplete color changed
            // or when checkbox state changed
            if (context.oldCustomAutocompleteColor != context.customAutocompleteColor ||
                (context.oldIsCustomAutocompleteColorEnabled !=
                    context.isCustomAutocompleteColorEnabled)) {
              ConfigUtil.getAllEditors()
                  .forEach(
                      Consumer { editor: Editor? ->
                        AutocompleteRenderUtils.rerenderAllAutocompleteInlays(editor)
                      })
            }

            // clear autocomplete inlays for blacklisted language editors
            val languageIdsToClear: List<String> =
                context.newBlacklistedAutocompleteLanguageIds.diff(
                    context.oldBlacklistedAutocompleteLanguageIds)
            if (languageIdsToClear.isNotEmpty())
                CodyAutocompleteManager.getInstance()
                    .clearAutocompleteSuggestionsForLanguageIds(languageIdsToClear)
          }
        })
  }
}
