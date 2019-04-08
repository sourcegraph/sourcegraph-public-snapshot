import { ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { EditorService } from '../../services/editorService'

/** @internal */
export interface ClientEditorAPI extends ProxyValue {
    $setCollapsed(editorId: string, collapsed: boolean): void
}

/** @internal */
export class ClientEditor implements ClientEditorAPI {
    public readonly [proxyValueSymbol] = true

    constructor(private editorService: EditorService) {}

    public $setCollapsed(editorId: string, collapsed: boolean): void {
        this.editorService.setCollapsed({ editorId }, collapsed)
    }
}
