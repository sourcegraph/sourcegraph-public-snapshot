package com.sourcegraph.cody.ui

import com.sourcegraph.Icons
import java.awt.Graphics
import java.awt.Graphics2D
import java.awt.geom.AffineTransform
import javax.swing.JPanel
import javax.swing.Timer

class RotatingLogoAnimation : JPanel() {
  private val svgIcon = Icons.CodyLogo
  private var currentRotation = 0
  private val animationTimer: Timer =
      Timer(5) {
        currentRotation = (currentRotation + 1) % 360
        repaint()
      }

  init {
    animationTimer.start()
  }

  override fun paintComponent(g: Graphics) {
    super.paintComponent(g)

    val centerX = width / 2
    val centerY = height / 2
    val g2d = g.create() as Graphics2D
    g2d.translate(centerX, centerY)

    val angle = Math.toRadians(currentRotation.toDouble())
    val transform = AffineTransform.getRotateInstance(angle)
    g2d.transform(transform)

    val iconWidth = svgIcon.iconWidth
    val iconHeight = svgIcon.iconHeight
    val iconX = -iconWidth / 2
    val iconY = -iconHeight / 2
    svgIcon.paintIcon(this, g2d, iconX, iconY)
    g2d.dispose()
  }
}
