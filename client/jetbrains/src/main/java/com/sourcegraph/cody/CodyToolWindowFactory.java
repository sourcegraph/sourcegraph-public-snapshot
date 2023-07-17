package com.sourcegraph.cody;

import com.intellij.openapi.actionSystem.ActionManager;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.project.DumbAware;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.wm.ToolWindow;
import com.intellij.openapi.wm.ToolWindowFactory;
import com.intellij.ui.content.Content;
import com.intellij.ui.content.ContentFactory;
import com.sourcegraph.config.ConfigUtil;
import java.util.ArrayList;
import java.util.List;
import org.jetbrains.annotations.NotNull;

public class CodyToolWindowFactory implements ToolWindowFactory, DumbAware {
  @Override
  public boolean isApplicable(@NotNull Project project) {
    return ConfigUtil.isCodyEnabled() && ToolWindowFactory.super.isApplicable(project);
  }

  @Override
  public void createToolWindowContent(@NotNull Project project, @NotNull ToolWindow toolWindow) {
    CodyToolWindowContent toolWindowContent = new CodyToolWindowContent(project);
    UpdatableChatHolderService projectService =
        project.getService(UpdatableChatHolderService.class);
    projectService.setUpdatableChat(toolWindowContent);
    Content content =
        ContentFactory.SERVICE
            .getInstance()
            .createContent(toolWindowContent.getContentPanel(), "", false);
    content.setPreferredFocusableComponent(toolWindowContent.getPreferredFocusableComponent());
    toolWindow.getContentManager().addContent(content);
    List<AnAction> titleActions = new ArrayList<>();
    createTitleActions(titleActions);
    if (!titleActions.isEmpty()) {
      toolWindow.setTitleActions(titleActions);
    }
  }

  private void createTitleActions(@NotNull List<? super AnAction> titleActions) {
    AnAction action = ActionManager.getInstance().getAction("CodyChatActionsGroup");
    if (action != null) {
      titleActions.add(action);
    }
  }
}
