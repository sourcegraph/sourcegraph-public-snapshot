package com.sourcegraph.cody.autocomplete.render

import com.intellij.openapi.editor.Editor
import com.intellij.openapi.editor.EditorCustomElementRenderer
import com.intellij.openapi.editor.Inlay
import com.intellij.openapi.editor.colors.EditorFontType
import com.intellij.openapi.editor.impl.EditorImpl
import com.intellij.openapi.editor.markup.TextAttributes
import com.sourcegraph.cody.vscode.InlineAutocompleteItem
import com.sourcegraph.config.ConfigUtil.getCustomAutocompleteColor
import com.sourcegraph.config.ConfigUtil.isCustomAutocompleteColorEnabled
import java.awt.Font
import java.util.function.Supplier

abstract class CodyAutocompleteElementRenderer(
    val text: String,
    val completionItems: List<InlineAutocompleteItem>,
    protected val editor: Editor,
    val type: AutocompleteRendererType
) : EditorCustomElementRenderer {
  protected val themeAttributes: TextAttributes

  init {
    val textAttributesFallback = Supplier {
      AutocompleteRenderUtil.getTextAttributesForEditor(editor)
    }
    themeAttributes =
        if (isCustomAutocompleteColorEnabled())
            getCustomAutocompleteColor()?.let {
              AutocompleteRenderUtil.getCustomTextAttributes(editor, it)
            }
                ?: textAttributesFallback.get()
        else textAttributesFallback.get()
  }

  override fun calcWidthInPixels(inlay: Inlay<*>): Int {
    val editor = inlay.editor as EditorImpl
    return editor.getFontMetrics(Font.PLAIN).stringWidth(text)
  }

  protected val font: Font
    get() {
      val editorFont = editor.colorsScheme.getFont(EditorFontType.PLAIN)
      return editorFont.deriveFont(Font.ITALIC) ?: editorFont
    }

  protected fun fontYOffset(): Int = AutocompleteRenderUtil.fontYOffset(font, editor).toInt()
}
