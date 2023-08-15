package com.sourcegraph.cody.autocomplete;

import com.intellij.openapi.editor.Caret;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.Inlay;
import com.intellij.openapi.editor.InlayModel;
import com.sourcegraph.cody.autocomplete.render.CodyAutoCompleteElementRenderer;
import java.util.Collection;
import java.util.List;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import org.jetbrains.annotations.NotNull;

public class InlayModelUtils {
  public static List<Inlay<?>> getAllInlays(
      @NotNull InlayModel inlayModel, int startOffset, int endOffset) {
    // can't use inlineModel.getInlineElementAt(caret.getVisualPosition()) here, as it
    // requires a write EDT thread;
    // we work around it by just looking at a range (potentially containing a single point)
    return Stream.of(
            inlayModel.getInlineElementsInRange(
                startOffset, endOffset, CodyAutoCompleteElementRenderer.class),
            inlayModel.getBlockElementsInRange(
                startOffset, endOffset, CodyAutoCompleteElementRenderer.class),
            inlayModel.getAfterLineEndElementsInRange(
                startOffset, endOffset, CodyAutoCompleteElementRenderer.class))
        .flatMap(Collection::stream)
        .collect(Collectors.toList());
  }

  public static List<Inlay<?>> getAllInlaysForEditor(@NotNull Editor editor) {
    return getAllInlays(editor.getInlayModel(), 0, editor.getDocument().getTextLength());
  }

  public static List<Inlay<?>> getAllInlaysForCaret(@NotNull Caret caret) {
    return getAllInlays(caret.getEditor().getInlayModel(), caret.getOffset(), caret.getOffset());
  }
}
