package com.sourcegraph.cody.autocomplete.render;

import com.intellij.openapi.editor.DefaultLanguageHighlighterColors;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.impl.FontInfo;
import com.intellij.openapi.editor.markup.TextAttributes;
import com.intellij.ui.JBColor;
import com.sourcegraph.cody.autocomplete.InlayModelUtils;
import java.awt.Font;
import java.awt.FontMetrics;
import org.jetbrains.annotations.NotNull;

public class AutocompleteRenderUtils {
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
      return editor
          .getColorsScheme()
          .getAttributes(DefaultLanguageHighlighterColors.INLAY_TEXT_WITHOUT_BACKGROUND);
    } catch (Exception ignored) {
      return editor
          .getColorsScheme()
          .getAttributes(DefaultLanguageHighlighterColors.INLINE_PARAMETER_HINT);
    }
  }

  public static TextAttributes getCustomTextAttributes(
      @NotNull Editor editor, @NotNull Integer fontColor) {
    JBColor color = new JBColor(fontColor, fontColor); // set light & dark mode colors explicitly
    TextAttributes attrs = getTextAttributesForEditor(editor).clone();
    attrs.setForegroundColor(color);
    return attrs;
  }

  public static void rerenderAllAutocompleteInlays(Editor editor) {
    InlayModelUtils.getAllInlaysForEditor(editor).stream()
        .filter(inlay -> inlay.getRenderer() instanceof CodyAutocompleteElementRenderer)
        .forEach(
            inlayAutocomplete -> {
              CodyAutocompleteElementRenderer renderer =
                  (CodyAutocompleteElementRenderer) inlayAutocomplete.getRenderer();
              if (renderer instanceof CodyAutocompleteSingleLineRenderer) {
                editor
                    .getInlayModel()
                    .addInlineElement(
                        inlayAutocomplete.getOffset(),
                        new CodyAutocompleteSingleLineRenderer(
                            renderer.getText(),
                            renderer.completionItem,
                            editor,
                            renderer.getType()));
                inlayAutocomplete.dispose();
              } else if (renderer instanceof CodyAutocompleteBlockElementRenderer) {
                editor
                    .getInlayModel()
                    .addInlineElement(
                        inlayAutocomplete.getOffset(),
                        new CodyAutocompleteBlockElementRenderer(
                            renderer.getText(), renderer.completionItem, editor));
                inlayAutocomplete.dispose();
              }
            });
  }
}
