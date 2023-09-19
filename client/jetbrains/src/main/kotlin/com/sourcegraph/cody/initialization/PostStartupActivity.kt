package com.sourcegraph.cody.initialization

import com.intellij.openapi.project.Project
import com.intellij.openapi.startup.StartupActivity
import com.sourcegraph.cody.auth.SelectOneOfTheAccountsAsActive
import com.sourcegraph.cody.config.SettingsMigration
import com.sourcegraph.config.CodyAuthNotificationActivity
import com.sourcegraph.telemetry.TelemetryInitializerActivity

class PostStartupActivity : StartupActivity.DumbAware {
  override fun runActivity(project: Project) {
    TelemetryInitializerActivity().runActivity(project)
    SettingsMigration().runActivity(project)
    SelectOneOfTheAccountsAsActive().runActivity(project)
    CodyAuthNotificationActivity().runActivity(project)
  }
}
