package com.sourcegraph.cody.chat;

import com.intellij.ide.highlighter.HighlighterFactory;
import com.intellij.lang.Language;
import com.intellij.openapi.editor.Document;
import com.intellij.openapi.editor.EditorFactory;
import com.intellij.openapi.editor.EditorSettings;
import com.intellij.openapi.editor.colors.EditorColorsManager;
import com.intellij.openapi.editor.ex.EditorEx;
import com.intellij.openapi.editor.highlighter.EditorHighlighter;
import com.intellij.openapi.fileTypes.FileType;
import com.intellij.openapi.fileTypes.FileTypeManager;
import com.intellij.openapi.fileTypes.PlainTextFileType;
import com.intellij.ui.ColorUtil;
import com.intellij.util.concurrency.annotations.RequiresEdt;
import com.intellij.util.ui.JBInsets;
import com.intellij.util.ui.SwingHelper;
import com.intellij.util.ui.UIUtil;
import com.sourcegraph.cody.api.Speaker;
import com.sourcegraph.cody.ui.HtmlViewer;
import java.awt.*;
import java.util.List;
import java.util.Optional;
import javax.swing.*;
import javax.swing.border.EmptyBorder;
import org.commonmark.ext.gfm.tables.TablesExtension;
import org.commonmark.node.*;
import org.commonmark.node.Image;
import org.commonmark.renderer.html.HtmlRenderer;
import org.jetbrains.annotations.Nls;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/**
 * This class is to be used with a Markdown document like this: Node document =
 * parser.parse(message.getDisplayText()); document.accept(messageContentCreator); It converts a
 * single chat message to a JPanel and other Swing components inside it.
 */
public class MessageContentCreatorFromMarkdownNodes extends AbstractVisitor {
  private final HtmlRenderer htmlRenderer =
      HtmlRenderer.builder().extensions(List.of(TablesExtension.create())).build();
  private final JPanel messagePanel;
  private final Speaker speaker;
  private final int gradientWidth;
  private StringBuilder htmlContent = new StringBuilder();
  private int textPaneIndex = 0;
  private JEditorPane textPane;

  public MessageContentCreatorFromMarkdownNodes(
      @NotNull JPanel messagePanel, @NotNull Speaker speaker, int gradientWidth) {
    this.messagePanel = messagePanel;
    this.speaker = speaker;
    this.gradientWidth = gradientWidth;
    createNewEmptyTextPane();
  }

  private void createNewEmptyTextPane() {
    textPane = HtmlViewer.createHtmlViewer(getInlineCodeBackgroundColor(this.speaker));
    messagePanel.add(textPane, textPaneIndex++);
  }

  @NotNull
  private static Color getInlineCodeBackgroundColor(@NotNull Speaker speaker) {
    return speaker == Speaker.ASSISTANT
        ? ColorUtil.darker(UIUtil.getPanelBackground(), 3)
        : ColorUtil.brighter(UIUtil.getPanelBackground(), 3);
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
    /* Create document */
    Document codeDocument = EditorFactory.getInstance().createDocument(codeContent);

    /* Create editor */
    EditorEx editor = (EditorEx) EditorFactory.getInstance().createViewer(codeDocument);
    setHighlighting(editor, languageName);
    fillEditorSettings(editor.getSettings());
    editor.setVerticalScrollbarVisible(false);
    editor.getGutterComponentEx().setPaintBackground(false);

    /* Create editor panel and add it to a parent */
    JPanel editorPanel = new JPanel(new BorderLayout());
    editorPanel.setBorder(new EmptyBorder(JBInsets.create(new Insets(0, gradientWidth, 0, 0))));
    editorPanel.add(editor.getComponent(), BorderLayout.CENTER);
    editorPanel.setOpaque(false);
    messagePanel.add(editorPanel, BorderLayout.CENTER, textPaneIndex++);

    /* Start a new block */
    htmlContent = new StringBuilder();
    createNewEmptyTextPane();
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
    textPane.setText(buildHtmlContent(htmlContent.toString()));
  }

  @NotNull
  private static @Nls String buildHtmlContent(@NotNull String bodyContent) {
    return SwingHelper.buildHtml(
        UIUtil.getCssFontDeclaration(
            UIUtil.getLabelFont(), UIUtil.getActiveTextColor(), null, null),
        bodyContent);
  }

  private static void setHighlighting(@NotNull EditorEx editor, @Nullable String languageName) {
    FileType fileType =
        Language.getRegisteredLanguages().stream()
            .filter(it -> it.getDisplayName().equalsIgnoreCase(languageName))
            .findFirst()
            .flatMap(
                it -> Optional.ofNullable(FileTypeManager.getInstance().findFileTypeByLanguage(it)))
            .orElse(PlainTextFileType.INSTANCE);

    EditorHighlighter editorHighlighter =
        HighlighterFactory.createHighlighter(
            fileType, EditorColorsManager.getInstance().getSchemeForCurrentUITheme(), null);
    editor.setHighlighter(editorHighlighter);
  }

  private static void fillEditorSettings(@NotNull final EditorSettings editorSettings) {
    editorSettings.setAdditionalColumnsCount(0);
    editorSettings.setAdditionalLinesCount(0);
    editorSettings.setGutterIconsShown(false);
    editorSettings.setWhitespacesShown(false);
    editorSettings.setLineMarkerAreaShown(false);
    editorSettings.setIndentGuidesShown(false);
    editorSettings.setLineNumbersShown(false);
    editorSettings.setUseSoftWraps(false);
    editorSettings.setCaretRowShown(false);
  }
}
