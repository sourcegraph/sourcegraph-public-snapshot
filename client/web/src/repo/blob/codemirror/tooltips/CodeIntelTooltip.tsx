import { EditorView, Tooltip, TooltipView } from '@codemirror/view'
import { of } from 'rxjs'

import { HoverMerged } from '@sourcegraph/client-api'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { hasFindImplementationsSupport } from '@sourcegraph/shared/src/codeintel/api'
import { Occurrence } from '@sourcegraph/shared/src/codeintel/scip'
import { BlobViewState, toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'

import { blobPropsFacet } from '..'
import { HovercardView, HoverData } from '../hovercard'
import { rangeToCmSelection } from '../occurrence-utils'
import { modifierClickDescription } from '../token-selection/modifier-click'

export interface HoverResult {
    markdownContents: string
    hoverMerged?: HoverMerged | null
    isPrecise?: boolean
}
export const emptyHoverResult: HoverResult = { markdownContents: '' }

// CodeMirror tooltip wrapper around the "code intel" popover.  Implemented as a
// class so that we can detect it with instanceof checks. This class
// reimplements logic from `getHoverActions` in
// 'client/shared/src/hover/actions.ts' because that function is difficult to
// reason about and has surprising behavior.
export class CodeIntelTooltip implements Tooltip {
    public readonly above = true
    public readonly pos: number
    public readonly end: number
    public readonly create: () => TooltipView
    constructor(view: EditorView, occurrence: Occurrence, hover: HoverResult) {
        const { markdownContents } = hover
        const range = rangeToCmSelection(view.state, occurrence.range)
        this.pos = range.from
        this.end = range.to
        this.create = () => {
            const blobInfo = view.state.facet(blobPropsFacet).blobInfo
            const referencesURL = toPrettyBlobURL({
                ...blobInfo,
                range: occurrence.range.withIncrementedValues(),
                viewState: 'references',
            })
            const actions: ActionItemAction[] = [
                {
                    active: true,
                    action: {
                        id: 'findReferences',
                        title: 'Find references',
                        command: 'open',
                        commandArguments: [referencesURL],
                    },
                },
            ]
            if (hover.isPrecise && hasFindImplementationsSupport(view.state.facet(blobPropsFacet).blobInfo.mode)) {
                const implementationsURL = toPrettyBlobURL({
                    ...blobInfo,
                    range: occurrence.range.withIncrementedValues(),
                    viewState: `implementations_${blobInfo.mode}` as BlobViewState,
                })
                actions.push({
                    active: true,
                    action: {
                        id: 'findImplementations',
                        title: 'Find implementations',
                        command: 'open',
                        commandArguments: [implementationsURL],
                    },
                })
            }
            actions.push({
                active: true,
                action: {
                    id: 'goToDefinition',
                    title: '?', // special marker for the MDI "Help" icon.
                    description: `Go to definition with ${modifierClickDescription}, long-click, or by pressing Enter with the keyboard. Display this popover by pressing Space with the keyboard.`,
                    command: '',
                },
            })
            const data: HoverData = {
                actionsOrError: actions,
                hoverOrError: {
                    range: occurrence.range,
                    aggregatedBadges: hover.hoverMerged?.aggregatedBadges,
                    contents: [
                        {
                            value: markdownContents,
                            kind: MarkupKind.Markdown,
                        },
                    ],
                },
            }
            return new HovercardView(view, occurrence.range.withIncrementedValues(), false, of(data))
        }
    }
}
