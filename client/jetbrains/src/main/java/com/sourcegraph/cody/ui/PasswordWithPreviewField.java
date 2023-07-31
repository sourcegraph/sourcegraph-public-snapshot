package com.sourcegraph.cody.ui;

import com.intellij.icons.AllIcons;
import com.intellij.ui.components.JBPasswordField;
import com.sourcegraph.cody.Icons;
import java.awt.event.FocusAdapter;
import java.awt.event.FocusEvent;
import java.util.Optional;
import java.util.function.Supplier;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;
import org.jetbrains.annotations.NotNull;

public class PasswordWithPreviewField extends ComponentWithButton<JBPasswordField> {
  private final char echoChar;
  private boolean passwordVisible = false;
  private boolean passwordChanged = false;
  private final JBPasswordField passwordField;

  // A function that retrieves the current password from storage.
  private final Supplier<String> passwordLoader;

  public PasswordWithPreviewField(JBPasswordField component, Supplier<String> passwordLoader) {
    super(component, null);
    this.passwordField = component;
    this.echoChar = component.getEchoChar();
    this.passwordLoader = passwordLoader;

    // Disable the password field by default so that the user can't type into it.
    setComponentDisabledOverride(true);

    hidePassword(component);
    addActionListener(
        e -> {
          passwordVisible = !passwordVisible;
          if (passwordVisible) {
            showPassword(component);
          } else {
            hidePassword(component);
          }
        });

    DocumentListener listener = new DocumentListener() {
      @Override
      public void insertUpdate(DocumentEvent e) {
        passwordChanged = true;
      }

      @Override
      public void removeUpdate(DocumentEvent e) {
        passwordChanged = true;
      }

      @Override
      public void changedUpdate(DocumentEvent e) {
        passwordChanged = true;
      }
    };
    component.getDocument().addDocumentListener(listener);

    component.addFocusListener(
        new FocusAdapter() {
          @Override
          public void focusGained(FocusEvent e) {
            if (!passwordVisible) {
              component.setText("");
            }
          }
        });

    addActionListener(
        e -> {
          if (!passwordChanged) {
            String oldPassword = passwordLoader.get();
            if (oldPassword != null) {
              setPassword(oldPassword);
              passwordChanged = false;
            }
          }
        });
  }

  private void showPassword(@NotNull JBPasswordField component) {
    setComponentDisabledOverride(false);
    component.setEchoChar((char) 0);
    setButtonIcon(Icons.Actions.Hide);
    setIconTooltip("Hide");
  }

  private void hidePassword(@NotNull JBPasswordField component) {
    setComponentDisabledOverride(true);
    component.setEchoChar(echoChar);
    setButtonIcon(AllIcons.Actions.Show);
    setIconTooltip("Show");
  }

  public void setEmptyText(@NotNull String emptyText) {
    passwordField.getEmptyText().setText(emptyText);
  }

  @NotNull
  public String getPassword() {
    return Optional.ofNullable(passwordField.getPassword()).map(String::copyValueOf).orElse("");
  }

  private void setPassword(@NotNull String password) {
    passwordField.setText(password);
  }

  @NotNull
  public Document getDocument() {
    return passwordField.getDocument();
  }

  public void resetPassword(boolean isOldPasswordSet, String mockToken) {
    if (passwordVisible) {
      if (isOldPasswordSet) {
        setPassword(passwordLoader.get());
      } else {
        setPassword("");
      }
    } else {
      if (isOldPasswordSet) {
        setPassword(mockToken);
      } else {
        setPassword("");
      }
    }
    passwordChanged = false;
  }

  public boolean hasPasswordChanged() {
    return passwordChanged;
  }
}
