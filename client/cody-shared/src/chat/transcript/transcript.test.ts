import assert from 'assert'

import { CodebaseContext } from '../../codebase-context'
import { MAX_AVAILABLE_PROMPT_LENGTH } from '../../prompt/constants'
import { Message } from '../../sourcegraph-api'
import {
    defaultKeywordContextFetcher,
    MockEditor,
    MockEmbeddingsClient,
    MockIntentDetector,
    newRecipeContext,
} from '../../test/mocks'
import { ChatQuestion } from '../recipes/chat-question'

import { Transcript } from '.'

async function generateLongTranscript(): Promise<{ transcript: Transcript; tokensPerInteraction: number }> {
    // Add enough interactions to exceed the maximum prompt length.
    const numInteractions = 100
    const transcript = new Transcript()
    for (let i = 0; i < numInteractions; i++) {
        const interaction = await new ChatQuestion().getInteraction(
            'ABCD'.repeat(256), // 256 tokens, 1 token is ~4 chars
            newRecipeContext()
        )
        transcript.addInteraction(interaction)

        const assistantResponse = 'EFGH'.repeat(256) // 256 tokens
        transcript.addAssistantResponse(assistantResponse)
    }

    return {
        transcript,
        tokensPerInteraction: 512, // 256 for question + 256 for response.
    }
}

