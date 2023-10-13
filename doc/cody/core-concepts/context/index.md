# Cody Context

<p class="subtitle">Understand how context helps Cody write better and accurate code.</p>

Context refers to any additional information provided to help Cody understand and write code relevant to your codebase. While LLMs have extensive knowledge, they lack context about an individual or organization's codebase. Cody's ability to provide context-aware code responses is what sets it apart.

## Why is context important?

Context awareness is the key to Cody's ability to deliver high-quality responses to users. When Cody has access to the most relevant context about your codebase, it can:

- Answer questions about your codebase
- Produce unit tests and documentation
- Generate code that aligns with the libraries and style of your codebase
- Significantly reduce your work that's required to translate LLM-provided answers into actionable value for your users

## How context works with Cody prompts?

Cody works in conjunction with a LLM to provide codebase-aware answers. The LLM is a machine learning model that generates text in response to natural language prompts. However, the LLM doesn't inherently possess knowledge about your codebase or any specific coding requirements. Cody bridges this gap by generating context-aware prompts.

A typical prompt has three parts:

- **Prefix**: An optional description of the desired output, often derived from predefined [Commands](./../../capabilities.md#commands) that specify tasks the LLM can perform
- **User input**: The information provided including your code query or request
- **Context**: Additional information that helps the LLM provide a relevant answer based on your specific codebase

For example, a user querying Cody to "Explain the following Go code at a high level" might receive a prompt like this:
