# Chat

<p class="subtitle">Use Cody's chat to get contextually-aware answers to your questions.</p>

Cody **chat** allows you to ask coding-related questions about your entire codebase or a specific code snippet. You can do it from the **Chat** panel of the supported editor extensions (VS Code, JetBrains, and Neovim) or via the **Ask Cody** button in the web app.

The feature allows you to have an engaging conversation experience with Cody across VS Code, JetBrains, and Neovim. Key functionalities include managing chat history, stopping chat generation, editing sent messages, slash (/) commands, chat predictions, and more.

You can learn more about the IDE support for these functionalities in the [feature parity reference](./../feature-reference.md#chat).

#### *ADD VIDEO*

Demo video coming soon!

## Prerequisites

To use Cody's chat, you'll need to have the following:

- An active Sourcegraph instance (free or paid)
- A supported editor extension (VS Code, JetBrains, and Neovim) installed

## How does chat works?

Cody uses several search methods (including keyword and semantic search) to find files in your codebase that are relevant to your chat questions. It then uses context from those files to provide an informed response based on your codebase. Cody also tells you which code files it reads to generate its responses.

Cody can assist you with various use cases by answering questions such as:

- Generating an API call: Cody can analyze your API schema to provide context for the code it generates
- Locating a specific component in your codebase: Cody can identify and describe the files where a particular component is defined
- Handling questions that involve multiple files, like understanding data population in a React app: Cody can locate React component definitions, helping you understand how data is passed and where it originates

## Ask Cody your first question

Let's use Cody VS Code extension's chat interface to answer your first question.

- Click the Cody icon in the sidebar to view the **Chats** panel
- Next, click the icon for **Start a New Chat Panel** to open a new chat window
- Here, you can directly write your question or type slash `/` to select one of the commands and then press **Enter**

For example, ask Cody to "What does this file do?"

Cody will take a few seconds to process your question, providing contextual information about the files it reads and generating the answer.

## Ask Cody to write code

The chat feature can also write code for your questions. For example, ask Cody to "write a function that sorts an array in ascending order".

You are provided with code suggestions in the chat window along with the following options for using the code.

- The **Copy Code** icon to your clipboard and paste the code suggestion into your code editor
- Insert the code suggestion at the current cursor location by the **Insert Code at Cursor** icon
- The **Save Code to New File** icon to save the code suggestion to a new file in your project

During the chat, if Cody needs additional context, it can ask you to provide more information with a follow-up question. If your question is beyond the scope of the context, Cody will ask you to provide an alternate question aligned with the context of your codebase.

#### *This video will update*

![Example of Cody chat. You see the user ask Cody to describe what a file does, and Cody returns an answers that explains how the file is working in the context of the project.](https://storage.googleapis.com/sourcegraph-assets/cody/Docs/Sept2023/Context_Chat_SM.gif)
