package com.sourcegraph.cody.ui

import com.intellij.ui.AnimatedIcon
import com.intellij.ui.components.fields.ExtendableTextField
import java.awt.Graphics
import javax.swing.JComponent
import javax.swing.JLayer
import javax.swing.plaf.LayerUI

class LoadingLayerUI : LayerUI<ExtendableTextField>() {

  private var isLoading = false
  private val icon: AnimatedIcon = AnimatedIcon.Default.INSTANCE

  fun startLoading() {
    isLoading = true
  }

  fun stopLoading() {
    isLoading = false
  }

  override fun paint(g: Graphics, c: JComponent) {
    super.paint(g, c)
    if (isLoading && c is JLayer<*>) {
      val x = c.getWidth() - icon.iconWidth - 5
      val y = (c.getHeight() - icon.iconHeight) / 2
      icon.paintIcon(c, g, x, y)
    }
  }
}
