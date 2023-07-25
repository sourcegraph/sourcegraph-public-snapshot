package com.sourcegraph.cody.ui;

import static com.intellij.openapi.actionSystem.PlatformDataKeys.UI_DISPOSABLE;

import com.intellij.ide.DataManager;
import com.intellij.openapi.Disposable;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.ui.FixedSizeButton;
import com.intellij.openapi.util.Disposer;
import com.intellij.openapi.util.IconLoader;
import com.intellij.openapi.util.SystemInfo;
import com.intellij.openapi.wm.IdeFocusManager;
import com.intellij.util.ui.StartupUiUtil;
import com.intellij.util.ui.update.Activatable;
import com.intellij.util.ui.update.UiNotifyConnector;
import java.awt.*;
import java.awt.event.ActionListener;
import java.awt.event.KeyEvent;
import java.lang.ref.WeakReference;
import javax.swing.*;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/**
 * This is a modification of a component from the IntelliJ Platform. We have removed some
 * unnecessary code that couldn't be overridden
 *
 * @see com.intellij.openapi.ui.ComponentWithBrowseButton
 */
public class ComponentWithButton<Comp extends JComponent> extends JPanel implements Disposable {
  private final Comp myComponent;
  private final FixedSizeButton myBrowseButton;

  public ComponentWithButton(
      @NotNull Comp component, @Nullable ActionListener browseActionListener) {
    super(new BorderLayout(SystemInfo.isMac || StartupUiUtil.isUnderDarcula() ? 0 : 2, 0));

    myComponent = component;
    // required! otherwise JPanel will occasionally gain focus instead of the component
    setFocusable(false);
    add(myComponent, BorderLayout.CENTER);

    myBrowseButton = new FixedSizeButton(myComponent);
    if (isBackgroundSet()) {
      myBrowseButton.setBackground(getBackground());
    }
    if (browseActionListener != null) {
      myBrowseButton.addActionListener(browseActionListener);
    }
    add(myBrowseButton, BorderLayout.EAST);

    new LazyDisposable(this);
  }

  public void setIconTooltip(@NotNull String tooltip) {
    myBrowseButton.setToolTipText(tooltip);
  }

  @Override
  public void setEnabled(boolean enabled) {
    super.setEnabled(enabled);
    myBrowseButton.setEnabled(enabled);
    myComponent.setEnabled(enabled);
  }

  public void setButtonIcon(@NotNull Icon icon) {
    myBrowseButton.setIcon(icon);
    myBrowseButton.setDisabledIcon(IconLoader.getDisabledIcon(icon));
  }

  @Override
  public void setBackground(Color color) {
    super.setBackground(color);
    if (myBrowseButton != null) {
      myBrowseButton.setBackground(color);
    }
  }

  /** Adds specified {@code listener} to the browse button. */
  public void addActionListener(ActionListener listener) {
    myBrowseButton.addActionListener(listener);
  }

  @Override
  public void dispose() {
    ActionListener[] listeners = myBrowseButton.getActionListeners();
    for (ActionListener listener : listeners) {
      myBrowseButton.removeActionListener(listener);
    }
  }

  @Override
  public final void requestFocus() {
    IdeFocusManager.getGlobalInstance()
        .doWhenFocusSettlesDown(
            () -> IdeFocusManager.getGlobalInstance().requestFocus(myComponent, true));
  }

  @SuppressWarnings("deprecation")
  @Override
  public final void setNextFocusableComponent(Component aComponent) {
    super.setNextFocusableComponent(aComponent);
    myComponent.setNextFocusableComponent(aComponent);
  }

  private KeyEvent myCurrentEvent = null;

  @Override
  protected final boolean processKeyBinding(
      KeyStroke ks, KeyEvent e, int condition, boolean pressed) {
    if (condition == WHEN_FOCUSED && myCurrentEvent != e) {
      try {
        myCurrentEvent = e;
        myComponent.dispatchEvent(e);
      } finally {
        myCurrentEvent = null;
      }
    }
    if (e.isConsumed()) return true;
    return super.processKeyBinding(ks, e, condition, pressed);
  }

  private static final class LazyDisposable implements Activatable {
    private final WeakReference<ComponentWithButton<?>> reference;

    private LazyDisposable(ComponentWithButton<?> component) {
      reference = new WeakReference<>(component);
      new UiNotifyConnector.Once(component, this);
    }

    @Override
    public void showNotify() {
      ComponentWithButton<?> component = reference.get();
      if (component == null) return; // component is collected
      Disposable disposable =
          ApplicationManager.getApplication() == null
              ? null
              : UI_DISPOSABLE.getData(DataManager.getInstance().getDataContext(component));
      if (disposable == null) return; // parent disposable not found
      Disposer.register(disposable, component);
    }
  }
}
