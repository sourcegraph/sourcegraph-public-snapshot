package com.sourcegraph.cody.ui;

import com.intellij.icons.AllIcons;
import com.intellij.ui.components.JBPasswordField;
import com.sourcegraph.cody.Icons;
import java.awt.event.ActionListener;
import java.awt.event.MouseAdapter;
import java.awt.event.MouseEvent;
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
  private boolean passwordLoaded = false;
  private boolean passwordVisible = false;
  private boolean passwordChanged = false;
  private boolean errorState = false;
  @NotNull private final JBPasswordField passwordField;

  // A function that retrieves the current password from storage.
  private final Supplier<String> passwordLoader;

  public PasswordFieldWithShowHideButton(
      @NotNull JBPasswordField passwordField, Supplier<String> passwordLoader) {
    this(passwordField, passwordLoader, 40);
  }

  private PasswordFieldWithShowHideButton(
      @NotNull JBPasswordField passwordField,
      Supplier<String> passwordLoader,
      int placeholderLength) {
    super(passwordField);
    this.passwordField = passwordField;
    this.echoChar = passwordField.getEchoChar();
    this.passwordLoader = passwordLoader;
    this.placeholder = StringUtils.repeat("x", placeholderLength);

    // Disable the password field by default so that the user can't type into it.
    setComponentDisabledOverride(true);

    ActionListener buttonActionListener =
        e -> {
          // Toggle password visibility
          passwordVisible = !passwordVisible;
          update();
        };
    addButtonActionListener(buttonActionListener);

    // Mark the password as changed whenever the user types into the field.
    DocumentListener documentListener = createDocumentListener();
    passwordField.getDocument().addDocumentListener(documentListener);

    // Handle click events on passwordField
    passwordField.addMouseListener(
        new MouseAdapter() {
          @Override
          public void mouseClicked(MouseEvent e) {
            if (!passwordVisible && !componentDisabledOverride) {
              buttonActionListener.actionPerformed(null);
            }
          }
        });

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

  public void resetUI() {
    passwordChanged = false;
    passwordVisible = false;
    passwordLoaded = false;
    setErrorState(false);
    update();
  }

  @Override
  public void setEnabled(boolean enabled) {
    if (!errorState) {
      super.setEnabled(enabled);
    }
  }

  private void setErrorState(boolean errorState) {
    this.errorState = errorState;
    if (errorState) {
      passwordField.setText("Access was denied to secure storage.");
      setEnabled(false);
    }
  }

  /** This should be called "updateUI" but that was already taken. :shrug: */
  private void update() {
    setComponentDisabledOverride(!passwordVisible);
    passwordField.setEchoChar(passwordVisible ? (char) 0 : echoChar);
    setButtonIcon(passwordVisible ? AllIcons.Actions.Show : Icons.Actions.Hide);
    if (!passwordLoaded) {
      if (passwordVisible) {
        String storedPassword = passwordLoader.get();
        if (storedPassword != null) {
          passwordField.setText(storedPassword);
          passwordLoaded = true;
        } else {
          setErrorState(true);
        }
      } else {
        passwordField.setText(placeholder);
      }
    }
    setIconTooltip(passwordVisible ? "Hide" : "Show");
  }

  public void setEmptyText(@NotNull String emptyText) {
    passwordField.getEmptyText().setText(emptyText);
  }

  /**
   * @return Null means we don't know the token because it wasn't loaded from the secure storage. An
   *     empty value means the user has explicitly set it to empty.
   */
  @Nullable
  public String getPassword() {
    String password =
        Optional.ofNullable(passwordField.getPassword()).map(String::copyValueOf).orElse("");
    // Known edge case: if the user's password is exactly the placeholder, we will think there's no
    // password.
    // We won't fix it because we currently only use the component for access tokens where this is
    // not a problem.
    return password.equals(placeholder) || errorState ? null : password;
  }

  @NotNull
  public Document getDocument() {
    return passwordField.getDocument();
  }

  public boolean hasPasswordChanged() {
    return passwordChanged;
  }
}
