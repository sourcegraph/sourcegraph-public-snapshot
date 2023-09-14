package com.sourcegraph.cody.chat

import com.intellij.ui.components.JBTextArea
import java.util.*

class CodyChatMessageHistory(private val capacity: Int) {
  var currentValue: String = ""
  var upperStack: Stack<String> = Stack<String>()
  private var lowerStack: Stack<String> = Stack<String>()

  fun popUpperMessage(promptInput: JBTextArea) {
    resetHistoryIfPromptCleared(promptInput)
    if (upperStack.isNotEmpty()) {
      val pop = upperStack.pop()
      lowerStack.push(promptInput.text)
      promptInput.text = pop
      currentValue = pop
    }
  }

  fun popLowerMessage(promptInput: JBTextArea) {
    resetHistoryIfPromptCleared(promptInput)
    if (lowerStack.isNotEmpty()) {
      val pop = lowerStack.pop()
      upperStack.push(promptInput.text)
      promptInput.text = pop
      currentValue = pop
    }
  }

  /**
   * When new message is sent it is pushing all messages from lower stack to upper stack and at the
   * end pushes new message
   */
  fun messageSent(promptInput: JBTextArea) {
    resetHistory()
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

  private fun resetHistory() {
    if (currentValue.isNotEmpty()) {
      upperStack.push(currentValue)
    }
    while (lowerStack.isNotEmpty()) {
      val pop: String = lowerStack.pop()
      if (pop.isNotEmpty()) upperStack.push(pop)
    }
    currentValue = ""
  }

  private fun resetHistoryIfPromptCleared(promptInput: JBTextArea) {
    if (promptInput.text.isEmpty() && currentValue.isNotEmpty()) {
      resetHistory()
    }
  }
}
