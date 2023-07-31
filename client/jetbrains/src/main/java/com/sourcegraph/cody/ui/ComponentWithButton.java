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
import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Component;
import java.awt.event.ActionListener;
import java.awt.event.KeyEvent;
import java.lang.ref.WeakReference;
import javax.swing.Icon;
import javax.swing.JComponent;
import javax.swing.JPanel;
import javax.swing.KeyStroke;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/**
 * This is a modification of a component from the IntelliJ Platform. We have removed some
 * unnecessary code that couldn't be overridden
 *
 * @see com.intellij.openapi.ui.ComponentWithBrowseButton
 */
public class ComponentWithButton<Comp extends JComponent> extends JPanel implements Disposable {
  @NotNull private final Comp component;
  @Nullable private final FixedSizeButton button;
  protected boolean componentDisabledOverride = false;

  public ComponentWithButton(@NotNull Comp component) {
    // Mac and Darcula have no horizontal gap, while other themes have a 2px gap.
    super(new BorderLayout(SystemInfo.isMac || StartupUiUtil.isUnderDarcula() ? 0 : 2, 0));

    // Add the component to the panel.
    this.component = component;
    // Required! Otherwise, the JPanel will occasionally gain focus instead of the component.
    setFocusable(false);
    add(this.component, BorderLayout.CENTER);

    // Create a button with a fixed size and add it to the panel.
    button = new FixedSizeButton(this.component);
    if (isBackgroundSet()) {
      button.setBackground(getBackground());
    }
    add(button, BorderLayout.EAST);

    new LazyDisposable(this);
  }

  public void setIconTooltip(@NotNull String tooltip) {
    if (button != null) {
      button.setToolTipText(tooltip);
    }
  }

  @Override
  public void setEnabled(boolean enabled) {
    super.setEnabled(enabled);
    if (button != null) {
      button.setEnabled(enabled);
    }
    component.setEnabled(enabled && !componentDisabledOverride);
  }

  public void setComponentDisabledOverride(boolean disabled) {
    componentDisabledOverride = disabled;
    if (button != null) {
      component.setEnabled(button.isEnabled() && !disabled);
    }
  }

  public void setButtonIcon(@NotNull Icon icon) {
    if (button != null) {
      button.setIcon(icon);
      button.setDisabledIcon(IconLoader.getDisabledIcon(icon));
    }
  }

  @Override
  public void setBackground(Color color) {
    super.setBackground(color);
    if (button != null) {
      button.setBackground(color);
    }
  }

  /** Adds specified {@code listener} to the button. */
  public void addButtonActionListener(ActionListener listener) {
    if (button != null) {
      button.addActionListener(listener);
    }
  }

  @Override
  public void dispose() {
    if (button != null) {
      ActionListener[] listeners = button.getActionListeners();
      for (ActionListener listener : listeners) {
        button.removeActionListener(listener);
      }
    }
  }

  @Override
  public final void requestFocus() {
    IdeFocusManager.getGlobalInstance()
        .doWhenFocusSettlesDown(
            () -> IdeFocusManager.getGlobalInstance().requestFocus(component, true));
  }

  @SuppressWarnings("deprecation")
  @Override
  public final void setNextFocusableComponent(Component aComponent) {
    super.setNextFocusableComponent(aComponent);
    component.setNextFocusableComponent(aComponent);
  }

  private KeyEvent currentEvent = null;

  /**
   * This method is overridden to dispatch the event to the component. This is necessary because
   * otherwise the event is dispatched to the parent component, which is the panel, and the event is
   * not dispatched to the component.
   *
   * @param ks The <code>KeyStroke</code> queried
   * @param e The <code>KeyEvent</code> forwarded to the focused component
   * @param condition one of the following values:
   *     <ul>
   *       <li>JComponent.WHEN_FOCUSED
   *       <li>JComponent.WHEN_ANCESTOR_OF_FOCUSED_COMPONENT
   *       <li>JComponent.WHEN_IN_FOCUSED_WINDOW
   *     </ul>
   *
   * @param pressed true if the key is pressed
   * @return true if there was a binding to the event, false otherwise.
   */
  @Override
  protected final boolean processKeyBinding(
      KeyStroke ks, KeyEvent e, int condition, boolean pressed) {
    if (condition == WHEN_FOCUSED && currentEvent != e) {
      try {
        currentEvent = e;
        component.dispatchEvent(e);
      } finally {
        currentEvent = null;
      }
    }
    if (e.isConsumed()) {
      return true;
    }
    return super.processKeyBinding(ks, e, condition, pressed);
  }

  /**
   * We need to register this component in the parent disposable. But we can't do it in the
   * constructor because the parent disposable is not yet available. So we do it lazily when the
   * component is shown.
   */
  private static final class LazyDisposable implements Activatable {
    private final WeakReference<ComponentWithButton<?>> reference;

    private LazyDisposable(ComponentWithButton<?> component) {
      reference = new WeakReference<>(component);
      new UiNotifyConnector.Once(component, this);
    }

    @Override
    public void showNotify() {
      ComponentWithButton<?> component = reference.get();
      if (component == null) {
        return; // component is collected
      }
      Disposable disposable =
          ApplicationManager.getApplication() == null
              ? null
              : UI_DISPOSABLE.getData(DataManager.getInstance().getDataContext(component));
      if (disposable == null) {
        return; // parent disposable not found
      }
      Disposer.register(disposable, component);
    }
  }
}
