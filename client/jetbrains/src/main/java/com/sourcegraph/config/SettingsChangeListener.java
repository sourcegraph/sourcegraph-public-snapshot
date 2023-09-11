package com.sourcegraph.config;

import com.intellij.openapi.Disposable;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.wm.ToolWindow;
import com.intellij.openapi.wm.ToolWindowManager;
import com.intellij.util.messages.MessageBus;
import com.intellij.util.messages.MessageBusConnection;
import com.sourcegraph.cody.CodyAgentProjectListener;
import com.sourcegraph.cody.CodyToolWindowFactory;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.agent.CodyAgentServer;
import com.sourcegraph.cody.autocomplete.CodyAutocompleteManager;
import com.sourcegraph.cody.autocomplete.render.AutocompleteRenderUtils;
import com.sourcegraph.cody.config.CodyApplicationSettings;
import com.sourcegraph.cody.statusbar.CodyAutocompleteStatus;
import com.sourcegraph.cody.statusbar.CodyAutocompleteStatusService;
import com.sourcegraph.find.browser.JavaToJSBridge;
import com.sourcegraph.telemetry.GraphQlLogger;
import com.sourcegraph.utils.CollectionUtil;
import java.util.List;
import java.util.Objects;
import org.jetbrains.annotations.NotNull;

/**
 * Listens to changes in the plugin settings and: - notifies JCEF about them. - logs
 * install/uninstall events. - notifies the user about a successful connection.
 */
public class SettingsChangeListener implements Disposable {
  private final MessageBusConnection connection;
  private JavaToJSBridge javaToJSBridge;
  private final Logger logger = Logger.getInstance(SettingsChangeListener.class);

  public SettingsChangeListener(@NotNull Project project) {
    MessageBus bus = project.getMessageBus();

    connection = bus.connect();
    connection.subscribe(
        PluginSettingChangeActionNotifier.TOPIC,
        new PluginSettingChangeActionNotifier() {
          @Override
          public void beforeAction(boolean serverUrlChanged) {
            CodyApplicationSettings codyApplicationSettings = CodyApplicationSettings.getInstance();
            if (serverUrlChanged) {
              GraphQlLogger.logUninstallEvent(project);
              codyApplicationSettings.setInstallEventLogged(false);
            }
          }

          @Override
          public void afterAction(@NotNull PluginSettingChangeContext context) {
            CodyApplicationSettings codyApplicationSettings = CodyApplicationSettings.getInstance();
            // Notify JCEF about the config changes
            if (javaToJSBridge != null) {
              javaToJSBridge.callJS("pluginSettingsChanged", ConfigUtil.getConfigAsJson(project));
            }

            if (context.newCodyEnabled) {
              // Starting the agent is idempotent, so it's OK if we call startAgent multiple times.
              new CodyAgentProjectListener().startAgent(project);
            } else {
              // Stopping the agent is idempotent, so it's OK if we call stopAgent multiple times.
              new CodyAgentProjectListener().stopAgent(project);
            }

            // Notify Cody Agent about config changes.
            CodyAgentServer agentServer = CodyAgent.getServer(project);
            if (context.newCodyEnabled && agentServer != null) {
              agentServer.configurationDidChange(ConfigUtil.getAgentConfiguration(project));
            }

            // Log install events
            if (context.serverUrlChanged) {
              GraphQlLogger.logInstallEvent(project)
                  .thenAccept(codyApplicationSettings::setInstallEventLogged);
            } else if (context.accessTokenChanged
                && !codyApplicationSettings.isInstallEventLogged()) {
              GraphQlLogger.logInstallEvent(project)
                  .thenAccept(codyApplicationSettings::setInstallEventLogged);
            }

            // clear autocomplete suggestions if freshly disabled
            if (context.oldCodyAutocompleteEnabled && !context.newCodyAutocompleteEnabled) {
              CodyAutocompleteManager.getInstance().clearAutocompleteSuggestionsForAllProjects();
            }

            // Disable/enable the Cody tool window depending on the setting
            if (!context.newCodyEnabled && context.oldCodyEnabled) {
              ToolWindowManager toolWindowManager = ToolWindowManager.getInstance(project);
              ToolWindow toolWindow =
                  toolWindowManager.getToolWindow(CodyToolWindowFactory.TOOL_WINDOW_ID);
              if (toolWindow != null) {
                toolWindow.setAvailable(false, null);
              }
            } else if (context.newCodyEnabled && !context.oldCodyEnabled) {
              ToolWindowManager toolWindowManager = ToolWindowManager.getInstance(project);
              ToolWindow toolWindow =
                  toolWindowManager.getToolWindow(CodyToolWindowFactory.TOOL_WINDOW_ID);
              if (toolWindow != null) {
                toolWindow.setAvailable(true, null);
              }
            }
            if (!context.newCodyEnabled) {
              CodyAutocompleteStatusService.notifyApplication(CodyAutocompleteStatus.CodyDisabled);
            } else if (!context.newCodyAutocompleteEnabled) {
              CodyAutocompleteStatusService.notifyApplication(
                  CodyAutocompleteStatus.AutocompleteDisabled);
            } else {
              CodyAutocompleteStatusService.notifyApplication(CodyAutocompleteStatus.Ready);
            }

            // Rerender autocompletions when custom autocomplete color changed
            // or when checkbox state changed
            if (!Objects.equals(context.oldCustomAutocompleteColor, context.customAutocompleteColor)
                || (context.oldIsCustomAutocompleteColorEnabled
                    != context.isCustomAutocompleteColorEnabled)) {
              ConfigUtil.getAllEditors()
                  .forEach(AutocompleteRenderUtils::rerenderAllAutocompleteInlays);
            }

            // clear autocomplete inlays for blacklisted language editors
            List<String> languageIdsToClear =
                CollectionUtil.Companion.diff(
                    context.newBlacklistedAutocompleteLanguageIds,
                    context.oldBlacklistedAutocompleteLanguageIds);
            if (!languageIdsToClear.isEmpty())
              CodyAutocompleteManager.getInstance()
                  .clearAutocompleteSuggestionsForLanguageIds(languageIdsToClear);
          }
        });
  }

  public void setJavaToJSBridge(JavaToJSBridge javaToJSBridge) {
    this.javaToJSBridge = javaToJSBridge;
  }

  @Override
  public void dispose() {
    connection.disconnect();
  }
}
