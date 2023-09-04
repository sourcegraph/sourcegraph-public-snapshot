package com.sourcegraph.cody.api

import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.ToggleAction
import com.intellij.openapi.actionSystem.UpdateInBackground
import com.intellij.openapi.components.Service
import com.intellij.openapi.components.service
import com.intellij.openapi.project.DumbAware
import com.intellij.openapi.util.NlsSafe

@Service
class SourcegraphRequestExecutorBreaker {

  @Volatile var isRequestsShouldFail = false

  class Action : ToggleAction(actionText), DumbAware, UpdateInBackground {
    override fun isSelected(e: AnActionEvent) =
        service<SourcegraphRequestExecutorBreaker>().isRequestsShouldFail

    override fun setSelected(e: AnActionEvent, state: Boolean) {
      service<SourcegraphRequestExecutorBreaker>().isRequestsShouldFail = state
    }

    companion object {
      @NlsSafe private val actionText = "Break Sourcegraph API Requests"
    }
  }
}
