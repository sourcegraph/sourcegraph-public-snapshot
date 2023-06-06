import * as vscode from 'vscode'

export class FixupDiagnostics implements vscode.CodeActionProvider<vscode.CodeAction> {
    public static metadata: vscode.CodeActionProviderMetadata = {
        // TODO: Generalize this to other toil, particularly refactorings
        providedCodeActionKinds: [vscode.CodeActionKind.QuickFix],
    }

    provideCodeActions(
        document: vscode.TextDocument,
        // TODO: How should we combine the range from the provideCodeActions and the range in the diagnostic?
        range: vscode.Range | vscode.Selection,
        context: vscode.CodeActionContext,
        token: vscode.CancellationToken
    ): vscode.ProviderResult<(vscode.CodeAction | vscode.Command)[]> {
        if (context.triggerKind !== vscode.CodeActionTriggerKind.Invoke) {
            return
        }
        const diagnostics = context.diagnostics
        if (!diagnostics.length) {
            return
        }
        const result = []
        const omnibus = []
        // TODO: Add a heuristic to filter diagnostics we could solve now,
        // versus ones which are too complex to solve.
        for (const diagnostic of diagnostics) {
            if (diagnostic.severity == vscode.DiagnosticSeverity.Information) {
                // TODO: Should we skip this, also skip warnings, etc.?
                continue
            }
            const action = new vscode.CodeAction(`Cody, fix ${diagnostic.message}`, vscode.CodeActionKind.QuickFix)
            action.command = {
                // TODO: Take advantage of diagnostic.relatedinformation to provide context to Cody
                arguments: [`Fix this error: ${diagnostic.message}`, { uri: document.uri, range: diagnostic.range }],
                command: 'cody.non-stop.fixup-diagnostics',
                title: `Cody, fix ${diagnostic.message}`,
            }
            result.push(action)
            omnibus.push(diagnostic.message)
        }
        const everything = new vscode.CodeAction(`Cody, fix all these diagnostics`, vscode.CodeActionKind.QuickFix)
        everything.command = {
            arguments: [`Fix all these errors:\n${omnibus.join('\n')}`],
            command: 'cody.non-stop.fixup-diagnostics',
            title: 'Cody, fix all these diagnostics',
        }
        result.push()
        return result
    }

    resolveCodeAction?(
        codeAction: vscode.CodeAction,
        token: vscode.CancellationToken
    ): vscode.ProviderResult<vscode.CodeAction> {
        console.log('resolveCodeAction NYI')
        return codeAction
    }

    public async command(instruction: string, { uri, range }: { uri: vscode.Uri; range: vscode.Range }): Promise<void> {
        // Select around the diagnostic
        const editor = await vscode.window.showTextDocument(uri, {
            // TODO: Use a real range, not the whole file
            selection: new vscode.Selection(new vscode.Position(0, 0), new vscode.Position(1000, 0)),
            preserveFocus: false,
        })
        console.assert(vscode.window.activeTextEditor === editor)
        vscode.commands.executeCommand('cody.recipe.non-stop', `${instruction}`)
    }
}
