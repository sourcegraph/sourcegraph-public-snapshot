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
  private final JBPasswordField component;

  private final Supplier<String> oldPasswordLoader;

  public PasswordWithPreviewField(JBPasswordField component, Supplier<String> oldPasswordLoader) {
    super(component, null);
    this.component = component;
    this.echoChar = component.getEchoChar();
    this.oldPasswordLoader = oldPasswordLoader;

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

    component
        .getDocument()
        .addDocumentListener(
            new DocumentListener() {
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
            });

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
            String oldPassword = oldPasswordLoader.get();
            if (oldPassword != null) {
              setPassword(oldPassword);
              passwordChanged = false;
            }
          }
        });
  }

  private void showPassword(@NotNull JBPasswordField component) {
    component.setEchoChar((char) 0);
    setButtonIcon(Icons.Actions.Hide);
    setIconTooltip("Hide");
  }

  private void hidePassword(@NotNull JBPasswordField component) {
    component.setEchoChar(echoChar);
    setButtonIcon(AllIcons.Actions.Show);
    setIconTooltip("Show");
  }

  public void setEmptyText(@NotNull String emptyText) {
    component.getEmptyText().setText(emptyText);
  }

  @NotNull
  public String getPassword() {
    return Optional.ofNullable(component.getPassword()).map(String::copyValueOf).orElse("");
  }

  private void setPassword(@NotNull String password) {
    component.setText(password);
  }

  @NotNull
  public Document getDocument() {
    return component.getDocument();
  }

  public void resetPassword(boolean isOldPasswordSet, String mockToken) {
    if (passwordVisible) {
      if (isOldPasswordSet) {
        setPassword(oldPasswordLoader.get());
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
