package com.sourcegraph.cody.chat;

import com.intellij.util.concurrency.annotations.RequiresEdt;
import org.commonmark.node.*;
import org.commonmark.node.Image;
import org.commonmark.renderer.html.HtmlRenderer;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/**
 * This class is to be used with a Markdown document like this: Node document =
 * parser.parse(message.getDisplayText()); document.accept(messageContentCreator); It converts a
 * single chat message to a JPanel and other Swing components inside it.
 */
public class MessageContentCreatorFromMarkdownNodes extends AbstractVisitor {
  private final HtmlRenderer htmlRenderer;
  private final MessagePanel messagePanel;
  private StringBuilder htmlContent = new StringBuilder();

  public MessageContentCreatorFromMarkdownNodes(
      @NotNull MessagePanel messagePanel, @NotNull HtmlRenderer htmlRenderer) {
    this.messagePanel = messagePanel;
    this.htmlRenderer = htmlRenderer;
  }

  @Override
  public void visit(@NotNull Paragraph paragraph) {
    addContentOfNodeAsHtml(htmlRenderer.render(paragraph));
  }

  @Override
  public void visit(@NotNull Code code) {
    addContentOfNodeAsHtml(htmlRenderer.render(code));
    super.visit(code);
  }

  @Override
  public void visit(@NotNull IndentedCodeBlock indentedCodeBlock) {
    insertCodeEditor(indentedCodeBlock.getLiteral(), "");
    super.visit(indentedCodeBlock);
  }

  @Override
  public void visit(@NotNull Text text) {
    addContentOfNodeAsHtml(htmlRenderer.render(text));
    super.visit(text);
  }

  @Override
  public void visit(@NotNull BlockQuote blockQuote) {
    addContentOfNodeAsHtml(htmlRenderer.render(blockQuote));
    super.visit(blockQuote);
  }

  @Override
  public void visit(@NotNull BulletList bulletList) {
    addContentOfNodeAsHtml(htmlRenderer.render(bulletList));
  }

  @Override
  public void visit(@NotNull OrderedList orderedList) {
    addContentOfNodeAsHtml(htmlRenderer.render(orderedList));
  }

  @Override
  public void visit(@NotNull Emphasis emphasis) {
    addContentOfNodeAsHtml(htmlRenderer.render(emphasis));
    super.visit(emphasis);
  }

  @Override
  public void visit(@NotNull FencedCodeBlock fencedCodeBlock) {
    insertCodeEditor(fencedCodeBlock.getLiteral(), fencedCodeBlock.getInfo());
    super.visit(fencedCodeBlock);
  }

  @RequiresEdt
  private void insertCodeEditor(@NotNull String codeContent, @Nullable String languageName) {
    messagePanel.addOrUpdateCode(codeContent, languageName);
    htmlContent = new StringBuilder();
  }

  @Override
  public void visit(@NotNull HardLineBreak hardLineBreak) {
    addContentOfNodeAsHtml(htmlRenderer.render(hardLineBreak));
    super.visit(hardLineBreak);
  }

  @Override
  public void visit(@NotNull Heading heading) {
    addContentOfNodeAsHtml(htmlRenderer.render(heading));
    super.visit(heading);
  }

  @Override
  public void visit(@NotNull ThematicBreak thematicBreak) {
    addContentOfNodeAsHtml(htmlRenderer.render(thematicBreak));
    super.visit(thematicBreak);
  }

  @Override
  public void visit(@NotNull HtmlInline htmlInline) {
    addContentOfNodeAsHtml(htmlRenderer.render(htmlInline));
    super.visit(htmlInline);
  }

  @Override
  public void visit(@NotNull HtmlBlock htmlBlock) {
    addContentOfNodeAsHtml(htmlRenderer.render(htmlBlock));
    super.visit(htmlBlock);
  }

  @Override
  public void visit(@NotNull Image image) {
    addContentOfNodeAsHtml(htmlRenderer.render(image));
    super.visit(image);
  }

  @Override
  public void visit(@NotNull Link link) {
    addContentOfNodeAsHtml(htmlRenderer.render(link));
  }

  @Override
  public void visit(@NotNull ListItem listItem) {
    addContentOfNodeAsHtml(htmlRenderer.render(listItem));
    super.visit(listItem);
  }

  @Override
  public void visit(@NotNull SoftLineBreak softLineBreak) {
    addContentOfNodeAsHtml(htmlRenderer.render(softLineBreak));
    super.visit(softLineBreak);
  }

  @Override
  public void visit(@NotNull StrongEmphasis strongEmphasis) {
    addContentOfNodeAsHtml(htmlRenderer.render(strongEmphasis));
    super.visit(strongEmphasis);
  }

  @Override
  public void visit(@NotNull LinkReferenceDefinition linkReferenceDefinition) {
    addContentOfNodeAsHtml(htmlRenderer.render(linkReferenceDefinition));
    super.visit(linkReferenceDefinition);
  }

  @Override
  public void visit(@NotNull CustomBlock customBlock) {
    addContentOfNodeAsHtml(htmlRenderer.render(customBlock));
  }

  private void addContentOfNodeAsHtml(@Nullable String renderedHtml) {
    htmlContent.append(renderedHtml);
    messagePanel.addOrUpdateText(htmlContent.toString());
  }
}
