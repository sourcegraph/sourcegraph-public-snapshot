package com.sourcegraph.cody.chat

import com.intellij.openapi.ui.VerticalFlowLayout
import com.intellij.ui.ColorUtil
import com.intellij.util.ui.UIUtil
import com.sourcegraph.cody.api.Speaker
import com.sourcegraph.cody.ui.Colors
import java.awt.GradientPaint
import java.awt.Graphics
import java.awt.Graphics2D
import javax.swing.BorderFactory
import javax.swing.JPanel
import javax.swing.border.Border

open class PanelWithGradientBorder(private val gradientWidth: Int, speaker: Speaker) : JPanel() {

  private val isHuman: Boolean

  init {
    val emptyBorder = BorderFactory.createEmptyBorder(0, 0, 0, 0)
    val background = UIUtil.getPanelBackground()
    val topBorder: Border =
        BorderFactory.createMatteBorder(1, 0, 0, 0, ColorUtil.brighter(background, 2))
    val bottomBorder: Border =
        BorderFactory.createMatteBorder(0, 0, 1, 0, ColorUtil.brighter(background, 3))
    val topAndBottomBorder: Border = BorderFactory.createCompoundBorder(topBorder, bottomBorder)
    isHuman = speaker == Speaker.HUMAN
    this.border = if (isHuman) emptyBorder else topAndBottomBorder
    this.background = if (isHuman) ColorUtil.darker(background, 2) else background
    this.layout = VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 0, true, false)
  }

  override fun paintComponent(g: Graphics) {
    super.paintComponent(g)
    paintLeftBorderGradient(g)
  }

  private fun paintLeftBorderGradient(g: Graphics) {
    if (isHuman) return
    val halfOfHeight = height / 2
    val firstPartGradient =
        GradientPaint(0f, 0f, Colors.PURPLE, 0f, halfOfHeight.toFloat(), Colors.ORANGE)
    val secondPartGradient =
        GradientPaint(0f, halfOfHeight.toFloat(), Colors.ORANGE, 0f, height.toFloat(), Colors.CYAN)
    val g2d = g as Graphics2D
    g2d.paint = firstPartGradient
    g2d.fillRect(0, 0, gradientWidth, halfOfHeight)
    g2d.paint = secondPartGradient
    g2d.fillRect(0, halfOfHeight, gradientWidth, height)
  }
}
