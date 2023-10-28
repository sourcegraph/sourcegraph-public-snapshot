import { Facet } from '@codemirror/state'

import { LineOrPositionOrRange } from '@sourcegraph/common'

export const pinnedLocation = Facet.define<LineOrPositionOrRange | null, LineOrPositionOrRange | null>({
    combine(values) {
        return values[0] ?? null
    },
})
