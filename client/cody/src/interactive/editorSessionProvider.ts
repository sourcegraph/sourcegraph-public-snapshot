import * as vscode from 'vscode'

import { BotResponseMultiplexer } from '@sourcegraph/cody-shared/src/chat/bot-response-multiplexer'
import { ChatClient } from '@sourcegraph/cody-shared/src/chat/chat'
import { getPreamble } from '@sourcegraph/cody-shared/src/chat/preamble'
import { FixupInteractive } from '@sourcegraph/cody-shared/src/chat/recipes/fixupInteractive'
import { Transcript } from '@sourcegraph/cody-shared/src/chat/transcript'
import { reformatBotMessage } from '@sourcegraph/cody-shared/src/chat/viewHelpers'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { ConfigurationWithAccessToken } from '@sourcegraph/cody-shared/src/configuration'
import { Editor } from '@sourcegraph/cody-shared/src/editor'
import { highlightTokens } from '@sourcegraph/cody-shared/src/hallucinations-detector'
import { IntentDetector } from '@sourcegraph/cody-shared/src/intent-detector'

import { ChatViewProvider } from '../chat/ChatViewProvider'
import { logEvent } from '../event-logger'

interface Session extends vscode.InteractiveEditorSession {
    text: string | null
}

interface Request extends vscode.InteractiveEditorRequest {
    session: Session
}

interface Response extends vscode.InteractiveEditorResponse {
    session: Session
}

interface MessageResponse extends vscode.InteractiveEditorMessageResponse {
    session: Session
}

type Config = Pick<
    ConfigurationWithAccessToken,
    'codebase' | 'serverEndpoint' | 'debug' | 'customHeaders' | 'accessToken'
>

export class InteractiveEditorSessionProvider
    implements vscode.InteractiveEditorSessionProvider<Session, Response | MessageResponse>
{
    private transcript: Transcript = new Transcript()
    private multiplexer: BotResponseMultiplexer = new BotResponseMultiplexer()
    private cancelCompletionCallback: (() => void) | null = null

    constructor(
        private chatViewProvider: ChatViewProvider,
        private config: Config,
        private chat: ChatClient,
        private intentDetector: IntentDetector,
        private codebaseContext: CodebaseContext,
        private editor: Editor
    ) {}

    // Create a session. The lifetime of this session is the duration of the editing session with the input mode widget.
    public prepareInteractiveEditorSession(
        context: vscode.TextDocumentContext,
        token: vscode.CancellationToken
    ): Session {
        const text = context.selection.isEmpty ? null : context.document.getText(context.selection)
        return { placeholder: "Ask Cody anything (questions, refactors, fixes, etc.) or type '/' for commands", text }
    }

    public async provideInteractiveEditorResponse(request: Request): Promise<Response | MessageResponse> {
        logEvent('CodyVSCodeExtension:interactiveEditorSession:requestSent')

        if (request.session.text !== null) {
            this.transcript.reset()

            let onSelectionChange2: ((content: string) => Promise<void>) | undefined
            const changedSelection = new Promise<string>(resolve => {
                onSelectionChange2 = content => {
                    resolve(content)
                    return Promise.resolve()
                }
            })
            if (!onSelectionChange2) {
                throw new Error('XXX')
            }
            const onSelectionChange = onSelectionChange2

            const recipe = new FixupInteractive(onSelectionChange)
            this.multiplexer = new BotResponseMultiplexer()
            const interaction = await recipe.getInteraction(request.prompt, {
                codebaseContext: this.codebaseContext,
                intentDetector: this.intentDetector,
                editor: this.editor,
                responseMultiplexer: this.multiplexer,
            })
            if (!interaction) {
                return { contents: new vscode.MarkdownString('Error'), session: request.session }
            }

            this.transcript.addInteraction(interaction)

            // TODO(sqs): need transcript.toPrompt?
            const promptMessages = await this.transcript.toPrompt(getPreamble(this.config.codebase))
            const responsePrefix = interaction.getAssistantMessage().prefix ?? ''

            /// ////////////////////////////////////////////////////

            this.cancelCompletion()

            let text = ''
            void new Promise<void>((resolve, reject) => {
                this.multiplexer.sub(BotResponseMultiplexer.DEFAULT_TOPIC, {
                    onResponse: (content: string) => {
                        text += content
                        this.transcript.addAssistantResponse(reformatBotMessage(text, responsePrefix))
                        return Promise.resolve()
                    },
                    onTurnComplete: async () => {
                        const lastInteraction = this.transcript.getLastInteraction()
                        if (lastInteraction) {
                            const { text, displayText } = lastInteraction.getAssistantMessage()
                            const { text: highlightedDisplayText } = await highlightTokens(displayText || '')
                            this.transcript.addAssistantResponse(text || '', highlightedDisplayText)
                            resolve()
                        }
                        this.cancelCompletionCallback = null
                    },
                })

                let textConsumed = 0

                this.cancelCompletionCallback = this.chat.chat(promptMessages, {
                    onChange: text => {
                        // TODO(dpc): The multiplexer can handle incremental text. Change chat to provide incremental text.
                        text = text.slice(textConsumed)
                        textConsumed += text.length
                        void this.multiplexer.publish(text)
                    },
                    onComplete: () => {
                        void this.multiplexer.notifyTurnComplete()
                    },
                    onError: err => {
                        void vscode.window.showErrorMessage(err)
                    },
                })
            })

            /// ////////////////////////////////////////////////////

            const changedSel2 = await changedSelection
            const response: Response = {
                session: request.session,
                edits: [new vscode.TextEdit(request.selection, changedSel2)],
            }
            return response
        }

        const response: MessageResponse = {
            session: request.session,
            contents: new vscode.MarkdownString('Because foo bar...'),
        }
        return response
    }

    private cancelCompletion(): void {
        this.cancelCompletionCallback?.()
        this.cancelCompletionCallback = null
    }

    public releaseInteractiveEditorSession?(session: Session): any {}

    public async handleInteractiveEditorResponseFeedback?(
        session: Session,
        response: vscode.InteractiveEditorResponse | vscode.InteractiveEditorMessageResponse,
        kind: vscode.InteractiveEditorResponseFeedbackKind
    ): void {}
}
