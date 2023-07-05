package com.sourcegraph.cody.autocomplete;

import com.intellij.openapi.editor.DefaultLanguageHighlighterColors;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.EditorCustomElementRenderer;
import com.intellij.openapi.editor.Inlay;
import com.intellij.openapi.editor.colors.EditorFontType;
import com.intellij.openapi.editor.impl.EditorImpl;
import com.intellij.openapi.editor.impl.FontInfo;
import com.intellij.openapi.editor.markup.TextAttributes;
import java.awt.*;
import org.jetbrains.annotations.NotNull;

/** Implements the logic to render a completion item inline in the editor. */
public class CodyAutoCompleteElementRenderer implements EditorCustomElementRenderer {
  public final String text;
  private final TextAttributes themeAttributes;
  private final Editor editor;

  public CodyAutoCompleteElementRenderer(String text, Editor editor) {
    this.text = text;
    this.themeAttributes = CodyAutoCompleteElementRenderer.getTextAttributes(editor);
    this.editor = editor;
  }

  @Override
  public int calcWidthInPixels(@NotNull Inlay inlay) {
    EditorImpl editor = (EditorImpl) inlay.getEditor();
    return editor.getFontMetrics(Font.PLAIN).stringWidth(text);
  }

  private int fontYOffset(Font font) {
    FontMetrics metrics =
        FontInfo.getFontMetrics(font, FontInfo.getFontRenderContext(editor.getContentComponent()));
    double fontBaseline =
        font.createGlyphVector(metrics.getFontRenderContext(), "Hello world!")
            .getVisualBounds()
            .getHeight();
    double linePadding = (editor.getLineHeight() - fontBaseline) / 2;
    return (int) Math.ceil(fontBaseline + linePadding);
  }

  @Override
  public void paint(
      @NotNull Inlay inlay,
      @NotNull Graphics g,
      @NotNull Rectangle targetRegion,
      @NotNull TextAttributes textAttributes) {
    Font font = this.editor.getColorsScheme().getFont(EditorFontType.PLAIN).deriveFont(Font.ITALIC);
    g.setFont(font);
    g.setColor(this.themeAttributes.getForegroundColor());
    int x = targetRegion.x;
    int y = targetRegion.y + fontYOffset(font);
    g.drawString(this.text, x, y);
  }

  @Override
  public String toString() {
    return "CodyCompletionElementRenderer{" + "text='" + text + '\'' + '}';
  }

  private static TextAttributes getTextAttributes(Editor editor) {
    try {
      //noinspection MissingRecentApi
      return editor
          .getColorsScheme()
          .getAttributes(DefaultLanguageHighlighterColors.INLAY_TEXT_WITHOUT_BACKGROUND);
    } catch (Exception ignored) {
      return editor
          .getColorsScheme()
          .getAttributes(DefaultLanguageHighlighterColors.INLINE_PARAMETER_HINT);
    }
  }
}
