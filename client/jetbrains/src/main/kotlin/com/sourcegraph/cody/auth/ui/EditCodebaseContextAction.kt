package com.sourcegraph.cody.auth.ui

import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.diagnostic.logger
import com.intellij.openapi.progress.EmptyProgressIndicator
import com.intellij.openapi.progress.ProgressIndicator
import com.intellij.openapi.progress.ProgressManager
import com.intellij.openapi.progress.Task.Backgroundable
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.DialogWrapper
import com.intellij.ui.DocumentAdapter
import com.intellij.ui.components.fields.ExtendableTextField
import com.intellij.ui.dsl.builder.panel
import com.sourcegraph.cody.agent.CodyAgent
import com.sourcegraph.cody.agent.protocol.GetRepoID
import com.sourcegraph.cody.ui.LoadingLayerUI
import java.awt.event.ActionEvent
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicReference
import javax.swing.AbstractAction
import javax.swing.JComponent
import javax.swing.JLayer
import javax.swing.event.DocumentEvent

class EditCodebaseContextAction(val project: Project) : AbstractAction("Cody Context Selection") {
  val logger = logger<EditCodebaseContextAction>()
  private val currentIndicator: AtomicReference<ProgressIndicator> = AtomicReference()

  private inner class EditCodebaseDialog : DialogWrapper(null, true) {
    val gitURL =
        ExtendableTextField(CodyAgent.getClient(project).codebase?.currentCodebase() ?: "", 40)
    val loadingLayerUI = LoadingLayerUI()
    val layeredGitURL = JLayer(gitURL, loadingLayerUI)

    init {
      init()
      title = "Context Selection"
      setValidationDelay(1000)

      gitURL.document.addDocumentListener(
          object : DocumentAdapter() {
            override fun textChanged(e: DocumentEvent) {
              isOKActionEnabled = false
              validateInput()
            }
          })
    }

    private fun validateInput() {
      val repoName = gitURL.text
      currentIndicator.get()?.cancel()

      ProgressManager.getInstance()
          .runProcessWithProgressAsynchronously(
              object : Backgroundable(project, "Validating git url") {
                override fun run(indicator: ProgressIndicator) {
                  currentIndicator.set(indicator)
                  if (indicator.isCanceled) {
                    return
                  }
                  try {
                    loadingLayerUI.startLoading()
                    val server = CodyAgent.getInitializedServer(project).get(1, TimeUnit.SECONDS)
                    val isRepoIdValid =
                        server?.getRepoId(GetRepoID(repoName))?.get(4, TimeUnit.SECONDS) != null
                    setOKButtonIcon(null)
                    loadingLayerUI.stopLoading()
                    ApplicationManager.getApplication().invokeLater {
                      if (!indicator.isCanceled) {
                        if (isRepoIdValid) {
                          isOKActionEnabled = true
                          setErrorText(null, gitURL)
                        } else {
                          isOKActionEnabled = false
                          setErrorText("Repository $repoName does not exist", gitURL)
                        }
                      }
                    }
                  } catch (e: Exception) {
                    if (!indicator.isCanceled) {
                      ApplicationManager.getApplication().invokeLater {
                        isOKActionEnabled = false
                        setErrorText("Failed to validate git url", gitURL)
                      }
                    }
                    logger.warn("Failed to validate git url", e)
                  } finally {
                    loadingLayerUI.stopLoading()
                  }
                }
              },
              EmptyProgressIndicator())
    }

    override fun doOKAction() {
      CodyAgent.getClient(project).codebase?.onNewExplicitCodebase(gitURL.text)
      super.doOKAction()
    }

    override fun createCenterPanel(): JComponent {
      return panel {
        row {
              label("Git URL:")
              cell(layeredGitURL)
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
