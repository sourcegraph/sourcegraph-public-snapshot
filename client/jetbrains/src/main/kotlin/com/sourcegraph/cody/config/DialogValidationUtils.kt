package com.sourcegraph.cody.config

import com.intellij.openapi.ui.ValidationInfo
import com.intellij.openapi.util.NlsContexts
import javax.swing.JTextField

object DialogValidationUtils {
  /** Returns [ValidationInfo] with [message] if [textField] is blank */
  fun notBlank(textField: JTextField, @NlsContexts.DialogMessage message: String): ValidationInfo? {
    return custom(textField, message) { !textField.text.isNullOrBlank() }
  }

  /** Returns [ValidationInfo] with [message] if [isValid] returns false */
  fun custom(
      textField: JTextField,
      @NlsContexts.DialogMessage message: String,
      isValid: () -> Boolean
  ): ValidationInfo? {
    return if (!isValid()) ValidationInfo(message, textField) else null
  }

  /**
   * Chains the [validators] so that if one of them returns non-null [ValidationInfo] the rest of
   * them are not checked
   */
  fun chain(vararg validators: Validator): Validator = {
    validators.asSequence().mapNotNull { it() }.firstOrNull()
  }
}

typealias Validator = () -> ValidationInfo?
