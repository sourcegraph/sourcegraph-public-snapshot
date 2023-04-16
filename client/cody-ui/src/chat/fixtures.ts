import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

export const FIXTURE_TRANSCRIPT: Record<string, ChatMessage[]> = {
    simple: [
        { speaker: 'human', text: 'Hello, world!', displayText: 'Hello, world!' },
        { speaker: 'assistant', text: 'Thank you', displayText: 'Thank you' },
    ],
    codeQuestion: [
        {
            speaker: 'human',
            text: '',
            displayText: 'What does `document.getSelection()?.isCollapsed` mean?',
        },
        {
            speaker: 'assistant',
            text: '',
            displayText:
                '`document.getSelection()?.isCollapsed` means that the current selection in the document is collapsed, meaning it is a caret (no text is selected).\n\nThe `?.` operator is optional chaining - it will return `undefined` if `document.getSelection()` returns `null` or `undefined`.\n\nSo in short, that line is checking if there is currently a text selection in the document, and if not, focusing the textarea.\n\n',
        },
    ],
    explainCode: [
        {
            speaker: 'human',
            text: '',
            displayText:
                "Explain the following code at a high level:\n\n```\nprivate getNonce(): string {\n  let text = ''\n  const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'\n  for (let i = 0; i < 32; i++) {\n    text += possible.charAt(Math.floor(Math.random() * possible.length))\n  }\n  return text\n}\n```",
        },
        {
            speaker: 'assistant',
            text: '',
            displayText: 'This code generates a random 32-character string (nonce) using characters A-Z, a-z, and 0-9.',
            contextFiles: ['client/cody/src/chat/ChatViewProvider.ts', 'client/cody-shared/src/timestamp.ts'],
        },
        {
            speaker: 'human',
            text: '',
            displayText: 'Rewrite it to only use hexadecimal encoding.',
        },
        {
            speaker: 'assistant',
            text: '',
            displayText:
                "Here is the rewritten code using only hexadecimal encoding:\n\n```\nprivate getNonce(): string {\n  let text = ''\n  const possible = '0123456789ABCDEF'\n  for (let i = 0; i < 32; i++) {\n    text += possible.charAt(Math.floor(Math.random() * possible.length))\n  }\n  return text\n}\n```",
        },
    ],
}
