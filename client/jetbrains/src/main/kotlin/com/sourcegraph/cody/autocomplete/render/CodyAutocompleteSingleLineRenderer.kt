package com.sourcegraph.cody.autocomplete.render

import com.intellij.openapi.editor.Editor
import com.intellij.openapi.editor.Inlay
import com.intellij.openapi.editor.markup.TextAttributes
import com.sourcegraph.cody.vscode.InlineAutocompleteItem
import java.awt.Graphics
import java.awt.Rectangle

class CodyAutocompleteSingleLineRenderer(
    text: String,
    completionItem: InlineAutocompleteItem,
    editor: Editor,
    type: AutocompleteRendererType
) : CodyAutocompleteElementRenderer(text, completionItem, editor, type) {
  override fun paint(
      inlay: Inlay<*>,
      g: Graphics,
      targetRegion: Rectangle,
      textAttributes: TextAttributes
  ) {
    g.font = font
    g.color = themeAttributes.foregroundColor
    val x = targetRegion.x
    val y = targetRegion.y + fontYOffset()
    g.drawString(text, x, y)
  }
}
