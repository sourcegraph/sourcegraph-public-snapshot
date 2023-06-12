package com.sourcegraph.cody.recipes;

public class GenerateDocStringPromptProvider implements PromptProvider {
  @Override
  public PromptContext getPromptContext(
      String languageName, String selectedText, String truncatedSelectedText) {
    String promptPrefix =
        String.format(
            "Generate a comment documenting the parameters and functionality of the following %s code:\n",
            languageName);
    String additionalInstructions =
        String.format(
            "Use the %s documentation style to generate a %s comment.\n",
            languageName, languageName);
    if (languageName.equals("Java")) {
      additionalInstructions = "Use the JavaDoc documentation style to generate a Java comment.\n";
    } else if (languageName.equals("Python")) {
      additionalInstructions = "Use a Python docstring to generate a Python multi-line string.\n";
    }
    String promptMessage =
        String.format(
            promptPrefix
                + PromptMessages.CODE_SNIPPET_IN_LANGUAGE_FORMAT
                + "\nOnly generate the documentation, do not generate the code."
                + additionalInstructions
                + PromptMessages.MARKDOWN_FORMAT_PROMPT,
            languageName.toLowerCase(),
            truncatedSelectedText);

    String displayText =
        String.format(
            "Generate documentation for the following code:\n"
                + PromptMessages.CODE_SNIPPET_IN_LANGUAGE_FORMAT,
            languageName,
            selectedText);

    return new PromptContext(promptMessage, displayText);
  }
}
