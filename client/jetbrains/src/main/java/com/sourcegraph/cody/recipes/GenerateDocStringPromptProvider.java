package com.sourcegraph.cody.recipes;

public class GenerateDocStringPromptProvider implements PromptProvider {
  @Override
  public PromptContext getPromptContext(
      Language language, SelectedText selectedText, TruncatedText truncatedSelectedText) {
    MessageBuilder promptMessageBuilder =
        new MessageBuilder(language)
            .appendText(
                "Generate a comment documenting the parameters and functionality of the following"
                    + language.getValue()
                    + " code:")
            .appendCodeSnippet(truncatedSelectedText);
    if (language.getValue().equals("Java")) {
      promptMessageBuilder.appendText(
          "Use the JavaDoc documentation style to generate a Java comment.");
    } else if (language.getValue().equals("Python")) {
      promptMessageBuilder.appendText(
          "Use a Python docstring to generate a Python multi-line string.");
    } else {
      promptMessageBuilder.appendText(
          String.format(
              "Use the %s documentation style to generate a %s comment.",
              language.getValue(), language.getValue()));
    }
    String promptMessage =
        promptMessageBuilder
            .appendNewLine()
            .appendText(PromptMessages.MARKDOWN_FORMAT_PROMPT)
            .build();

    String displayText =
        new MessageBuilder(language)
            .appendText("Generate documentation for the following code:")
            .appendCodeSnippet(selectedText)
            .build();

    return new PromptContext(promptMessage, displayText, "Here is the generated documentation:\n");
  }
}
