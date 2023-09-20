package com.sourcegraph.cody.ui

import com.intellij.openapi.actionSystem.AnAction
import com.intellij.ui.components.AnActionLink
import java.awt.Graphics
import java.awt.Rectangle
import org.jetbrains.annotations.Nls

class UnderlinedActionLink(@Nls text: String, anAction: AnAction) : AnActionLink(text, anAction) {
  init {
    foreground = Colors.SECONDARY_LINK_COLOR
  }

  override fun paint(g: Graphics) {
    super.paint(g)
    val bounds: Rectangle = g.clipBounds
    g.color = Colors.SECONDARY_LINK_COLOR
    g.drawLine(0, bounds.height, getFontMetrics(font).stringWidth(text), bounds.height)
  }
}
