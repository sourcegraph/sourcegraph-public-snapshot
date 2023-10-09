import type { KeyboardEvent, MouseEvent } from 'react'

import type { EditorView, Tooltip, TooltipView } from '@codemirror/view'
import { concat, from, type Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import type { HoverMerged } from '@sourcegraph/client-api'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import type { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { hasFindImplementationsSupport } from '@sourcegraph/shared/src/codeintel/api'
import type { Occurrence } from '@sourcegraph/shared/src/codeintel/scip'
import { type BlobViewState, toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'

import { blobPropsFacet } from '..'
import { type HoverData, createHovercard } from '../hovercard'
import { rangeToCmSelection } from '../occurrence-utils'
import { type DefinitionResult, goToDefinitionAtOccurrence } from '../token-selection/definition'
import { modifierClickDescription } from '../token-selection/modifier-click'
import { MOUSE_MAIN_BUTTON } from '../utils'

export interface HoverResult {
    markdownContents: string
    hoverMerged?: HoverMerged | null
    isPrecise?: boolean
}
export const emptyHoverResult: HoverResult = { markdownContents: '' }

// Helper to handle the cases for "No definition found" and "You are at the definition".
interface AsyncDefinitionResult extends DefinitionResult {
    asyncHandler: () => Promise<void>
}

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
    constructor(
        private readonly view: EditorView,
        private readonly occurrence: Occurrence,
        private readonly hover: HoverResult,
        // eslint-disable-next-line @typescript-eslint/explicit-member-accessibility
        readonly pinned: boolean
    ) {
        const range = rangeToCmSelection(view.state, occurrence.range)
        this.pos = range.from
        this.end = range.to
        this.create = () => {
            // To prevent the "Go to definition" from delaying the loading of
            // the popover, we provide an instant result that doesn't handle the
            // "No definition found" or "You are at the definition" cases. This
            // instant result gets dynamically replaced by the actual result once
            // it finishes loading.
            const instantDefinitionResult: AsyncDefinitionResult = {
                locations: [{ uri: '' }],
                handler: () => {},
                asyncHandler: () => {
                    const startTime = Date.now()
                    return goToDefinitionAtOccurrence(view, occurrence).then(
                        ({ handler }) => handler(occurrence.range.start, startTime),
                        () => {}
                    )
                },
            }
            const definitionResults: Observable<AsyncDefinitionResult> = concat(
                // Show active "Go to definition" button until we have resolved
                // a definition.
                of(instantDefinitionResult),
                // Trigger "Go to definition" to identify if this hover message
                // is already at the definition or if there are no references.
                from(goToDefinitionAtOccurrence(view, occurrence)).pipe(
                    map(result => ({ ...result, asyncHandler: instantDefinitionResult.asyncHandler }))
                )
            )
            const hovercardData: Observable<HoverData> = definitionResults.pipe(
                map(result => this.hovercardData(result))
            )

            return createHovercard(view, occurrence.range.withIncrementedValues(), pinned, hovercardData)
        }
    }
    private hovercardData(definition: AsyncDefinitionResult): HoverData {
        const { markdownContents } = this.hover
        const blobInfo = this.view.state.facet(blobPropsFacet).blobInfo
        const referencesURL = toPrettyBlobURL({
            ...blobInfo,
            range: this.occurrence.range.withIncrementedValues(),
            viewState: 'references',
        })
        const noDefinitionFound = definition.locations.length === 0
        const actions: ActionItemAction[] = [
            {
                active: true,
                disabledWhen: noDefinitionFound || definition.atTheDefinition,
                action: {
                    id: 'invokeFunction',
                    title: 'Go to definition',
                    disabledTitle: noDefinitionFound ? 'No definition found' : 'You are at the definition',
                    command: definition.url ? 'open' : 'invokeFunction-new',
                    commandArguments: definition.url
                        ? [
                              definition.url,
                              (event: MouseEvent<HTMLElement> | KeyboardEvent<HTMLElement>): boolean => {
                                  if (isRegularEvent(event)) {
                                      // "regular events" are basic clicks with the main button or keyboard
                                      // events without modifier keys.
                                      // We treat these the same way as Cmd-Click on the token itself.
                                      event.preventDefault()
                                      definition.asyncHandler().then(
                                          () => {},
                                          () => {}
                                      )
                                      return true
                                  }
                                  // Don't override `onSelect` unless it's a regular event with modifier keys
                                  // or with non-main buttons.
                                  // We do this to fallback to the browser's default behavior for links, for example to allow
                                  // the user to open the definition in a new browser tab.
                                  return false
                              },
                          ]
                        : [() => definition.asyncHandler()],
                },
            },
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
        if (
            this.hover.isPrecise &&
            hasFindImplementationsSupport(this.view.state.facet(blobPropsFacet).blobInfo.mode)
        ) {
            const implementationsURL = toPrettyBlobURL({
                ...blobInfo,
                range: this.occurrence.range.withIncrementedValues(),
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
        return {
            actionsOrError: actions,
            hoverOrError: {
                range: this.occurrence.range,
                aggregatedBadges: this.hover.hoverMerged?.aggregatedBadges,
                contents: [
                    {
                        value: markdownContents,
                        kind: MarkupKind.Markdown,
                    },
                ],
            },
        }
    }
}

// Returns true if this event is "regular", meaning the user is not holding down
// modifier keys or clicking with a non-main button.
function isRegularEvent(event: MouseEvent<HTMLElement> | KeyboardEvent<HTMLElement>): boolean {
    return (
        ('button' in event ? event.button === MOUSE_MAIN_BUTTON : true) &&
        !event.metaKey &&
        !event.shiftKey &&
        !event.ctrlKey
    )
}