describe('Transcript', () => {
    it('generates an empty prompt with no interactions', async () => {
        const transcript = new Transcript()
        const prompt = await transcript.toPrompt()
        assert.deepStrictEqual(prompt, [])
    })

    it('generates a prompt without context for a chat question', async () => {
        const interaction = await new ChatQuestion().getInteraction(
            'how do access tokens work in sourcegraph',
            newRecipeContext()
        )

        const transcript = new Transcript()
        transcript.addInteraction(interaction)

        const prompt = await transcript.toPrompt()
        const expectedPrompt = [
            { speaker: 'human', text: 'how do access tokens work in sourcegraph' },
            { speaker: 'assistant', text: '' },
        ]
        assert.deepStrictEqual(prompt, expectedPrompt)
    })

    it('generates a prompt with context for a chat question', async () => {
        const embeddings = new MockEmbeddingsClient({
            search: async () =>
                Promise.resolve({
                    codeResults: [{ fileName: 'src/main.go', startLine: 0, endLine: 1, content: 'package main' }],
                    textResults: [{ fileName: 'docs/README.md', startLine: 0, endLine: 1, content: '# Main' }],
                }),
        })

        const interaction = await new ChatQuestion().getInteraction(
            'how do access tokens work in sourcegraph',
            newRecipeContext({
                intentDetector: new MockIntentDetector({
                    isCodebaseContextRequired: async () => Promise.resolve(true),
                }),
                codebaseContext: new CodebaseContext(
                    { useContext: 'embeddings' },
                    embeddings,
                    defaultKeywordContextFetcher
                ),
            })
        )

        const transcript = new Transcript()
        transcript.addInteraction(interaction)

        const prompt = await transcript.toPrompt()
        const expectedPrompt = [
            { speaker: 'human', text: 'Use the following text from file `docs/README.md`:\n# Main' },
            { speaker: 'assistant', text: 'Ok.' },
            { speaker: 'human', text: 'Use following code snippet from file `src/main.go`:\n```go\npackage main\n```' },
            { speaker: 'assistant', text: 'Ok.' },
            { speaker: 'human', text: 'how do access tokens work in sourcegraph' },
            { speaker: 'assistant', text: '' },
        ]
        assert.deepStrictEqual(prompt, expectedPrompt)
    })

    it('generates a prompt for multiple chat questions, includes context for last question only', async () => {
        const embeddings = new MockEmbeddingsClient({
            search: async () =>
                Promise.resolve({
                    codeResults: [{ fileName: 'src/main.go', startLine: 0, endLine: 1, content: 'package main' }],
                    textResults: [{ fileName: 'docs/README.md', startLine: 0, endLine: 1, content: '# Main' }],
                }),
        })
        const intentDetector = new MockIntentDetector({ isCodebaseContextRequired: async () => Promise.resolve(true) })
        const codebaseContext = new CodebaseContext(
            { useContext: 'embeddings' },
            embeddings,
            defaultKeywordContextFetcher
        )

        const chatQuestionRecipe = new ChatQuestion()
        const transcript = new Transcript()

        const firstInteraction = await chatQuestionRecipe.getInteraction(
            'how do access tokens work in sourcegraph',
            newRecipeContext({
                intentDetector,
                codebaseContext,
            })
        )
        transcript.addInteraction(firstInteraction)

        const assistantResponse = 'By setting the Authorization header.'
        transcript.addAssistantResponse(assistantResponse)

        const secondInteraction = await chatQuestionRecipe.getInteraction(
            'how to create a batch change',
            newRecipeContext({
                intentDetector,
                codebaseContext,
            })
        )
        transcript.addInteraction(secondInteraction)

        const prompt = await transcript.toPrompt()
        const expectedPrompt = [
            { speaker: 'human', text: 'how do access tokens work in sourcegraph' },
            { speaker: 'assistant', text: assistantResponse },
            { speaker: 'human', text: 'Use the following text from file `docs/README.md`:\n# Main' },
            { speaker: 'assistant', text: 'Ok.' },
            { speaker: 'human', text: 'Use following code snippet from file `src/main.go`:\n```go\npackage main\n```' },
            { speaker: 'assistant', text: 'Ok.' },
            { speaker: 'human', text: 'how to create a batch change' },
            { speaker: 'assistant', text: '' },
        ]
        assert.deepStrictEqual(prompt, expectedPrompt)
    })

    it('should limit prompts to a maximum number of tokens', async () => {
        const { transcript, tokensPerInteraction } = await generateLongTranscript()

        const numExpectedInteractions = Math.floor(MAX_AVAILABLE_PROMPT_LENGTH / tokensPerInteraction)
        const numExpectedMessages = numExpectedInteractions * 2 // Each interaction has two messages.

        const prompt = await transcript.toPrompt()
        assert.deepStrictEqual(prompt.length, numExpectedMessages)
    })

    it('should limit prompts to a maximum number of tokens with preamble always included', async () => {
        const { transcript, tokensPerInteraction } = await generateLongTranscript()

        const preamble: Message[] = [
            { speaker: 'human', text: 'PREA'.repeat(tokensPerInteraction / 2) },
            { speaker: 'assistant', text: 'MBLE'.repeat(tokensPerInteraction / 2) },
            { speaker: 'human', text: 'PREA'.repeat(tokensPerInteraction / 2) },
            { speaker: 'assistant', text: 'MBLE'.repeat(tokensPerInteraction / 2) },
        ]

        const numExpectedInteractions = Math.floor(MAX_AVAILABLE_PROMPT_LENGTH / tokensPerInteraction)
        const numExpectedMessages = numExpectedInteractions * 2 // Each interaction has two messages.

        const prompt = await transcript.toPrompt(preamble)
        assert.deepStrictEqual(prompt.length, numExpectedMessages)
        assert.deepStrictEqual(preamble, prompt.slice(0, 4))
    })

    it('includes currently visible content from the editor', async () => {
        const editor = new MockEditor({
            getActiveTextEditorVisibleContent: () => ({ fileName: 'internal/lib.go', content: 'package lib' }),
        })
        const embeddings = new MockEmbeddingsClient({
            search: async () =>
                Promise.resolve({
                    codeResults: [{ fileName: 'src/main.go', startLine: 0, endLine: 1, content: 'package main' }],
                    textResults: [{ fileName: 'docs/README.md', startLine: 0, endLine: 1, content: '# Main' }],
                }),
        })
        const intentDetector = new MockIntentDetector({ isCodebaseContextRequired: async () => Promise.resolve(true) })
        const codebaseContext = new CodebaseContext(
            { useContext: 'embeddings' },
            embeddings,
            defaultKeywordContextFetcher
        )

        const chatQuestionRecipe = new ChatQuestion()
        const transcript = new Transcript()

        const interaction = await chatQuestionRecipe.getInteraction(
            'how do access tokens work in sourcegraph',
            newRecipeContext({
                editor,
                intentDetector,
                codebaseContext,
            })
        )
        transcript.addInteraction(interaction)

        const prompt = await transcript.toPrompt()
        const expectedPrompt = [
            { speaker: 'human', text: 'Use the following text from file `docs/README.md`:\n# Main' },
            { speaker: 'assistant', text: 'Ok.' },
            {
                speaker: 'human',
                text: 'Use following code snippet from file `src/main.go`:\n```go\npackage main\n```',
            },
            { speaker: 'assistant', text: 'Ok.' },
            {
                speaker: 'human',
                text: 'I have the `internal/lib.go` file opened in my editor. Use following code snippet from file `internal/lib.go`:\n```go\npackage lib\n```',
            },
            {
                speaker: 'assistant',
                text: 'Ok.',
            },
            { speaker: 'human', text: 'how do access tokens work in sourcegraph' },
            { speaker: 'assistant', text: '' },
        ]
        assert.deepStrictEqual(prompt, expectedPrompt)
    })

    it('does not include currently visible content from the editor if no codebase context is required', async () => {
        const editor = new MockEditor({
            getActiveTextEditorVisibleContent: () => ({ fileName: 'internal/lib.go', content: 'package lib' }),
        })
        const intentDetector = new MockIntentDetector({ isCodebaseContextRequired: async () => Promise.resolve(false) })

        const transcript = new Transcript()
        const interaction = await new ChatQuestion().getInteraction(
            'how do access tokens work in sourcegraph',
            newRecipeContext({
                editor,
                intentDetector,
            })
        )
        transcript.addInteraction(interaction)

        const prompt = await transcript.toPrompt()
        const expectedPrompt = [
            { speaker: 'human', text: 'how do access tokens work in sourcegraph' },
            { speaker: 'assistant', text: '' },
        ]
        assert.deepStrictEqual(prompt, expectedPrompt)
    })

    it('adds context for last interaction with non-empty context', async () => {
        const embeddings = new MockEmbeddingsClient({
            search: async () =>
                Promise.resolve({
                    codeResults: [{ fileName: 'src/main.go', startLine: 0, endLine: 1, content: 'package main' }],
                    textResults: [{ fileName: 'docs/README.md', startLine: 0, endLine: 1, content: '# Main' }],
                }),
        })
        const intentDetector = new MockIntentDetector({ isCodebaseContextRequired: async () => Promise.resolve(true) })
        const codebaseContext = new CodebaseContext(
            { useContext: 'embeddings' },
            embeddings,
            defaultKeywordContextFetcher
        )

        const chatQuestionRecipe = new ChatQuestion()
        const transcript = new Transcript()

        const firstInteraction = await chatQuestionRecipe.getInteraction(
            'how do batch changes work in sourcegraph',
            newRecipeContext({
                intentDetector,
                codebaseContext,
            })
        )
        transcript.addInteraction(firstInteraction)
        transcript.addAssistantResponse('Smartly.')

        const secondInteraction = await chatQuestionRecipe.getInteraction(
            'how do access tokens work in sourcegraph',
            newRecipeContext({
                intentDetector,
                codebaseContext,
            })
        )
        transcript.addInteraction(secondInteraction)
        transcript.addAssistantResponse('By setting the Authorization header.')

        const thirdInteraction = await chatQuestionRecipe.getInteraction(
            'how do to delete them',
            newRecipeContext({
                // Here, we use the default intent detector to disable context fetching.
                codebaseContext,
            })
        )
        transcript.addInteraction(thirdInteraction)

        const prompt = await transcript.toPrompt()
        const expectedPrompt = [
            { speaker: 'human', text: 'how do batch changes work in sourcegraph' },
            { speaker: 'assistant', text: 'Smartly.' },
            { speaker: 'human', text: 'Use the following text from file `docs/README.md`:\n# Main' },
            { speaker: 'assistant', text: 'Ok.' },
            { speaker: 'human', text: 'Use following code snippet from file `src/main.go`:\n```go\npackage main\n```' },
            { speaker: 'assistant', text: 'Ok.' },
            { speaker: 'human', text: 'how do access tokens work in sourcegraph' },
            { speaker: 'assistant', text: 'By setting the Authorization header.' },
            { speaker: 'human', text: 'how do to delete them' },
            { speaker: 'assistant', text: '' },
        ]
        assert.deepStrictEqual(prompt, expectedPrompt)
    })
})
