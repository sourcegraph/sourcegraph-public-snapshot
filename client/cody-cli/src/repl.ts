import prompts from 'prompts'

import { Transcript } from '@sourcegraph/cody-shared/src/chat/transcript'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { IntentDetector } from '@sourcegraph/cody-shared/src/intent-detector'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

import { streamCompletions } from './completions'
import { interactionFromMessage } from './interactions'
import { getPreamble } from './preamble'

export async function startREPL(
    codebase: string,
    prompt: any,
    intentDetector: IntentDetector,
    codebaseContext: CodebaseContext,
    completionsClient: SourcegraphNodeCompletionsClient
) {
    if (prompt === undefined || prompt === '') {
        const response = await prompts({
            type: 'text',
            name: 'value',
            message: 'What do you want to ask Cody?',
        })

        prompt = response.value
    }

    const transcript = new Transcript()

    // TODO: Keep track of all user input if we add REPL mode

    const initialMessage: Message = { speaker: 'human', text: prompt }
    const messages: { human: Message; assistant?: Message }[] = [{ human: initialMessage }]
    for (const [index, message] of messages.entries()) {
        const interaction = await interactionFromMessage(
            message.human,
            intentDetector,
            // Fetch codebase context only for the last message
            index === messages.length - 1 ? codebaseContext : null
        )

        transcript.addInteraction(interaction)

        if (message.assistant?.text) {
            transcript.addAssistantResponse(message.assistant?.text)
        }
    }

    const finalPrompt = await transcript.toPrompt(getPreamble(codebase))

    let text = ''
    streamCompletions(completionsClient, finalPrompt, {
        onChange: chunk => {
            text = chunk
        },
        onComplete: () => {
            console.log(text)
        },
        onError: err => {
            console.error(err)
        },
    })
}
