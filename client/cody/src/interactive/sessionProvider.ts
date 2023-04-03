import * as vscode from 'vscode'

export class InteractiveSessionProvider implements vscode.InteractiveSessionProvider {
    constructor() {}

    public async provideInitialSuggestions?(token: vscode.CancellationToken): Promise<string[]> {
        return ['Hello, world!']
    }

    public async provideWelcomeMessage?(
        token: vscode.CancellationToken
    ): Promise<vscode.InteractiveWelcomeMessageContent[]> {
        return ['Hello, world!']
    }

    public provideFollowups?(
        session: InteractiveSession,
        token: vscode.CancellationToken
    ): Promise<(string | vscode.InteractiveSessionFollowup)[]> {}

    public provideSlashCommands?(
        session: InteractiveSession,
        token: vscode.CancellationToken
    ): Promise<vscode.InteractiveSessionSlashCommand[]> {}

    public prepareSession(
        initialState: vscode.InteractiveSessionState | undefined,
        token: vscode.CancellationToken
    ): vscode.InteractiveSession {
        return {
            requester: { name: 'sqs' },
            responder: { name: 'Cody' },
            inputPlaceholder: "Ask Cody anything (questions, refactors, fixes, etc.) or type '/' for commands",
        }
    }

    public async resolveRequest(
        session: vscode.InteractiveSession,
        context: vscode.InteractiveSessionRequestArgs | string,
        token: vscode.CancellationToken
    ): Promise<vscode.InteractiveRequest> {
        const request: vscode.InteractiveRequest = { session, message: 'Thank you, my name is Cody' }
        return request
    }

    public async provideResponseWithProgress(
        request: vscode.InteractiveRequest,
        progress: vscode.Progress<vscode.InteractiveProgress>,
        token: vscode.CancellationToken
    ): Promise<vscode.InteractiveResponseForProgress> {
        progress.report({ content: 'Got it, this is a response' })
    }
}
