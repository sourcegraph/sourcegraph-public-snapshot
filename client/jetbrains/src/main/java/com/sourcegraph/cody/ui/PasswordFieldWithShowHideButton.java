package com.sourcegraph.cody.ui;

import com.intellij.icons.AllIcons;
import com.intellij.ui.components.JBPasswordField;
import com.sourcegraph.cody.Icons;
import java.util.Optional;
import java.util.function.Supplier;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;
import org.apache.commons.lang3.StringUtils;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class PasswordFieldWithShowHideButton extends ComponentWithButton<JBPasswordField> {
  private final String placeholder;
  private final char echoChar;
  private boolean passwordVisible = false;
  private boolean passwordChanged = false;
  @NotNull private final JBPasswordField passwordField;

  // A function that retrieves the current password from storage.
  private final Supplier<String> passwordLoader;

  public PasswordFieldWithShowHideButton(@NotNull JBPasswordField passwordField, Supplier<String> passwordLoader, int placeholderLength) {
    super(passwordField);
    this.passwordField = passwordField;
    this.echoChar = passwordField.getEchoChar();
    this.passwordLoader = passwordLoader;
    this.placeholder = StringUtils.repeat("x", placeholderLength);

    // Disable the password field by default so that the user can't type into it.
    setComponentDisabledOverride(true);

    addButtonActionListener(
        e -> {
          // Toggle password visibility
          passwordVisible = !passwordVisible;
          update();

          // If the password hasn't been changed, load the old password from storage.
          if (!passwordChanged && passwordVisible) {
            String oldPassword = passwordLoader.get();
            if (oldPassword != null) {
              setPassword(oldPassword);
            }
          }
        });

    // Mark the password as changed whenever the user types into the field.
    DocumentListener documentListener = createDocumentListener();
    passwordField.getDocument().addDocumentListener(documentListener);

    // Update UI
    update();
  }

  @NotNull
  private DocumentListener createDocumentListener() {
    return new DocumentListener() {
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
  }

  @Override
  public void updateUI() {
    super.updateUI();
  }

  private void update() {
    setComponentDisabledOverride(!passwordVisible);
    passwordField.setEchoChar(passwordVisible ? (char) 0 : echoChar);
    setButtonIcon(passwordVisible ? AllIcons.Actions.Show : Icons.Actions.Hide);
    setIconTooltip(passwordVisible ? "Hide" : "Show");
  }

  public void setEmptyText(@NotNull String emptyText) {
    passwordField.getEmptyText().setText(emptyText);
  }

  /**
   * @return Null means we don't know the token because it wasn't loaded from the secure storage.
   *         An empty value means the user has explicitly set it to empty.
   */
  @Nullable
  public String getPassword() {
    String password = Optional.ofNullable(passwordField.getPassword()).map(String::copyValueOf).orElse("");
    // Known edge case: if the user's password is exactly the placeholder, we will think there's no password.
    // We won't fix it because we currently only use the component for access tokens where this is not a problem.
    return password.equals(placeholder) ? null : password;
  }

  private void setPassword(@NotNull String password) {
    passwordField.setText(password);
  }

  @NotNull
  public Document getDocument() {
    return passwordField.getDocument();
  }

  /**
   * Resets the password field to the old password if it was set, or to the mock token if it wasn't.
   */
  public void resetPassword(boolean storedPasswordExists) {
    if (storedPasswordExists) {
      setPassword(passwordVisible ? passwordLoader.get() : placeholder);
    } else {
      setPassword("");
    }

    passwordChanged = false;
  }

  public boolean hasPasswordChanged() {
    return passwordChanged;
  }
}
