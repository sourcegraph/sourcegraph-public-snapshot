import { Editor } from '@sourcegraph/cody-shared/src/editor'
import { Guardrails } from '@sourcegraph/cody-shared/src/guardrails'
import { isError } from '@sourcegraph/cody-shared/src/utils'

export class GuardrailsProvider {
    // TODO(keegancsmith) this provider should create the client since the guardrails client requires a dotcom graphql connection.
    constructor(private client: Guardrails, private editor: Editor) {}

    public async debugEditorSelection(): Promise<void> {
        const snippet = this.editor.getActiveTextEditorSelection()?.selectedText
        if (snippet === undefined) {
            return
        }

        const msg = await this.client.searchAttribution(snippet).then(result => {
            if (isError(result)) {
                return `guardrails attribution search failed: ${result.message}`
            }

            const count = result.length
            if (count === 0) {
                return 'no matching repositories found'
            }

            const summary = result.slice(0, count < 5 ? count : 5).map(repo => repo.name)
            if (count > 5) {
                summary.push('...')
            }

            return `found ${count} matching repositories ${summary.join(', ')}`
        })

        await this.editor.showWarningMessage(msg)
    }
}
