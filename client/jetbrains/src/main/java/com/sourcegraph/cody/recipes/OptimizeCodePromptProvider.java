package com.sourcegraph.cody.recipes;

public class OptimizeCodePromptProvider implements PromptProvider {
  @Override
  public PromptContext getPromptContext(
      Language language, OriginalText originalText, TruncatedText truncatedText) {
    String promptMessage =
        new MessageBuilder(language)
            .appendText(
                "Optimize the memory and time consumption of this code in " + language.getValue())
            .appendText(
                "You first tell if the code can/cannot be optimized, then"
                    + "if the code is optimizable, suggest a numbered list of possible optimizations in less than 50 words each,"
                    + "then provide Big O time/space comparison for old and new code and finally return updated code."
                    + "Include inline comments to explain the optimizations in updated code."
                    + "Show the old vs. new time/space optimizations in a table."
                    + "Don't include the input code in your response. Beautify the response for better readability."
                    + "Response format should be: This code can/cannot be optimzed. Optimization Steps: \n{}\n Time and Space Usage: \n{}\n Updated Code: \n{}"
                    + "However if no optimization is possible; just say the code is already optimized.")
            .appendCodeSnippet(truncatedText)
            .appendText(PromptMessages.MARKDOWN_FORMAT_PROMPT)
            .build();

    String displayMessage =
        new MessageBuilder(language)
            .appendText("Optimize the time and space consumption of the following code:")
            .appendCodeSnippet(originalText)
            .build();
    return new PromptContext(promptMessage, displayMessage);
  }
}
