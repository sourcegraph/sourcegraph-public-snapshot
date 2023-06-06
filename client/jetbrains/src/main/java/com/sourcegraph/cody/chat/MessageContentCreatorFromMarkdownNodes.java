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
import com.intellij.ui.BrowserHyperlinkListener;
import com.intellij.ui.ColorUtil;
import com.intellij.util.ui.JBInsets;
import com.intellij.util.ui.UIUtil;
import java.awt.*;
import java.util.Optional;
import javax.swing.*;
import javax.swing.border.EmptyBorder;
import org.commonmark.node.*;
import org.commonmark.node.Image;
import org.commonmark.renderer.html.HtmlRenderer;
import org.jetbrains.annotations.NotNull;

public class MessageContentCreatorFromMarkdownNodes extends AbstractVisitor {
  private static final int TEXT_MARGIN = 14;
  private final HtmlRenderer htmlRenderer = HtmlRenderer.builder().build();
  private final JPanel messagePanel;
  private final int gradientWidth;
  private StringBuilder htmlContent = new StringBuilder();
  private int textPaneIndex = 0;
  private JEditorPane textPane;

  public MessageContentCreatorFromMarkdownNodes(JPanel messagePanel, int gradientWidth) {
    this.messagePanel = messagePanel;
    this.gradientWidth = gradientWidth;
    textPane = createNewEmptyTextPane();
  }

  @NotNull
  private JEditorPane createNewEmptyTextPane() {
    JEditorPane jEditorPane = new JEditorPane();
    jEditorPane.setFont(UIUtil.getLabelFont());
    jEditorPane.setContentType("text/html");
    jEditorPane.setEditable(false);
    jEditorPane.addHyperlinkListener(BrowserHyperlinkListener.INSTANCE);
    jEditorPane.setOpaque(false);
    jEditorPane.setMargin(
        JBInsets.create(new Insets(TEXT_MARGIN, TEXT_MARGIN, TEXT_MARGIN, TEXT_MARGIN)));
    textPane = jEditorPane;
    messagePanel.add(textPane, textPaneIndex++);
    return jEditorPane;
  }

  @Override
  public void visit(Code code) {
    addContentOfNodeAsHtml(htmlRenderer.render(code));

    super.visit(code);
  }

  @Override
  public void visit(IndentedCodeBlock indentedCodeBlock) {
    insertCodeEditor(indentedCodeBlock.getLiteral(), "");
    super.visit(indentedCodeBlock);
  }

  @Override
  public void visit(Text text) {
    addContentOfNodeAsHtml(htmlRenderer.render(text));
    super.visit(text);
  }

  @Override
  public void visit(BlockQuote blockQuote) {
    addContentOfNodeAsHtml(htmlRenderer.render(blockQuote));
    super.visit(blockQuote);
  }

  @Override
  public void visit(BulletList bulletList) {
    addContentOfNodeAsHtml(htmlRenderer.render(bulletList));
  }

  @Override
  public void visit(Emphasis emphasis) {
    addContentOfNodeAsHtml(htmlRenderer.render(emphasis));
    super.visit(emphasis);
  }

  @Override
  public void visit(FencedCodeBlock fencedCodeBlock) {
    insertCodeEditor(fencedCodeBlock.getLiteral(), fencedCodeBlock.getInfo());
    super.visit(fencedCodeBlock);
  }

  private void insertCodeEditor(String codeContent, String languageName) {
    JPanel editorPanel = new JPanel(new BorderLayout());
    Document codeDocument = EditorFactory.getInstance().createDocument(codeContent);
    EditorEx editor = (EditorEx) EditorFactory.getInstance().createViewer(codeDocument);
    setHighlighting(editor, languageName);
    fillEditorSettings(editor.getSettings());
    editor.setVerticalScrollbarVisible(false);
    editor.getGutterComponentEx().setPaintBackground(false);
    editorPanel.setBorder(new EmptyBorder(JBInsets.create(new Insets(0, gradientWidth, 0, 0))));
    editorPanel.add(editor.getComponent(), BorderLayout.CENTER);
    editorPanel.setOpaque(false);
    messagePanel.add(editorPanel, BorderLayout.CENTER, textPaneIndex++);
    htmlContent = new StringBuilder();
    textPane = createNewEmptyTextPane();
  }

