package com.sourcegraph.cody.ui

import com.intellij.ui.components.JBScrollPane
import java.awt.event.MouseAdapter
import java.awt.event.MouseEvent
import java.util.concurrent.atomic.AtomicBoolean
import java.util.concurrent.atomic.AtomicInteger
import javax.swing.BorderFactory
import javax.swing.JPanel
import javax.swing.JScrollBar

class ChatScrollPane(messagesPanel: JPanel?) :
    JBScrollPane(messagesPanel, VERTICAL_SCROLLBAR_AS_NEEDED, HORIZONTAL_SCROLLBAR_NEVER) {
  private val wasMouseWheelScrolled = AtomicBoolean(false)
  private val wasScrollBarDragged = AtomicBoolean(false)
  private val lastManualValue = AtomicInteger(-1)

  private fun refreshLastManualValue(e: MouseEvent) =
      (e.source as? JScrollBar)?.model?.let { lastManualValue.set(it.value) }

  init {
    setBorder(BorderFactory.createEmptyBorder())
    // this hack allows us to guess if an adjustment event isn't caused by the user mouse wheel
    // scrolling and drag scrolling;
    // we need to do that, as AdjustmentEvent doesn't account for the mouse wheel/drag scroll,
    // and we don't want to rob the user of those
    addMouseWheelListener {
      wasMouseWheelScrolled.set(true)
      refreshLastManualValue(it)
    }
    // Scroll all the way down after each message
    val chatPanelVerticalScrollBar = getVerticalScrollBar()
    chatPanelVerticalScrollBar.addMouseListener(
        object : MouseAdapter() {
          override fun mousePressed(e: MouseEvent) {
            wasScrollBarDragged.set(true)
            refreshLastManualValue(e)
          }

          override fun mouseReleased(e: MouseEvent) {
            wasScrollBarDragged.set(false)
            refreshLastManualValue(e)
          }
        })
    chatPanelVerticalScrollBar.addAdjustmentListener {
      (it.source as? JScrollBar)?.model?.let { brm ->
        if (!brm.valueIsAdjusting &&
            !wasMouseWheelScrolled.getAndSet(false) &&
            !wasScrollBarDragged.getAndSet(false) &&
            lastManualValue.get() < brm.maximum &&
            brm.value + brm.extent != brm.maximum) {
          if (lastManualValue.get() > -1 && lastManualValue.get() < brm.maximum)
              brm.value = lastManualValue.get()
          else {
            brm.value = brm.maximum
            lastManualValue.set(-1)
          }
        }
      }
    }
  }
}
