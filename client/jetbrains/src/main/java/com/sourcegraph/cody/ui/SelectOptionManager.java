package com.sourcegraph.cody.ui;

import com.intellij.ide.IdeEventQueue;
import com.intellij.ide.actions.BigPopupUI;
import com.intellij.ide.actions.runAnything.RunAnythingPopupUI;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.components.ServiceManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.popup.JBPopup;
import com.intellij.openapi.ui.popup.JBPopupFactory;
import com.intellij.openapi.util.Disposer;
import com.intellij.openapi.util.WindowStateService;
import com.intellij.util.ui.JBInsets;
import java.awt.*;
import java.util.List;
import java.util.function.Consumer;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class SelectOptionManager {
  private static final String LOCATION_SETTINGS_KEY = "cody.select.option.popup";
  @NotNull private final Project myProject;
  @Nullable private JBPopup myBalloon;
  @Nullable private SelectOptionPopupUI mySelectOptionPopupUI;
  @Nullable private Dimension myBalloonFullSize;

  public SelectOptionManager(@NotNull Project project) {
    myProject = project;
  }

  public static SelectOptionManager getInstance(@NotNull Project project) {
    return ServiceManager.getService(project, SelectOptionManager.class);
  }

  public void show(@NotNull Project project, List<String> options, Consumer<String> runOnSelect) {
    IdeEventQueue.getInstance().getPopupManager().closeAllPopups(false);

    mySelectOptionPopupUI = createView(project, options, runOnSelect);

    myBalloon =
        JBPopupFactory.getInstance()
            .createComponentPopupBuilder(
                mySelectOptionPopupUI, mySelectOptionPopupUI.getSearchField())
            .setProject(myProject)
            .setModalContext(false)
            .setCancelOnClickOutside(true)
            .setRequestFocus(true)
            .setCancelKeyEnabled(false)
            .addUserData("SIMPLE_WINDOW") // NON-NLS
            .setResizable(true)
            .setMovable(true)
            .setDimensionServiceKey(myProject, LOCATION_SETTINGS_KEY, true)
            .setLocateWithinScreenBounds(false)
            .createPopup();
    Disposer.register(myBalloon, mySelectOptionPopupUI);
    if (project != null) {
      Disposer.register(project, myBalloon);
    }

    Dimension size = mySelectOptionPopupUI.getMinimumSize();
    JBInsets.addTo(size, myBalloon.getContent().getInsets());
    myBalloon.setMinimumSize(size);

    Disposer.register(
        myBalloon,
        () -> {
          saveSize();
          mySelectOptionPopupUI = null;
          myBalloon = null;
          myBalloonFullSize = null;
        });

    if (mySelectOptionPopupUI.getViewType() == RunAnythingPopupUI.ViewType.SHORT) {
      myBalloonFullSize = WindowStateService.getInstance(myProject).getSize(LOCATION_SETTINGS_KEY);
      Dimension prefSize = mySelectOptionPopupUI.getPreferredSize();
      myBalloon.setSize(prefSize);
    }
    calcPositionAndShow(project, myBalloon);
  }

  private void calcPositionAndShow(Project project, JBPopup balloon) {
    Point savedLocation =
        WindowStateService.getInstance(myProject).getLocation(LOCATION_SETTINGS_KEY);

    if (project != null) {
      balloon.showCenteredInCurrentWindow(project);
    } else {
      balloon.showInFocusCenter();
    }

    // for first show and short mode popup should be shifted to the top screen half
    if (savedLocation == null && mySelectOptionPopupUI.getViewType() == BigPopupUI.ViewType.SHORT) {
      Point location = balloon.getLocationOnScreen();
      location.y /= 2;
      balloon.setLocation(location);
    }
  }

  public boolean isShown() {
    return mySelectOptionPopupUI != null && myBalloon != null && !myBalloon.isDisposed();
  }

  @SuppressWarnings("Duplicates")
  @NotNull
  private SelectOptionPopupUI createView(
      @NotNull Project project, List<String> options, Consumer<String> runOnSelect) {
    SelectOptionPopupUI view = new SelectOptionPopupUI(project, options, runOnSelect);

    view.setSearchFinishedHandler(
        () -> {
          if (isShown()) {
            myBalloon.cancel();
          }
        });

    view.addViewTypeListener(
        viewType -> {
          if (!isShown()) {
            return;
          }

          ApplicationManager.getApplication()
              .invokeLater(
                  () -> {
                    Dimension minSize = view.getMinimumSize();
                    JBInsets.addTo(minSize, myBalloon.getContent().getInsets());
                    myBalloon.setMinimumSize(minSize);

                    if (viewType == BigPopupUI.ViewType.SHORT) {
                      myBalloonFullSize = myBalloon.getSize();
                      JBInsets.removeFrom(myBalloonFullSize, myBalloon.getContent().getInsets());
                      myBalloon.pack(false, true);
                    } else {
                      if (myBalloonFullSize == null) {
                        myBalloonFullSize = view.getPreferredSize();
                        JBInsets.addTo(myBalloonFullSize, myBalloon.getContent().getInsets());
                      }
                      myBalloonFullSize.height =
                          Integer.max(myBalloonFullSize.height, minSize.height);
                      myBalloonFullSize.width = Integer.max(myBalloonFullSize.width, minSize.width);

                      myBalloon.setSize(myBalloonFullSize);
                    }
                  });
        });

    return view;
  }

  private void saveSize() {
    if (mySelectOptionPopupUI.getViewType() == BigPopupUI.ViewType.SHORT) {
      WindowStateService.getInstance(myProject).putSize(LOCATION_SETTINGS_KEY, myBalloonFullSize);
    }
  }
}
