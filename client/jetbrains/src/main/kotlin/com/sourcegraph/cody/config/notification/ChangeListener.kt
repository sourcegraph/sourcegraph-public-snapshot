package com.sourcegraph.cody.config.notification

import com.intellij.openapi.Disposable
import com.intellij.openapi.diagnostic.Logger
import com.intellij.openapi.project.Project
import com.intellij.util.messages.MessageBus
import com.intellij.util.messages.MessageBusConnection
import com.sourcegraph.find.browser.JavaToJSBridge

abstract class ChangeListener(protected val project: Project) : Disposable {
    protected var connection: MessageBusConnection? = null
    var javaToJSBridge: JavaToJSBridge? = null
    protected val logger = Logger.getInstance(ChangeListener::class.java)
    private val bus: MessageBus = project.messageBus

    init {
        connection = bus.connect()
    }

    override fun dispose() {
        connection!!.disconnect()
    }
}
