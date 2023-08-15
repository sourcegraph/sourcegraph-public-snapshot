package com.sourcegraph.cody.autocomplete.render;

import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.Inlay;
import com.intellij.openapi.editor.markup.TextAttributes;
import com.sourcegraph.cody.vscode.InlineAutoCompleteItem;
import java.awt.*;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class CodyAutoCompleteSingleLineRenderer extends CodyAutoCompleteElementRenderer {
  public CodyAutoCompleteSingleLineRenderer(
      String text,
      @Nullable InlineAutoCompleteItem completionItem,
      Editor editor,
      AutoCompleteRendererType type) {
    super(text, completionItem, editor, type);
  }

  @Override
  public void paint(
      @NotNull Inlay inlay,
      @NotNull Graphics g,
      @NotNull Rectangle targetRegion,
      @NotNull TextAttributes textAttributes) {
    g.setFont(getFont());
    g.setColor(this.themeAttributes.getForegroundColor());
    int x = targetRegion.x;
    int y = targetRegion.y + fontYOffset();
    g.drawString(this.text, x, y);
  }

  @Override
  public String toString() {
    return "CodyAutoCompleteSingleLineRenderer{" + "text='" + text + '\'' + '}';
  }
}
