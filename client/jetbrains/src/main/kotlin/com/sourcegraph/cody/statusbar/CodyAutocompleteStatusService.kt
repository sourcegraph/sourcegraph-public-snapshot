package com.sourcegraph.cody.statusbar

import com.intellij.openapi.Disposable
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.project.Project
import com.intellij.openapi.project.ProjectManager
import com.sourcegraph.cody.config.CodyAuthenticationManager
import com.sourcegraph.config.CodySignedOutNotification
import com.sourcegraph.config.ConfigUtil
import javax.annotation.concurrent.GuardedBy

class CodyAutocompleteStatusService : CodyAutocompleteStatusListener, Disposable {

  @GuardedBy("this") private var status: CodyAutocompleteStatus = CodyAutocompleteStatus.CodyUninit

  init {
    ApplicationManager.getApplication()
        .messageBus
        .connect(this)
        .subscribe(CodyAutocompleteStatusListener.TOPIC, this)
  }

  override fun onCodyAutocompleteStatus(codyAutocompleteStatus: CodyAutocompleteStatus) {
    val notify =
        synchronized(this) {
          val oldStatus = status
          status = codyAutocompleteStatus
          return@synchronized oldStatus != codyAutocompleteStatus
        }
    if (notify) {
      updateCodyStatusBarIcons()
    }
  }

  override fun onCodyAutocompleteStatusReset(project: Project) {
    val notify =
        synchronized(this) {
          val oldStatus = status
          ApplicationManager.getApplication()
          status =
              if (!ConfigUtil.isCodyEnabled()) {
                CodyAutocompleteStatus.CodyDisabled
              } else if (!ConfigUtil.isCodyAutocompleteEnabled()) {
                CodyAutocompleteStatus.AutocompleteDisabled
              } else if (CodyAuthenticationManager.getInstance().getActiveAccount(project) ==
                  null) {
                CodySignedOutNotification.show(project)
                CodyAutocompleteStatus.CodyNotSignedIn
              } else {
                CodySignedOutNotification.expire()
                CodyAutocompleteStatus.Ready
              }
          return@synchronized oldStatus != status
        }
    if (notify) {
      updateCodyStatusBarIcons()
    }
  }

  private fun updateCodyStatusBarIcons() {
    val action = Runnable {
      val openProjects = ProjectManager.getInstance().openProjects
      openProjects.forEach { project ->
        project.takeIf { !it.isDisposed }?.let { CodyStatusBarWidget.update(it) }
      }
    }
    val application = ApplicationManager.getApplication()
    if (application.isDispatchThread) {
      action.run()
    } else {
      application.invokeLater(action)
    }
  }

  private fun getStatus(): CodyAutocompleteStatus {
    synchronized(this) {
      return status
    }
  }

  override fun dispose() {}

  companion object {

    fun getCurrentStatus(): CodyAutocompleteStatus {
      return ApplicationManager.getApplication()
          .getService(CodyAutocompleteStatusService::class.java)
          .getStatus()
    }

    @JvmStatic
    fun notifyApplication(status: CodyAutocompleteStatus) {
      ApplicationManager.getApplication()
          .messageBus
          .syncPublisher(CodyAutocompleteStatusListener.TOPIC)
          .onCodyAutocompleteStatus(status)
    }

    @JvmStatic
    fun resetApplication(project: Project) {
      ApplicationManager.getApplication()
          .messageBus
          .syncPublisher(CodyAutocompleteStatusListener.TOPIC)
          .onCodyAutocompleteStatusReset(project)
    }
  }
}
