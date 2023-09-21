package com.sourcegraph.cody.statusbar

import com.intellij.openapi.Disposable
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.project.ProjectManager
import javax.annotation.concurrent.GuardedBy

class CodyAutocompleteStatusService : CodyAutocompleteStatusListener, Disposable {

  @GuardedBy("this") private var status: CodyAutocompleteStatus = CodyAutocompleteStatus.Ready

  init {
    ApplicationManager.getApplication()
        .messageBus
        .connect(this)
        .subscribe(CodyAutocompleteStatusListener.TOPIC, this)
  }

  override fun onCodyAutocompleteStatus(codyAutocompleteStatus: CodyAutocompleteStatus) {
    var notify: Boolean
    synchronized(this) {
      val oldStatus = status
      notify = oldStatus != codyAutocompleteStatus
      status = codyAutocompleteStatus
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
  }
}