  @Override
  public void visit(HardLineBreak hardLineBreak) {
    addContentOfNodeAsHtml(htmlRenderer.render(hardLineBreak));
    super.visit(hardLineBreak);
  }

  @Override
  public void visit(Heading heading) {
    addContentOfNodeAsHtml(htmlRenderer.render(heading));
    super.visit(heading);
  }

  @Override
  public void visit(ThematicBreak thematicBreak) {
    addContentOfNodeAsHtml(htmlRenderer.render(thematicBreak));
    super.visit(thematicBreak);
  }

  @Override
  public void visit(HtmlInline htmlInline) {
    addContentOfNodeAsHtml(htmlRenderer.render(htmlInline));
    super.visit(htmlInline);
  }

  @Override
  public void visit(HtmlBlock htmlBlock) {
    addContentOfNodeAsHtml(htmlRenderer.render(htmlBlock));
    super.visit(htmlBlock);
  }

  @Override
  public void visit(Image image) {
    String html = htmlRenderer.render(image);
    htmlContent.append(html);
    textPane.setText(wrapWithHtmlTag(htmlContent.toString()));
    super.visit(image);
  }

  @Override
  public void visit(Link link) {
    addContentOfNodeAsHtml(htmlRenderer.render(link));
  }

  @Override
  public void visit(ListItem listItem) {
    addContentOfNodeAsHtml(htmlRenderer.render(listItem));
    super.visit(listItem);
  }

  @Override
  public void visit(OrderedList orderedList) {
    addContentOfNodeAsHtml(htmlRenderer.render(orderedList));
    super.visit(orderedList);
  }

  @Override
  public void visit(SoftLineBreak softLineBreak) {
    addContentOfNodeAsHtml(htmlRenderer.render(softLineBreak));
    super.visit(softLineBreak);
  }

  @Override
  public void visit(StrongEmphasis strongEmphasis) {
    addContentOfNodeAsHtml(htmlRenderer.render(strongEmphasis));
    super.visit(strongEmphasis);
  }

  @Override
  public void visit(LinkReferenceDefinition linkReferenceDefinition) {
    addContentOfNodeAsHtml(htmlRenderer.render(linkReferenceDefinition));
    super.visit(linkReferenceDefinition);
  }

  @Override
  public void visit(CustomBlock customBlock) {
    addContentOfNodeAsHtml(htmlRenderer.render(customBlock));
    super.visit(customBlock);
  }

  private void addContentOfNodeAsHtml(String renderedHtml) {
    htmlContent.append(renderedHtml);
    textPane.setText(wrapWithHtmlTag(htmlContent.toString()));
  }

  private @NotNull String wrapWithHtmlTag(@NotNull String htmlContent) {
    String labelTextColor = ColorUtil.toHex(UIUtil.getLabelForeground());
    String linkColor = ColorUtil.toHex(UIUtil.getHeaderActiveColor());

    // Build HTML
    return "<html data-gramm=\"false\"><head><style>"
        + "body{ color:"
        + labelTextColor
        + "; },"
        + " p { margin:0; },"
        + " a { color:"
        + linkColor
        + "; }"
        + "</style></head><body>"
        + htmlContent
        + "</body></html>";
  }

  private static void setHighlighting(EditorEx editor, String languageName) {
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

  private static void fillEditorSettings(final EditorSettings editorSettings) {
    editorSettings.setAdditionalColumnsCount(0);
    editorSettings.setAdditionalLinesCount(0);
    editorSettings.setGutterIconsShown(false);
    editorSettings.setWhitespacesShown(false);
    editorSettings.setLineMarkerAreaShown(false);
    editorSettings.setIndentGuidesShown(false);
    editorSettings.setLineNumbersShown(false);
    editorSettings.setUseSoftWraps(false);
  }
}
