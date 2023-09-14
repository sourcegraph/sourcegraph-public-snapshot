package com.sourcegraph.cody.config.notification

import com.intellij.openapi.Disposable
import com.intellij.openapi.diagnostic.Logger
import com.intellij.openapi.project.Project
import com.intellij.util.messages.MessageBusConnection
import com.sourcegraph.find.browser.JavaToJSBridge

abstract class ChangeListener(protected val project: Project) : Disposable {
  protected val connection: MessageBusConnection = project.messageBus.connect()
  var javaToJSBridge: JavaToJSBridge? = null
  protected val logger = Logger.getInstance(ChangeListener::class.java)

  override fun dispose() {
    connection.disconnect()
  }
}
