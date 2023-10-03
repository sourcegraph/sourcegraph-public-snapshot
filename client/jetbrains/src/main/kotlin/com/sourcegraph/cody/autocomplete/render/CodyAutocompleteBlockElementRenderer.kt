package com.sourcegraph.cody.autocomplete.render

import com.intellij.openapi.editor.Editor
import com.intellij.openapi.editor.Inlay
import com.intellij.openapi.editor.impl.EditorImpl
import com.intellij.openapi.editor.markup.TextAttributes
import com.sourcegraph.cody.vscode.InlineAutocompleteItem
import java.awt.Font
import java.awt.Graphics
import java.awt.Rectangle

/** Implements the logic to render an autocomplete item inline in the editor. */
class CodyAutocompleteBlockElementRenderer(
    text: String,
    completionItems: List<InlineAutocompleteItem>,
    editor: Editor
) : CodyAutocompleteElementRenderer(text, completionItems, editor, AutocompleteRendererType.BLOCK) {
  override fun calcWidthInPixels(inlay: Inlay<*>): Int {
    val editor = inlay.editor as EditorImpl
    val longestLine: String =
        text.lines().maxWithOrNull(Comparator.comparingInt { it.length }) ?: ""
    return editor.getFontMetrics(Font.PLAIN).stringWidth(longestLine)
  }

  override fun calcHeightInPixels(inlay: Inlay<*>): Int {
    val lineHeight = inlay.editor.lineHeight
    val linesCount = text.lines().count()
    return lineHeight * linesCount
  }

  override fun paint(
      inlay: Inlay<*>,
      g: Graphics,
      targetRegion: Rectangle,
      textAttributes: TextAttributes
  ) {
    g.font = font
    g.color = themeAttributes.foregroundColor
    val x = targetRegion.x
    val baseYOffset = fontYOffset()
    for ((i, line) in text.lines().withIndex()) {
      val y = targetRegion.y + baseYOffset + i * editor.lineHeight
      g.drawString(line, x, y)
    }
  }
}
