package com.sourcegraph.cody.context;

import com.intellij.openapi.fileEditor.FileEditorManagerEvent;
import com.intellij.openapi.fileEditor.FileEditorManagerListener;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.sourcegraph.common.ProjectFileUtils;
import java.util.Optional;
import org.jetbrains.annotations.NotNull;

public class CurrentlyOpenFileListener implements FileEditorManagerListener {

  private final @NotNull Project project;
  private final @NotNull EmbeddingStatusView embeddingStatusView;

  public CurrentlyOpenFileListener(
      @NotNull Project project, @NotNull EmbeddingStatusView embeddingStatusView) {
    this.project = project;
    this.embeddingStatusView = embeddingStatusView;
  }

  @Override
  public void selectionChanged(@NotNull FileEditorManagerEvent event) {
    VirtualFile newFile = event.getNewFile();
    String openedFileName = Optional.ofNullable(newFile).map(VirtualFile::getName).orElse("");
    String relativeFilePath = null;
    if (newFile != null) {
      relativeFilePath = ProjectFileUtils.getRelativePathToProjectRoot(project, newFile);
    }
    embeddingStatusView.setOpenedFileName(openedFileName, relativeFilePath);
  }
}
