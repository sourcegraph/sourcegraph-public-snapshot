package com.sourcegraph.cody.chat

import com.intellij.openapi.project.Project
import com.intellij.ui.ColorUtil
import com.intellij.util.ui.SwingHelper
import com.intellij.util.ui.UIUtil
import com.sourcegraph.cody.api.Speaker
import com.sourcegraph.cody.ui.HtmlViewer.createHtmlViewer
import java.awt.Color
import javax.swing.JEditorPane
import javax.swing.JPanel
import org.commonmark.ext.gfm.tables.TablesExtension
import org.commonmark.node.Node
import org.commonmark.parser.Parser
import org.commonmark.renderer.html.HtmlRenderer

class MessagePanel(
    private val chatMessage: ChatMessage,
    private val project: Project,
    private val parentPanel: JPanel,
    private val gradientWidth: Int,
) : PanelWithGradientBorder(gradientWidth, chatMessage.speaker) {
  private var lastMessagePart: MessagePart? = null

  init {
    val markdownNodes: Node = markdownParser.parse(chatMessage.displayText)
    markdownNodes.accept(MessageContentCreatorFromMarkdownNodes(this, htmlRenderer))
  }

  fun updateContentWith(message: ChatMessage) {
    val markdownNodes = markdownParser.parse(message.displayText)
    val lastMarkdownNode = markdownNodes.lastChild
    if (lastMarkdownNode.isCodeBlock()) {
      val (code, language) = lastMarkdownNode.extractCodeAndLanguage()
      addOrUpdateCode(code, language)
    } else {
      val nodesAfterLastCodeBlock = markdownNodes.findNodeAfterLastCodeBlock()
      val renderedHtml = htmlRenderer.render(nodesAfterLastCodeBlock)
      addOrUpdateText(renderedHtml)
    }
  }

  fun addOrUpdateCode(code: String, language: String?) {
    val lastPart = lastMessagePart
    if (lastPart is CodeEditorPart) {
      lastPart.updateCode(project, code, language)
    } else {
      addAsNewCodeComponent(code, language)
    }
  }

  private fun addAsNewCodeComponent(code: String, info: String?) {
    val codeEditorComponent =
        CodeEditorFactory(project, parentPanel, gradientWidth).createCodeEditor(code, info)
    this.lastMessagePart = codeEditorComponent
    add(codeEditorComponent.component)
  }

  fun addOrUpdateText(text: String) {
    val lastPart = lastMessagePart
    if (lastPart is TextPart) {
      lastPart.updateText(text)
    } else {
      addAsNewTextComponent(text)
    }
  }

  private fun addAsNewTextComponent(renderedHtml: String) {
    val textPane: JEditorPane = createHtmlViewer(getInlineCodeBackgroundColor(chatMessage.speaker))
    SwingHelper.setHtml(textPane, renderedHtml, UIUtil.getLabelForeground())
    val textEditorComponent = TextPart(textPane)
    this.lastMessagePart = textEditorComponent
    add(textEditorComponent.component)
  }

  private fun getInlineCodeBackgroundColor(speaker: Speaker): Color {
    return if (speaker == Speaker.ASSISTANT) ColorUtil.darker(UIUtil.getPanelBackground(), 3)
    else ColorUtil.brighter(UIUtil.getPanelBackground(), 3)
  }

  companion object {
    private val extensions = listOf(TablesExtension.create())

    private val markdownParser = Parser.builder().extensions(extensions).build()
    private val htmlRenderer =
        HtmlRenderer.builder().softbreak("<br />").extensions(extensions).build()
  }
}
