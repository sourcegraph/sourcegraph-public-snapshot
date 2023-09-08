package com.sourcegraph.cody.config

import com.intellij.openapi.project.Project
import com.intellij.openapi.startup.StartupActivity
import com.sourcegraph.cody.CodyToolWindowContent
import com.sourcegraph.cody.chat.AssistantMessageWithSettingsButton
import java.awt.BorderLayout
import javax.swing.JPanel

class NotifyAboutNoDefaultAccount : StartupActivity {
  override fun runActivity(project: Project) {
    val defaultAccount = CodyAuthenticationManager.getInstance().getDefaultAccount(project)
    if (defaultAccount == null) {
      val codyToolWindowContent = CodyToolWindowContent.getInstance(project)
      val noAccessTokenText =
          """
                |<p>It looks like you don't have default Sourcegraph Account configured.</p>
                |<p>See our <a href="https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token">user docs</a> how to create an access token one and configure your default account in the settings to use Cody.</p>"""
              .trimMargin()
      val assistantMessageWithSettingsButton = AssistantMessageWithSettingsButton(noAccessTokenText)
      val messageContentPanel = JPanel(BorderLayout())
      messageContentPanel.add(assistantMessageWithSettingsButton)
      codyToolWindowContent.addComponentToChat(messageContentPanel)
    }
  }
}
