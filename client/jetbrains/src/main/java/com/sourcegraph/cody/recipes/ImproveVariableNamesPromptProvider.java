package com.sourcegraph.cody.recipes;

public class ImproveVariableNamesPromptProvider implements PromptProvider {
  @Override
  public PromptContext getPromptContext(
      Language language, SelectedText selectedText, TruncatedText truncatedSelectedText) {
    String promptMessage =
        new MessageBuilder(language)
            .appendText(
                "Improve the variable names in this "
                    + language.getValue()
                    + " code by replacing the variable names with new identifiers which succinctly capture the purpose of the variable. We want the new code to be a drop-in replacement, so do not change names bound outside the scope of this code, like function names or members defined elsewhere. Only change the names of local variables and parameters:")
            .appendCodeSnippet(truncatedSelectedText)
            .appendText(PromptMessages.MARKDOWN_FORMAT_PROMPT)
            .build();

    String displayText =
        new MessageBuilder(language)
            .appendText("Improve the variable names in the following code:")
            .appendCodeSnippet(selectedText)
            .build();

    return new PromptContext(promptMessage, displayText);
  }
}
