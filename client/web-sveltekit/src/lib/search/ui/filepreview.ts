import { Decoration } from '@codemirror/view'

import type { ContentMatch } from '$lib/shared'

const decoration = Decoration.mark({ class: 'match-highlight' })
export function mapResultToCodeMirrorDecorations(result: ContentMatch) {
    return Decoration.set(
        result.chunkMatches
            ?.flatMap(match => match.ranges.map(range => decoration.range(range.start.offset, range.end.offset)))
            .sort((a, b) => a.from - b.from) ?? []
    )
}
