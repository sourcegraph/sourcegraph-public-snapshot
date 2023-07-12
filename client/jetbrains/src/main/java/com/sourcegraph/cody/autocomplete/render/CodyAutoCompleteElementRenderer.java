package com.sourcegraph.cody.autocomplete.render;

import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.EditorCustomElementRenderer;
import com.intellij.openapi.editor.Inlay;
import com.intellij.openapi.editor.colors.EditorFontType;
import com.intellij.openapi.editor.impl.EditorImpl;
import com.intellij.openapi.editor.markup.TextAttributes;
import java.awt.*;
import java.util.Optional;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public abstract class CodyAutoCompleteElementRenderer implements EditorCustomElementRenderer {
  @NotNull public final String text;
  @NotNull protected final TextAttributes themeAttributes;
  @NotNull protected final Editor editor;
  @Nullable protected final AutoCompleteRendererType type;

  public CodyAutoCompleteElementRenderer(
      @NotNull String text, @NotNull Editor editor, @Nullable AutoCompleteRendererType type) {
    this.text = text;
    this.themeAttributes = AutoCompleteRenderUtils.getTextAttributesForEditor(editor);
    this.editor = editor;
    this.type = type;
  }

  @Override
  public int calcWidthInPixels(@NotNull Inlay inlay) {
    EditorImpl editor = (EditorImpl) inlay.getEditor();
    return editor.getFontMetrics(Font.PLAIN).stringWidth(text);
  }

  protected @NotNull Font getFont() {
    Font editorFont = this.editor.getColorsScheme().getFont(EditorFontType.PLAIN);
    return Optional.ofNullable(editorFont.deriveFont(Font.ITALIC)).orElse(editorFont);
  }

  protected int fontYOffset() {
    return (int) AutoCompleteRenderUtils.fontYOffset(getFont(), this.editor);
  }

  @Override
  public String toString() {
    return "CodyCompletionElementRenderer{" + "text='" + text + '\'' + '}';
  }

  public @NotNull String getText() {
    return this.text;
  }

  public @Nullable AutoCompleteRendererType getType() {
    return this.type;
  }
}
