package com.sourcegraph.cody;

import com.intellij.ide.lightEdit.LightEdit;
import com.intellij.openapi.project.Project;
import org.jetbrains.annotations.Nullable;

public class CodyCompatibility {

  public static boolean isSupportedProject(@Nullable Project project) {
    if (isProjectorEnabled()) {
      return true;
    } else {
      // Light edit is a mode when users can edit a single file with IntelliJ without loading an
      // entire project. We lean on the conservative side for now and don't support Cody for
      // LightEdit projects.
      return !LightEdit.owns(project);
    }
  }

  // Projector is a JetBrains project that runs IntelliJ on a server as a remote IDE. Users
  // interface with IntelliJ through a JavaScript client, either in the browser or an Electron.js
  // client.
  private static boolean isProjectorEnabled() {
    return "true".equals(System.getProperty("org.jetbrains.projector.server.enable"));
  }
}
