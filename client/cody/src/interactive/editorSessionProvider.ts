import * as vscode from 'vscode'

interface Session extends vscode.InteractiveEditorSession {
    text: string | null
}

interface Request extends vscode.InteractiveEditorRequest {
    session: Session
}

export class InteractiveEditorSessionProvider implements vscode.InteractiveEditorSessionProvider<Session, Request> {
    constructor() {}

    // Create a session. The lifetime of this session is the duration of the editing session with the input mode widget.
    public prepareInteractiveEditorSession(
        context: vscode.TextDocumentContext,
        token: vscode.CancellationToken
    ): Session {
        const text = context.selection.isEmpty ? null : context.document.getText(context.selection)
        return { placeholder: "Ask Cody anything (questions, refactors, fixes, etc.) or type '/' for commands", text }
    }

    public async provideInteractiveEditorResponse(
        request: Request,
        token: vscode.CancellationToken
    ): Promise<vscode.InteractiveEditorResponse | vscode.InteractiveEditorMessageResponse> {
        if (request.session.text !== null) {
            const response: vscode.InteractiveEditorResponse = {
                edits: [new vscode.TextEdit(request.selection, request.session.text?.replace(/Error/g, 'Foo'))],
            }
            return response
        }

        const response: vscode.InteractiveEditorMessageResponse = {
            contents: new vscode.MarkdownString('Because foo bar...'),
        }
        return response
    }

    public releaseInteractiveEditorSession?(session: Session): any {}

    public async handleInteractiveEditorResponseFeedback?(
        session: Session,
        response: vscode.InteractiveEditorResponse | vscode.InteractiveEditorMessageResponse,
        kind: vscode.InteractiveEditorResponseFeedbackKind
    ): void {}
}
