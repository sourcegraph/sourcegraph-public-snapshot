package com.sourcegraph.cody.chat

import com.intellij.util.ui.JBInsets
import com.sourcegraph.cody.ui.RotatingLogoAnimation
import java.awt.BorderLayout
import java.awt.Insets
import javax.swing.JPanel
import javax.swing.border.EmptyBorder

class WaitingForContentMessage : JPanel() {
  init {
    this.layout = BorderLayout()
    val comp = RotatingLogoAnimation()
    val margin =
        JBInsets.create(
            Insets(
                ChatUIConstants.TEXT_MARGIN,
                ChatUIConstants.TEXT_MARGIN,
                ChatUIConstants.TEXT_MARGIN,
                ChatUIConstants.TEXT_MARGIN))
    comp.border = EmptyBorder(margin)
    super.add(comp, BorderLayout.CENTER)
  }
}
