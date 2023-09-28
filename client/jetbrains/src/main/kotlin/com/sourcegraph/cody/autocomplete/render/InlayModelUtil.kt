package com.sourcegraph.cody.autocomplete.render

import com.intellij.openapi.editor.Caret
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.editor.Inlay
import com.intellij.openapi.editor.InlayModel

object InlayModelUtil {
  @JvmStatic
  fun getAllInlays(inlayModel: InlayModel, startOffset: Int, endOffset: Int): List<Inlay<*>> {
    // can't use inlineModel.getInlineElementAt(caret.getVisualPosition()) here, as it
    // requires a write EDT thread;
    // we work around it by just looking at a range (potentially containing a single point)
    return listOf(
            inlayModel.getInlineElementsInRange(
                startOffset, endOffset, CodyAutocompleteElementRenderer::class.java),
            inlayModel.getBlockElementsInRange(
                startOffset, endOffset, CodyAutocompleteElementRenderer::class.java),
            inlayModel.getAfterLineEndElementsInRange(
                startOffset, endOffset, CodyAutocompleteElementRenderer::class.java))
        .flatten()
  }

  @JvmStatic
  fun getAllInlaysForEditor(editor: Editor): List<Inlay<*>> {
    return getAllInlays(editor.inlayModel, 0, editor.document.textLength)
  }

  @JvmStatic
  fun getAllInlaysForCaret(caret: Caret): List<Inlay<*>> {
    return getAllInlays(caret.editor.inlayModel, caret.offset, caret.offset)
  }
}
