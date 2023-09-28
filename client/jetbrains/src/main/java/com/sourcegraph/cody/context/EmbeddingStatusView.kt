package com.sourcegraph.cody.context

import com.intellij.openapi.fileEditor.FileEditorManagerListener
import com.intellij.openapi.project.Project
import com.intellij.ui.components.JBLabel
import com.intellij.util.ui.JBUI
import com.sourcegraph.cody.agent.CodyAgent
import com.sourcegraph.cody.agent.CodyAgentServer
import com.sourcegraph.cody.agent.protocol.GetRepoID
import com.sourcegraph.cody.auth.ui.EditCodebaseContextAction
import com.sourcegraph.cody.chat.ChatUIConstants
import java.awt.Dimension
import java.awt.FlowLayout
import javax.swing.Box
import javax.swing.JButton
import javax.swing.JPanel
import javax.swing.border.EmptyBorder

class EmbeddingStatusView(private val project: Project) : JPanel() {
  private val embeddingStatusContent: JBLabel
  private val codebaseSelector: JButton
  private val openedFileContent: JBLabel
  private var embeddingStatus: EmbeddingStatus

  init {
    setLayout(FlowLayout(FlowLayout.LEFT))
    val innerPanel = Box.createHorizontalBox()
    embeddingStatusContent = JBLabel()
    codebaseSelector = JButton(EditCodebaseContextAction(project))

    openedFileContent = JBLabel()
    openedFileContent.text = "No file selected"
    embeddingStatus = EmbeddingStatusNotAvailableYet()
    updateViewBasedOnStatus()
    innerPanel.add(embeddingStatusContent)
    innerPanel.add(codebaseSelector)
    innerPanel.add(Box.createHorizontalStrut(5))
    innerPanel.add(openedFileContent)
    innerPanel.setBorder(
        EmptyBorder(
            JBUI.insets(
                ChatUIConstants.TEXT_MARGIN,
                ChatUIConstants.TEXT_MARGIN,
                0,
                ChatUIConstants.TEXT_MARGIN)))
    this.add(innerPanel)
    updateEmbeddingStatus()
    project.messageBus
        .connect()
        .subscribe(
            FileEditorManagerListener.FILE_EDITOR_MANAGER, CurrentlyOpenFileListener(project, this))
  }

  fun updateEmbeddingStatus() {
    val client = CodyAgent.getClient(project)
    val repoName = client.codebase?.currentCodebase()
    if (repoName == null) {
      setEmbeddingStatus(NoGitRepositoryEmbeddingStatus())
    } else {
      setEmbeddingStatus(RepositoryMissingEmbeddingStatus(repoName))
      CodyAgent.getInitializedServer(project)
          .thenCompose { server: CodyAgentServer? ->
            server?.getRepoIdIfEmbeddingExists(GetRepoID(repoName))
          }
          .thenAccept { id: String? ->
            if (id != null) {
              setEmbeddingStatus(RepositoryIndexedEmbeddingStatus(repoName))
            }
          }
    }
  }

  private fun updateViewBasedOnStatus() {
    val codebaseName = embeddingStatus.getMainText()
    codebaseSelector.text = codebaseName.ifEmpty { "No repository" }
    val icon = embeddingStatus.getIcon()
    if (icon != null) {
      embeddingStatusContent.icon = icon
      embeddingStatusContent.preferredSize =
          Dimension(icon.iconWidth + 10, embeddingStatusContent.height)
    }
    val tooltip = embeddingStatus.getTooltip(project)
    if (tooltip.isNotEmpty()) {
      embeddingStatusContent.setToolTipText(tooltip)
    }
  }

  fun setEmbeddingStatus(embeddingStatus: EmbeddingStatus) {
    this.embeddingStatus = embeddingStatus
    updateViewBasedOnStatus()
  }

  fun setOpenedFileName(fileName: String, filePath: String?) {
    openedFileContent.setText(fileName)
    openedFileContent.setToolTipText(filePath)
  }
}
