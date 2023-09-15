package com.sourcegraph.cody.context

import com.intellij.openapi.fileEditor.FileEditorManagerListener
import com.intellij.openapi.project.Project
import com.intellij.ui.SimpleColoredComponent
import com.intellij.ui.SimpleTextAttributes
import com.intellij.ui.components.JBLabel
import com.intellij.util.ui.JBUI
import com.sourcegraph.cody.agent.CodyAgent
import com.sourcegraph.cody.agent.CodyAgentServer
import com.sourcegraph.cody.agent.protocol.EmbeddingExistsParams
import com.sourcegraph.cody.chat.ChatUIConstants
import java.awt.FlowLayout
import javax.swing.Box
import javax.swing.JPanel
import javax.swing.border.EmptyBorder
import org.apache.commons.lang.StringUtils

class EmbeddingStatusView(private val project: Project) : JPanel() {
  private val embeddingStatusContent: SimpleColoredComponent
  private val openedFileContent: JBLabel
  private var embeddingStatus: EmbeddingStatus

  init {
    setLayout(FlowLayout(FlowLayout.LEFT))
    val innerPanel = Box.createHorizontalBox()
    embeddingStatusContent = SimpleColoredComponent()
    openedFileContent = JBLabel()
    openedFileContent.setText("No file selected")
    embeddingStatus = EmbeddingStatusNotAvailableYet()
    updateViewBasedOnStatus()
    innerPanel.add(embeddingStatusContent)
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
    val repoName = client.codebase?.currentCodebase ?: null
    if (repoName == null) {
      setEmbeddingStatus(NoGitRepositoryEmbeddingStatus())
    } else {
      setEmbeddingStatus(RepositoryMissingEmbeddingStatus(repoName))
      CodyAgent.getInitializedServer(project)
          .thenCompose { server: CodyAgentServer ->
            server.getRepoIdIfEmbeddingExists(EmbeddingExistsParams(repoName))
          }
          .thenAccept { id: String? ->
            if (id != null) {
              setEmbeddingStatus(RepositoryIndexedEmbeddingStatus(repoName))
            }
          }
    }
  }

  private fun updateViewBasedOnStatus() {
    embeddingStatusContent.clear()
    embeddingStatusContent.append(
        embeddingStatus.getMainText(), SimpleTextAttributes.REGULAR_ATTRIBUTES)
    val icon = embeddingStatus.getIcon()
    if (icon != null) {
      embeddingStatusContent.icon = icon
    }
    val tooltip = embeddingStatus.getTooltip(project)
    if (StringUtils.isNotEmpty(tooltip)) {
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
