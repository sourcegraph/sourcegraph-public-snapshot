package com.sourcegraph.cody.chat

import com.intellij.ui.components.JBTextArea
import java.util.*

class CodyChatMessageHistory(
    private val capacity: Int
) {
    private var upperStack: Stack<String> = Stack<String>()
    private var lowerStack: Stack<String> = Stack<String>()
    private var currentValue: String = ""

    fun popUpperMessage(promptInput: JBTextArea) {
        if (upperStack.isNotEmpty()) {
            val pop = upperStack.pop()
            lowerStack.push(currentValue)
            promptInput.text = pop
            currentValue = pop
        }
    }

    fun popLowerMessage(promptInput: JBTextArea) {
        if (lowerStack.isNotEmpty()) {
            val pop = lowerStack.pop()
            upperStack.push(currentValue)
            promptInput.text = pop
            currentValue = pop
        }
    }

    /**
     * When new message is sent it is pushing all messages
     *  from lower stack to upper stack and at the end pushes new message
     */
    fun messageSent(promptInput: JBTextArea) {
        if (currentValue.isNotEmpty()) {
            upperStack.push(currentValue)
        }
        while (lowerStack.isNotEmpty()) {
            val pop: String = lowerStack.pop()
            if (pop.isNotEmpty()) upperStack.push(pop)
        }
        currentValue = ""
        upperStack.push(promptInput.text)
        if (upperStack.size > capacity) {
            upperStack.removeFirst()
        }
    }

    fun clearHistory() {
        upperStack.clear()
        lowerStack.clear()
        currentValue = ""
    }
}