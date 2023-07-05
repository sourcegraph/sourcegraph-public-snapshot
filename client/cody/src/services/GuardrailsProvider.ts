import { Editor } from '@sourcegraph/cody-shared/src/editor'
import { DocumentOffsets } from '@sourcegraph/cody-shared/src/editor/offsets'
import { Guardrails, summariseAttribution } from '@sourcegraph/cody-shared/src/guardrails'

export class GuardrailsProvider {
    // TODO(keegancsmith) this provider should create the client since the guardrails client requires a dotcom graphql connection.
    constructor(private client: Guardrails, private editor: Editor) {}

    public async debugEditorSelection(): Promise<void> {
        const document = await this.editor.getFullTextDocument(this.editor.getActiveLightTextDocument()!)

        if (!document.selection) {
            return
        }

        const offset = new DocumentOffsets(document.content)

        const msg = await this.client
            .searchAttribution(offset.jointRangeSlice(document.selection))
            .then(summariseAttribution)

        await this.editor.warn(msg)
    }
}
