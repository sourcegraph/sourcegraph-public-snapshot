import { type EditorState, Facet } from '@codemirror/state'

import type { CodeIntelAPI } from '@sourcegraph/shared/src/codeintel/api'

export const codeIntelAPI = Facet.define<CodeIntelAPI, CodeIntelAPI>({
    combine(values) {
        return values[0] ?? null
    },
})

export function getCodeIntelAPI(state: EditorState): CodeIntelAPI {
    const api = state.facet(codeIntelAPI)
    if (!api) {
        throw new Error('A CodeIntelAPI instance has to be provided via the `codeIntelAPI` facet.')
    }
    return api
}
