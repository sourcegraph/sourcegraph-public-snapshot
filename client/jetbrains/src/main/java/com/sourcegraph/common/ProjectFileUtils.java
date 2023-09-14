package com.sourcegraph.common;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.roots.ProjectFileIndex;
import com.intellij.openapi.vfs.VirtualFile;
import java.io.File;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class ProjectFileUtils {

  public static @Nullable String getRelativePathToProjectRoot(
      @NotNull Project project, @NotNull VirtualFile file) {
    VirtualFile rootForFile = ProjectFileIndex.getInstance(project).getContentRootForFile(file);
    if (rootForFile != null) {
      return new File(rootForFile.getPath())
          .toURI()
          .relativize(new File(file.getPath()).toURI())
          .getPath();
    }
    return null;
  }
}
