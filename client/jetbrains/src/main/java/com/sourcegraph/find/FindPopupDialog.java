package com.sourcegraph.find;

import com.intellij.ide.IdeEventQueue;
import com.intellij.ide.ui.UISettings;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.project.ProjectManager;
import com.intellij.openapi.project.ProjectManagerListener;
import com.intellij.openapi.ui.DialogWrapper;
import com.intellij.openapi.ui.popup.ActiveIcon;
import com.intellij.openapi.util.DimensionService;
import com.intellij.openapi.util.WindowStateService;
import com.intellij.openapi.wm.WindowManager;
import com.intellij.openapi.wm.impl.IdeFrameImpl;
import com.intellij.openapi.wm.impl.IdeGlassPaneImpl;
import com.intellij.ui.PopupBorder;
import com.intellij.ui.TitlePanel;
import com.intellij.ui.WindowMoveListener;
import com.intellij.ui.WindowResizeListener;
import com.intellij.ui.awt.RelativePoint;
import com.intellij.util.ui.JBUI;
import com.intellij.util.ui.UIUtil;
import com.sourcegraph.Icons;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import javax.swing.*;
import javax.swing.border.Border;
import java.awt.*;

public class FindPopupDialog extends DialogWrapper {
    private static final String SERVICE_KEY = "sourcegraph.find.popup";
    private final JComponent mainPanel;
    private final Project project;

    public FindPopupDialog(@Nullable Project project, JComponent myMainPanel) {
        super(project, false);

        this.project = project;
        this.mainPanel = myMainPanel;

        setTitle("Sourcegraph");
        getWindow().setMinimumSize(new Dimension(750, 420));

        init();
        addNativeFindInFilesBehaviors();

        // Avoid the show method to be blocking
        this.setModal(false);
        // Prevent the dialog from being cancelable by any default behaviors
        myCancelAction.setEnabled(false);

        super.show();
    }


    @Override
    protected @Nullable JComponent createCenterPanel() {
        TitlePanel titlePanel = new TitlePanel(new ActiveIcon(Icons.Logo).getRegular(), new ActiveIcon(Icons.Logo).getInactive());
        titlePanel.setText(getTitle());

        addMoveListeners(titlePanel);

        // Adding the center panel
        return JBUI.Panels.simplePanel()
            .addToTop(titlePanel)
            .addToCenter(mainPanel);
    }

    // This adds behaviors found in JetBrain's native FindPopupPanel:
    // https://sourcegraph.com/github.com/JetBrains/intellij-community/-/blob/platform/lang-impl/src/com/intellij/find/impl/FindPopupPanel.java
    private void addNativeFindInFilesBehaviors() {
        this.setUndecorated(true);
        ApplicationManager.getApplication().getMessageBus().connect(this.getDisposable()).subscribe(ProjectManager.TOPIC, new ProjectManagerListener() {
            @Override
            public void projectClosed(@NotNull Project project) {
                FindPopupDialog.this.doCancelAction();
            }
        });

        Window window = WindowManager.getInstance().suggestParentWindow(project);
        Component parent = UIUtil.findUltimateParent(window);
        RelativePoint showPoint = null;
        Point screenPoint = DimensionService.getInstance().getLocation(SERVICE_KEY, project);
        if (screenPoint != null) {
            if (parent != null) {
                SwingUtilities.convertPointFromScreen(screenPoint, parent);
                showPoint = new RelativePoint(parent, screenPoint);
            } else {
                showPoint = new RelativePoint(screenPoint);
            }
        }
        if (parent != null && showPoint == null) {
            int height = UISettings.getInstance().getShowNavigationBar() ? 135 : 115;
            if (parent instanceof IdeFrameImpl && ((IdeFrameImpl) parent).isInFullScreen()) {
                height -= 20;
            }
            showPoint = new RelativePoint(parent, new Point((parent.getSize().width - getPreferredSize().width) / 2, height));
        }

        addMoveListeners(this.mainPanel);

        Dimension panelSize = getPreferredSize();
        Dimension prev = DimensionService.getInstance().getSize(SERVICE_KEY, project);
        if (prev != null && prev.height < panelSize.height) prev.height = panelSize.height;
        Window dialogWindow = this.getPeer().getWindow();
        JRootPane root = ((RootPaneContainer) dialogWindow).getRootPane();

        IdeGlassPaneImpl glass = (IdeGlassPaneImpl) this.getRootPane().getGlassPane();
        WindowResizeListener resizeListener = new WindowResizeListener(
            root,
            JBUI.insets(4),
            null) {
            private Cursor myCursor;

            @Override
            protected void setCursor(@NotNull Component content, Cursor cursor) {
                if (myCursor != cursor || myCursor != Cursor.getDefaultCursor()) {
                    glass.setCursor(cursor, this);
                    myCursor = cursor;

                    if (content instanceof JComponent) {
                        IdeGlassPaneImpl.savePreProcessedCursor((JComponent) content, content.getCursor());
                    }
                    super.setCursor(content, cursor);
                }
            }
        };
        glass.addMousePreprocessor(resizeListener, myDisposable);
        glass.addMouseMotionPreprocessor(resizeListener, myDisposable);

        root.setWindowDecorationStyle(JRootPane.NONE);
        root.setBorder(PopupBorder.Factory.create(true, true));
        UIUtil.markAsPossibleOwner((Dialog) dialogWindow);
        dialogWindow.setBackground(UIUtil.getPanelBackground());
        dialogWindow.setMinimumSize(panelSize);
        dialogWindow.setSize(prev != null ? prev : panelSize);

        IdeEventQueue.getInstance().getPopupManager().closeAllPopups(false);
        if (showPoint != null) {
            this.setLocation(showPoint.getScreenPoint());
        } else {
            dialogWindow.setLocationRelativeTo(null);
        }
    }

    private void addMoveListeners(Component component) {
        WindowMoveListener windowListener = new WindowMoveListener(component);
        component.addMouseListener(windowListener);
        component.addMouseMotionListener(windowListener);
    }

    @Override
    protected JComponent createSouthPanel() {
        return null;
    }

    @Override
    protected @Nullable Border createContentPaneBorder() {
        return null;
    }

    public void hide() {
        saveSize();
        getPeer().getWindow().setVisible(false);
    }

    // The automatic size saving behavior for DialogWrapper does not work for us as it relies on disposing of the
    // dialog to persist the changes. We need to manually implement this behavior instead.
    private void saveSize() {
        String serviceKey = this.getDimensionServiceKey();
        WindowStateService windowStateService = WindowStateService.getInstance(project);

        Point location = getLocation();
        Dimension size = getSize();
        windowStateService.putLocation(serviceKey, location);
        windowStateService.putSize(serviceKey, size);
    }

    @Override
    public void show() {
        getPeer().getWindow().setVisible(true);
    }

    @Override
    protected String getDimensionServiceKey() {
        return SERVICE_KEY;
    }
}
