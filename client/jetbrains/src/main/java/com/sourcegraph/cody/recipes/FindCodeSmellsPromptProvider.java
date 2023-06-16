package com.sourcegraph.cody.recipes;

public class FindCodeSmellsPromptProvider implements PromptProvider {
  @Override
  public PromptContext getPromptContext(
      Language language, OriginalText originalText, TruncatedText truncatedText) {
    String promptMessage =
        new MessageBuilder(language)
            .appendText(
                "Find code smells, potential bugs, and unhandled errors in my "
                    + language.getValue()
                    + " code:")
            .appendCodeSnippet(truncatedText)
            .appendText(
                "List maximum five of them as a list (if you have more in mind, mention that these are the top five), with a short context, reasoning, and suggestion on each.")
            .appendNewLine()
            .appendText(
                "If you have no ideas because the code looks fine, feel free to say that it already looks fine.")
            .build();

    String displayText =
        new MessageBuilder(language)
            .appendText("Find code smells in the following code: ")
            .appendCodeSnippet(originalText)
            .build();
    return new PromptContext(promptMessage, displayText);
  }
}
