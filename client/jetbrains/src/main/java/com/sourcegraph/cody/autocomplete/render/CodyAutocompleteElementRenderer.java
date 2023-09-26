package com.sourcegraph.cody.autocomplete.render;

import static com.sourcegraph.config.ConfigUtil.getCustomAutocompleteColor;
import static com.sourcegraph.config.ConfigUtil.isCustomAutocompleteColorEnabled;

import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.EditorCustomElementRenderer;
import com.intellij.openapi.editor.Inlay;
import com.intellij.openapi.editor.colors.EditorFontType;
import com.intellij.openapi.editor.impl.EditorImpl;
import com.intellij.openapi.editor.markup.TextAttributes;
import com.sourcegraph.cody.vscode.InlineAutocompleteItem;
import java.awt.*;
import java.util.Optional;
import java.util.function.Supplier;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public abstract class CodyAutocompleteElementRenderer implements EditorCustomElementRenderer {
  @NotNull public final String text;
  @Nullable public final InlineAutocompleteItem completionItem;
  @NotNull protected final TextAttributes themeAttributes;
  @NotNull protected final Editor editor;
  @Nullable protected final AutocompleteRendererType type;

  public CodyAutocompleteElementRenderer(
      @NotNull String text,
      @Nullable InlineAutocompleteItem completionItem,
      @NotNull Editor editor,
      @Nullable AutocompleteRendererType type) {
    this.text = text;
    this.completionItem = completionItem;
    Supplier<TextAttributes> textAttributesFallback =
        () -> AutocompleteRenderUtils.getTextAttributesForEditor(editor);
    this.themeAttributes =
        isCustomAutocompleteColorEnabled()
            ? Optional.ofNullable(getCustomAutocompleteColor())
                .map(c -> AutocompleteRenderUtils.getCustomTextAttributes(editor, c))
                .orElseGet(textAttributesFallback)
            : textAttributesFallback.get();
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
    return (int) AutocompleteRenderUtils.fontYOffset(getFont(), this.editor);
  }

  @Override
  public String toString() {
    return "CodyCompletionElementRenderer{" + "text='" + text + '\'' + '}';
  }

  public @NotNull String getText() {
    return this.text;
  }

  public @Nullable AutocompleteRendererType getType() {
    return this.type;
  }
}
