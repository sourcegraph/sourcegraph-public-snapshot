package com.sourcegraph.cody.autocomplete.render;

import com.intellij.openapi.editor.DefaultLanguageHighlighterColors;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.impl.FontInfo;
import com.intellij.openapi.editor.markup.TextAttributes;
import java.awt.*;
import org.jetbrains.annotations.NotNull;

public class AutoCompleteRenderUtils {
  public static double fontYOffset(@NotNull Font font, @NotNull Editor editor) {
    FontMetrics metrics =
        FontInfo.getFontMetrics(font, FontInfo.getFontRenderContext(editor.getContentComponent()));
    double fontBaseline =
        font.createGlyphVector(metrics.getFontRenderContext(), "Hello world!")
            .getVisualBounds()
            .getHeight();
    double linePadding = (editor.getLineHeight() - fontBaseline) / 2;
    return Math.ceil(fontBaseline + linePadding);
  }

  public static TextAttributes getTextAttributesForEditor(@NotNull Editor editor) {
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
