package com.sourcegraph.cody.ui

import com.intellij.ui.ColorUtil
import com.intellij.ui.components.JBTextArea
import java.awt.Graphics
import java.awt.Graphics2D
import java.awt.RenderingHints
import java.awt.geom.RoundRectangle2D
import javax.swing.BorderFactory

class RoundedJBTextArea(minRows: Int, private val cornerRadius: Int) : JBTextArea(minRows, 0) {
  init {
    isOpaque = false
    border = BorderFactory.createEmptyBorder(4, 4, 4, 4)
  }

  override fun paintComponent(g: Graphics) {
    val g2 = g.create() as Graphics2D
    g2.setRenderingHint(RenderingHints.KEY_ANTIALIASING, RenderingHints.VALUE_ANTIALIAS_ON)
    val roundRect =
        RoundRectangle2D.Float(
            0f,
            0f,
            (this.width - 1).toFloat(),
            (this.height - 1).toFloat(),
            cornerRadius.toFloat(),
            cornerRadius.toFloat())
    g2.color = background
    g2.fill(roundRect)
    g2.color = ColorUtil.brighter(background, 2)
    g2.draw(roundRect)
    g2.dispose()
    super.paintComponent(g)
  }
}
