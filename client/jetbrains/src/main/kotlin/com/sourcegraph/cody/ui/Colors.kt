package com.sourcegraph.cody.ui

import com.intellij.ui.JBColor
import com.intellij.util.ui.JBUI
import java.awt.Color

object Colors {

  @JvmField val CYAN: Color = JBColor(Color(0, 203, 236), Color(0, 203, 236))

  @JvmField val PURPLE: Color = JBColor(Color(161, 18, 255), Color(161, 18, 255))

  @JvmField val ORANGE: Color = JBColor(Color(255, 85, 67), Color(255, 85, 67))

  @JvmField
  val SECONDARY_LINK_COLOR: Color =
      JBColor(JBUI.CurrentTheme.Link.Foreground.ENABLED, Color(129, 133, 148))
}
