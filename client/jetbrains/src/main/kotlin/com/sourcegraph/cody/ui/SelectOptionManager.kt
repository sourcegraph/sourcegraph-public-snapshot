package com.sourcegraph.cody.ui

import com.intellij.ide.IdeEventQueue
import com.intellij.ide.actions.BigPopupUI
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.popup.JBPopup
import com.intellij.openapi.ui.popup.JBPopupFactory
import com.intellij.openapi.util.Disposer
import com.intellij.openapi.util.WindowStateService
import com.intellij.util.ui.JBInsets
import java.awt.Dimension
import java.util.function.Consumer

class SelectOptionManager(private val myProject: Project) {
  private var myBalloon: JBPopup? = null
  private var mySelectOptionPopupUI: SelectOptionPopupUI? = null
  private var myBalloonFullSize: Dimension? = null

  fun show(project: Project, options: List<String>, runOnSelect: Consumer<String>) {
    IdeEventQueue.getInstance().popupManager.closeAllPopups(false)
    val createdSelectPopupUi = createView(project, options, runOnSelect)
    mySelectOptionPopupUI = createdSelectPopupUi
    val createdBalloon =
        JBPopupFactory.getInstance()
            .createComponentPopupBuilder(
                createdSelectPopupUi,
                createdSelectPopupUi.searchField,
            )
            .setProject(myProject)
            .setModalContext(false)
            .setCancelOnClickOutside(true)
            .setRequestFocus(true)
            .setCancelKeyEnabled(false)
            .addUserData("SIMPLE_WINDOW") // NON-NLS
            .setResizable(true)
            .setMovable(true)
            .setDimensionServiceKey(myProject, LOCATION_SETTINGS_KEY, true)
            .setLocateWithinScreenBounds(false)
            .createPopup()
    myBalloon = createdBalloon
    Disposer.register(createdBalloon, createdSelectPopupUi)
    myBalloon?.let { Disposer.register(it, createdSelectPopupUi) }
    Disposer.register(createdBalloon, createdSelectPopupUi)
    Disposer.register(project, createdBalloon)
    val size = createdSelectPopupUi.minimumSize
    JBInsets.addTo(size, createdBalloon.content.insets)
    createdBalloon.setMinimumSize(size)
    Disposer.register(
        createdBalloon,
    ) {
      saveSize()
      mySelectOptionPopupUI = null
      myBalloon = null
      myBalloonFullSize = null
    }
    if (BigPopupUI.ViewType.SHORT == createdSelectPopupUi.viewType) {
      myBalloonFullSize = WindowStateService.getInstance(myProject).getSize(LOCATION_SETTINGS_KEY)
      val prefSize = createdSelectPopupUi.preferredSize
      createdBalloon.size = prefSize
    }
    calcPositionAndShow(project, createdBalloon)
  }

  private fun calcPositionAndShow(project: Project, balloon: JBPopup) {
    val savedLocation = WindowStateService.getInstance(myProject).getLocation(LOCATION_SETTINGS_KEY)
    balloon.showCenteredInCurrentWindow(project)

    // for first show and short mode popup should be shifted to the top screen half
    if (savedLocation == null && mySelectOptionPopupUI!!.viewType == BigPopupUI.ViewType.SHORT) {
      val location = balloon.locationOnScreen
      location.y /= 2
      balloon.setLocation(location)
    }
  }

  val isShown: Boolean
    get() = mySelectOptionPopupUI != null && myBalloon != null && !myBalloon!!.isDisposed

  private fun createView(
      project: Project,
      options: List<String>,
      runOnSelect: Consumer<String>,
  ): SelectOptionPopupUI {
    val view = SelectOptionPopupUI(project, options, runOnSelect)
    view.setSearchFinishedHandler {
      if (isShown) {
        myBalloon?.cancel()
      }
    }
    view.addViewTypeListener { viewType: BigPopupUI.ViewType ->
      if (!isShown) {
        return@addViewTypeListener
      }
      ApplicationManager.getApplication().invokeLater {
        val minSize = view.minimumSize
        myBalloon?.let { balloon ->
          JBInsets.addTo(minSize, balloon.content.insets)
          balloon.setMinimumSize(minSize)
          if (viewType == BigPopupUI.ViewType.SHORT) {
            myBalloonFullSize = balloon.size
            JBInsets.removeFrom(balloon.size, balloon.content.insets)
            myBalloon?.pack(false, true)
          } else {
            if (myBalloonFullSize == null) {
              myBalloonFullSize = view.preferredSize
              JBInsets.addTo(view.preferredSize, balloon.content.insets)
            }
            myBalloonFullSize?.let { balloonFullSize ->
              val height = Integer.max(balloonFullSize.height, minSize.height)
              val width = Integer.max(balloonFullSize.width, minSize.width)
              myBalloonFullSize = Dimension(width, height)
              balloon.size = balloonFullSize
            }
          }
        }
      }
    }
    return view
  }

  private fun saveSize() {
    if (BigPopupUI.ViewType.SHORT == mySelectOptionPopupUI?.viewType) {
      WindowStateService.getInstance(myProject).putSize(LOCATION_SETTINGS_KEY, myBalloonFullSize)
    }
  }

  companion object {
    private const val LOCATION_SETTINGS_KEY = "cody.select.option.popup"

    @JvmStatic
    fun getInstance(project: Project): SelectOptionManager {
      return project.getService(SelectOptionManager::class.java)
    }
  }
}
