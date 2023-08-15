package com.sourcegraph.cody.recipes;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vcs.FilePath;
import com.intellij.openapi.vcs.LocalFilePath;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.vcs.log.impl.VcsProjectLog;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.chat.ChatMessage;
import com.sourcegraph.cody.ui.SelectOptionManager;
import com.sourcegraph.cody.vcs.Last5ItemsFromCurrentFileFilterOption;
import com.sourcegraph.cody.vcs.VcsCommitsMetadataLoader;
import com.sourcegraph.cody.vcs.VcsFilter;
import com.sourcegraph.cody.vcs.VcsLogFilterOptionsRegistry;
import com.sourcegraph.telemetry.GraphQlLogger;
import java.util.Optional;
import java.util.function.Supplier;
import org.apache.commons.lang.ArrayUtils;
import org.apache.commons.lang.StringUtils;
import org.jetbrains.annotations.NotNull;

public class SummarizeRecentChangesRecipe {

  private final @NotNull Project project;
  private final @NotNull UpdatableChat chat;
  private final @NotNull RecipeRunner recipeRunner;

  public SummarizeRecentChangesRecipe(
      @NotNull Project project, @NotNull UpdatableChat chat, @NotNull RecipeRunner recipeRunner) {
    this.project = project;
    this.chat = chat;
    this.recipeRunner = recipeRunner;
  }

  public void summarizeRecentChanges() {
    GraphQlLogger.logCodyEvent(this.project, "recipe:summarize-recent-code-changes", "clicked");
    if (!this.isAnyVcsEnabled()) {
      chat.activateChatTab();
      chat.addMessageToChat(
          ChatMessage.createAssistantMessage(
              "No version control system is available in the project. Please enable one and try again."));
      return;
    }
    VcsLogFilterOptionsRegistry vcsLogFilterOptionsRegistry = new VcsLogFilterOptionsRegistry();
    FileEditorManager fileEditorManager = FileEditorManager.getInstance(project);
    VirtualFile[] selectedFiles = fileEditorManager.getSelectedFiles();
    if (!ArrayUtils.isEmpty(selectedFiles)) {
      VirtualFile currentlyOpenedFile = selectedFiles[0];
      FilePath filePath = new LocalFilePath(currentlyOpenedFile.getPath(), false);
      vcsLogFilterOptionsRegistry.addFilterOption(
          new Last5ItemsFromCurrentFileFilterOption(filePath));
    }
    SelectOptionManager selectOptionManager = SelectOptionManager.getInstance(project);
    selectOptionManager.show(
        project,
        vcsLogFilterOptionsRegistry.getAllOptions(),
        (selectedOption) -> {
          Supplier<VcsFilter> filterSupplierForOption =
              vcsLogFilterOptionsRegistry.getFilterSupplierForOption(selectedOption);
          VcsFilter vcsFilter = filterSupplierForOption.get();

          Optional<String> commitsMetadataDescription =
              new VcsCommitsMetadataLoader(project).loadCommitsMetadataDescription(vcsFilter);
          commitsMetadataDescription.ifPresent(
              it -> {
                if (StringUtils.isEmpty(it)) {
                  ApplicationManager.getApplication()
                      .invokeLater(
                          () -> {
                            chat.activateChatTab();
                            chat.addMessageToChat(
                                ChatMessage.createHumanMessage(
                                    "", vcsFilter.getFilterDescription()));
                            chat.addMessageToChat(
                                ChatMessage.createAssistantMessage("No recent changes found."));
                          });
                  return;
                }
                recipeRunner.runRecipe(
                    new SummarizeRecentChangesPromptProvider(vcsFilter.getFilterDescription()), it);
              });
        });
  }

  private boolean isAnyVcsEnabled() {
    return !VcsProjectLog.getLogProviders(project).isEmpty();
  }
}
