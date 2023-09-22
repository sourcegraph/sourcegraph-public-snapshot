package com.sourcegraph.cody.auth.ui

import com.intellij.openapi.diagnostic.logger
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.DialogWrapper
import com.intellij.openapi.ui.ValidationInfo
import com.intellij.ui.components.fields.ExtendableTextField
import com.intellij.ui.dsl.builder.panel
import com.sourcegraph.cody.agent.CodyAgent
import com.sourcegraph.cody.agent.protocol.GetRepoID
import java.awt.event.ActionEvent
import java.util.concurrent.TimeUnit
import javax.swing.*

class EditCodebaseContextAction(val project: Project) : AbstractAction("Cody Context Selection") {
  val logger = logger<EditCodebaseContextAction>()

  private inner class EditCodebaseDialog : DialogWrapper(null, true) {
    val gitURL =
        ExtendableTextField(CodyAgent.getClient(project).codebase?.currentCodebase() ?: "", 40)

    init {
      init()
      title = "Context Selection"
      setValidationDelay(1000)
    }

    override fun doValidate(): ValidationInfo? {
      val repoName = gitURL.text
      if (repoName.isNotEmpty()) {
        try {
          val server = CodyAgent.getInitializedServer(project).get(1, TimeUnit.SECONDS)
          server?.getRepoId(GetRepoID(repoName))?.get(4, TimeUnit.SECONDS)
              ?: return ValidationInfo("Repository $repoName does not exist", gitURL)
        } catch (e: Exception) {
          logger.warn("failed to validate git url", e)
          return null
        }
      }
      return null
    }

    override fun doOKAction() {
      CodyAgent.getClient(project).codebase?.onNewExplicitCodebase(gitURL.text)
      super.doOKAction()
    }

    override fun createCenterPanel(): JComponent {
      return panel {
        row {
              label("Git URL:")
              cell(gitURL)
            }
            .rowComment("Example: github.com/sourcegraph/cody")
        row { text("The URL will be used to retrieve the code context from the Cody API.") }
        row {
          text(
              "If left empty, Cody will automatically infer the URL from your project's Git metadata.")
        }
      }
    }
  }

  override fun actionPerformed(e: ActionEvent?) {
    EditCodebaseDialog().showAndGet()
  }
}
