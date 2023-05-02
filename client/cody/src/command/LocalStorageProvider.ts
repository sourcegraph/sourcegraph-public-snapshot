// VS Code Docs https://code.visualstudio.com/api/references/vscode-api#Memento
// A memento represents a storage utility. It can store and retrieve values.
import { Memento } from 'vscode'

import { OldUserLocalHistory, UserLocalHistory } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

export class LocalStorage {
    // Bump this on storage changes so we don't handle incorrectly formatted data
    private KEY_LOCAL_HISTORY = 'cody-local-chatHistory-v2'
    private KEY_LOCAL_HISTORY_MIGRATE = 'cody-local-chatHistory'

    constructor(private storage: Memento) {}

    public getChatHistory(): UserLocalHistory | null {
        let history = this.storage.get<UserLocalHistory | null>(this.KEY_LOCAL_HISTORY, null)
        if (this.storage.get(this.KEY_LOCAL_HISTORY_MIGRATE)) {
            // We override history as these users will have never used the new history key, so there's
            // no need to append - we can just set it outright
            history = this.getMigratedHistory()
            void this.storage.update(this.KEY_LOCAL_HISTORY_MIGRATE, null)
        }
        return history
    }

    public getMigratedHistory(): UserLocalHistory | null {
        const chunks = <T>(a: T[], size: number): T[][] =>
            Array.from(new Array(Math.ceil(a.length / size)), (_, i) => a.slice(i * size, i * size + size))

        const oldHistory = this.storage.get<OldUserLocalHistory | null>(this.KEY_LOCAL_HISTORY_MIGRATE, null)
        return oldHistory
            ? {
                  chat: Object.fromEntries(
                      Object.entries(oldHistory?.chat).map(([id, messages]) => [
                          id,
                          {
                              id,
                              // `Interaction.toChat()` flattens our messages into two elements
                              // so we iterate through messages in two elements chunks to reverse this
                              interactions: chunks(messages, 2).map(
                                  ([humanMessage, assistantMessageAndContextFiles]) => ({
                                      humanMessage,
                                      assistantMessage: assistantMessageAndContextFiles,
                                      context: assistantMessageAndContextFiles.contextFiles
                                          ? assistantMessageAndContextFiles.contextFiles.map(fileName => ({
                                                speaker: 'assistant',
                                                fileName,
                                            }))
                                          : [],
                                      // Timestamp not recoverable so we use the group timestamp
                                      timestamp: id,
                                  })
                              ),
                              lastInteractionTimestamp: id,
                          },
                      ])
                  ),
                  input: oldHistory.input,
              }
            : null
    }

    public async setChatHistory(history: UserLocalHistory): Promise<void> {
        try {
            await this.storage.update(this.KEY_LOCAL_HISTORY, history)
        } catch (error) {
            console.error(error)
        }
    }

    public async removeChatHistory(): Promise<void> {
        try {
            await this.storage.update(this.KEY_LOCAL_HISTORY, null)
        } catch (error) {
            console.error(error)
        }
    }

    public get(key: string): string | null {
        return this.storage.get(key, null)
    }

    public async set(key: string, value: string): Promise<void> {
        try {
            await this.storage.update(key, value)
        } catch (error) {
            console.error(error)
        }
    }
}
