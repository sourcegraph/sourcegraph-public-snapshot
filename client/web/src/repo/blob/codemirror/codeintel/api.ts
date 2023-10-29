import { type EditorState, Facet } from '@codemirror/state'
import { EditorView, Tooltip, TooltipView } from '@codemirror/view'
import { Observable, from } from 'rxjs'
import { map, startWith } from 'rxjs/operators'

import type { HoverMerged } from '@sourcegraph/client-api'
import { isMacPlatform } from '@sourcegraph/common'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import type { Range } from '@sourcegraph/extension-api-types'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import {
    CodeIntelAPI,
    findLanguageMatchingDocument,
    hasFindImplementationsSupport,
} from '@sourcegraph/shared/src/codeintel/api'
import { Occurrence } from '@sourcegraph/shared/src/codeintel/scip'
import { BlobViewState, parseRepoURI, toPrettyBlobURL, toURIWithPath } from '@sourcegraph/shared/src/util/url'
import type { UIRangeSpec } from '@sourcegraph/shared/src/util/url'

import { WebHoverOverlayProps } from '../../../../components/WebHoverOverlay'
import { syntaxHighlight } from '../highlight'
import { contains, isInteractiveOccurrence, occurrenceAt, rangeToCmSelection } from '../occurrence-utils'
import { isRegularEvent, locationToURL, positionToOffset } from '../utils'

/**
 * Hover information received from a hover source.
 */
type HoverData = Pick<WebHoverOverlayProps, 'hoverOrError' | 'actionsOrError'>

export interface TooltipViewOptions {
    readonly view: EditorView
    readonly token: UIRangeSpec['range']
    readonly hovercardData: Observable<HoverData>
}

export interface CodeIntelTooltipPosition {
    from: number
    to: number
}

export interface Location {
    readonly repoName: string
    readonly filePath: string
    readonly revision?: string
    readonly range: Range
}

type Definition =
    | { type: 'at-definition' | 'single' | 'multiple'; readonly destination: Location; readonly from: Location }
    | { type: 'none' }

interface GoToDefinitionOptions {
    newWindow: boolean
}

interface CodeIntelConfig {
    api: CodeIntelAPI
    documentInfo: {
        repoName: string
        filePath: string
        commitID: string
        revision?: string
    }
    mode: string
    createTooltipView: (options: TooltipViewOptions) => TooltipView
    goToDefinition(view: EditorView, definition: Definition, options?: GoToDefinitionOptions): void
}

export class CodeIntelAPIAdapter {
    private hoverCache = new Map<Occurrence, Promise<(Tooltip & { end: number }) | null>>()
    private documentHighlightCache = new Map<Occurrence, Promise<{ from: number; to: number }[]>>()
    private definitionCache = new Map<Occurrence, Promise<Definition>>()
    private occurrenceCache = new Map<number, Occurrence | null>()
    private documentURI: string
    private hasCodeIntelligenceSupport: boolean

    constructor(private config: CodeIntelConfig) {
        this.documentURI = toURIWithPath({
            repoName: config.documentInfo.repoName,
            filePath: config.documentInfo.filePath,
            commitID: config.documentInfo.commitID,
        })
        this.hasCodeIntelligenceSupport = !!findLanguageMatchingDocument({ uri: this.documentURI })
    }

    public async hasDefinitionAt(offset: number, state: EditorState): Promise<boolean> {
        const occurrence = this.findOccurrenceAt(offset, state)
        return !!occurrence && (await this.getDefinition(state, occurrence)).type !== 'none'
    }

    public findOccurrenceRangeAt(offset: number, state: EditorState): { from: number; to: number } | null {
        const occurrence = this.findOccurrenceAt(offset, state)
        if (occurrence === null || !isInteractiveOccurrence(occurrence)) {
            return null
        }
        const from = positionToOffset(state.doc, occurrence.range.start)
        if (from === null) {
            return null
        }
        const to = positionToOffset(state.doc, occurrence.range.end)

        return to === null ? null : { from, to }
    }

    private findOccurrenceAt(offset: number, state: EditorState): Occurrence | null {
        if (!this.hasCodeIntelligenceSupport) {
            return null
        }

        let occurrence = this.occurrenceCache.get(offset)
        if (occurrence) {
            return occurrence
        }

        occurrence = occurrenceAt(state, offset) ?? null
        for (
            let range = occurrence ? rangeToCmSelection(state.doc, occurrence.range) : { from: offset, to: offset },
                i = range.from;
            i <= range.to;
            i++
        ) {
            this.occurrenceCache.set(offset, occurrence)
        }

        return occurrence
    }

    private getDefinition(state: EditorState, occurrence: Occurrence): Promise<Definition> {
        // TODO: Remove syntaxHighlight dependency
        const occurrences = state.facet(syntaxHighlight).occurrences
        const fromCache = this.definitionCache.get(occurrence)
        if (fromCache) {
            return fromCache
        }

        const { documentInfo } = this.config
        const from = {
            ...documentInfo,
            range: occurrence.range,
        }

        const promise: Promise<Definition> = this.config.api
            .getDefinition(
                {
                    textDocument: { uri: this.documentURI },
                    position: occurrence.range.start,
                },
                { referenceOccurrence: occurrence, documentOccurrences: occurrences }
            )
            .then(locations => {
                for (const location of locations) {
                    const { repoName, filePath, revision } = parseRepoURI(location.uri)
                    if (
                        filePath &&
                        location.range &&
                        repoName === documentInfo.repoName &&
                        filePath === documentInfo.filePath &&
                        documentInfo.commitID === revision &&
                        contains(location.range, occurrence.range.start)
                    ) {
                        return {
                            type: 'at-definition',
                            from,
                            destination: {
                                repoName,
                                filePath,
                                revision,
                                range: location.range,
                            },
                        }
                    }
                }
                if (locations.length === 1) {
                    const location = locations[0]
                    const { repoName, filePath, revision } = parseRepoURI(location.uri)
                    if (!(filePath && location.range)) {
                        return { type: 'none' }
                    }
                    return {
                        type: 'single',
                        from,
                        destination: {
                            repoName,
                            filePath,
                            revision,
                            range: location.range,
                        },
                    }
                }
                return locations.length === 0 ? { type: 'none' } : { type: 'multiple', from, destination: from }
            })
        this.definitionCache.set(occurrence, promise)

        return promise
    }

    getDocumentHighlights(
        state: EditorState,
        range: { from: number; to: number }
    ): Promise<{ from: number; to: number }[]> {
        const occurrence = this.findOccurrenceAt(range.from, state)
        if (!occurrence) {
            return Promise.resolve([])
        }
        const fromCache = this.documentHighlightCache.get(occurrence)
        if (fromCache) {
            return fromCache
        }
        const promise = this.config.api
            .getDocumentHighlights({
                textDocument: { uri: this.documentURI },
                position: occurrence.range.start,
            })
            .then(result =>
                result
                    .map(({ range }) => ({
                        from: positionToOffset(state.doc, range.start),
                        to: positionToOffset(state.doc, range.end),
                    }))
                    .filter((range): range is { from: number; to: number } => range.to !== null && range.from !== null)
            )
        this.documentHighlightCache.set(occurrence, promise)
        return promise
    }

    getHoverTooltip(
        state: EditorState,
        position: CodeIntelTooltipPosition
    ): Promise<(Tooltip & { end: number }) | null> {
        const occurrence = this.findOccurrenceAt(position.from, state)
        if (!occurrence) {
            return Promise.resolve(null)
        }
        const fromCache = this.hoverCache.get(occurrence)
        if (fromCache) {
            return fromCache
        }

        // Preload definition
        void this.getDefinition(state, occurrence)

        const tooltip = this.config.api
            .getHover({
                position: occurrence.range.start,
                textDocument: { uri: this.documentURI },
            })
            .then((result): (Tooltip & { end: number }) | null => {
                let markdownContents: string =
                    result === null || result === undefined || result.contents.length === 0
                        ? ''
                        : result.contents
                              .map(({ value }) => value)
                              .join('\n\n----\n\n')
                              .trimEnd()
                const precise = isPrecise(result)
                if (!precise && markdownContents.length > 0 && !isInteractiveOccurrence(occurrence)) {
                    return null
                }
                if (markdownContents === '' && isInteractiveOccurrence(occurrence)) {
                    markdownContents = 'No hover information available'
                }
                return markdownContents
                    ? {
                          pos: position.from,
                          end: position.to,
                          above: true,
                          create: view =>
                              this.config.createTooltipView({
                                  view,
                                  token: occurrence.range.withIncrementedValues(),
                                  hovercardData: this.getActions(view, occurrence, precise).pipe(
                                      map(actions => ({
                                          actionsOrError: actions,
                                          hoverOrError: {
                                              range: occurrence.range,
                                              aggregatedBadges: result?.aggregatedBadges,
                                              contents: [
                                                  {
                                                      value: markdownContents,
                                                      kind: MarkupKind.Markdown,
                                                  },
                                              ],
                                          },
                                      }))
                                  ),
                              }),
                      }
                    : null
            })

        this.hoverCache.set(occurrence, tooltip)
        return tooltip
    }

    private getActions(
        view: EditorView,
        occurrence: Occurrence,
        isPrecise: boolean
    ): Observable<HoverData['actionsOrError']> {
        // Trigger "Go to definition" to identify if this hover message
        // is already at the definition or if there are no references.
        return from(this.getDefinition(view.state, occurrence)).pipe(
            // To prevent the "Go to definition" from delaying the loading of
            // the popover, we provide an instant result that doesn't handle the
            // "No definition found" or "You are at the definition" cases. This
            // instant result gets dynamically replaced by the actual result once
            // it finishes loading.
            startWith({ type: 'initial' as const }),
            map(definition => {
                const referencesURL = toPrettyBlobURL({
                    ...this.config.documentInfo,
                    range: occurrence.range.withIncrementedValues(),
                    viewState: 'references',
                })

                const actions: ActionItemAction[] = []
                if (definition.type === 'none' || definition.type === 'at-definition') {
                    actions.push({
                        disabledWhen: true,
                        active: true,
                        action: {
                            id: 'go-to-definition',
                            title: '',
                            disabledTitle:
                                definition.type === 'none' ? 'No definition found' : 'You are at the definition',
                            command: 'open',
                        },
                    })
                } else if (definition.type === 'initial') {
                    actions.push({
                        active: true,
                        action: {
                            id: 'go-to-definition',
                            title: 'Go to definition',
                            command: 'invokeFunction-new',
                            commandArguments: [() => this.goToDefinitionAtOccurrence(view, occurrence)],
                        },
                    })
                } else {
                    actions.push({
                        active: true,
                        action: {
                            id: 'go-to-definition',
                            title: 'Go to definition',
                            command: 'open',
                            commandArguments: [
                                locationToURL(this.config.documentInfo, definition.destination) ?? '',
                                (event: MouseEvent | KeyboardEvent): boolean => {
                                    if (isRegularEvent(event)) {
                                        // "regular events" are basic clicks with the main button or keyboard
                                        // events without modifier keys.
                                        // We treat these the same way as Cmd-Click on the token itself.
                                        event.preventDefault()
                                        this.goToDefinitionAtOccurrence(view, occurrence)
                                        return true
                                    }
                                    // Don't override `onSelect` unless it's a regular event with modifier keys
                                    // or with non-main buttons.
                                    // We do this to fallback to the browser's default behavior for links, for example to allow
                                    // the user to open the definition in a new browser tab.
                                    return false
                                },
                            ],
                        },
                    })
                }
                actions.push({
                    active: true,
                    action: {
                        id: 'findReferences',
                        title: 'Find references',
                        command: 'open',
                        commandArguments: [referencesURL],
                    },
                })

                if (isPrecise && hasFindImplementationsSupport(this.config.mode)) {
                    const implementationsURL = toPrettyBlobURL({
                        ...this.config.documentInfo,
                        range: occurrence.range.withIncrementedValues(),
                        viewState: `implementations_${this.config.mode}` as BlobViewState,
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
                        id: 'goToDefinition.help',
                        title: '?', // special marker for the MDI "Help" icon.
                        description: `Go to definition with ${modifierClickDescription}, long-click, or by pressing Enter with the keyboard. Display this popover by pressing Space with the keyboard.`,
                        command: '',
                    },
                })
                return actions
            })
        )
    }

    public async goToDefinitionAt(view: EditorView, offset: number, options?: GoToDefinitionOptions): Promise<void> {
        const occurrence = this.findOccurrenceAt(offset, view.state)
        if (occurrence) {
            this.goToDefinitionAtOccurrence(view, occurrence, options)
        }
    }

    private async goToDefinitionAtOccurrence(
        view: EditorView,
        occurrence: Occurrence,
        options?: GoToDefinitionOptions
    ): Promise<void> {
        this.config.goToDefinition(view, await this.getDefinition(view.state, occurrence), options)
    }
}

const modifierClickDescription = isMacPlatform() ? 'cmd+click' : 'ctrl+click'

function isPrecise(hover: HoverMerged | null | undefined): boolean {
    for (const badge of hover?.aggregatedBadges || []) {
        if (badge.text === 'precise') {
            return true
        }
    }
    return false
}

export const codeIntelAPI = Facet.define<CodeIntelAPIAdapter, CodeIntelAPIAdapter>({
    combine(values) {
        return values[0] ?? null
    },
})

export function getCodeIntelAPI(state: EditorState): CodeIntelAPIAdapter {
    const api = state.facet(codeIntelAPI)
    if (!api) {
        throw new Error('A CodeIntelAPI instance has to be provided via the `codeIntelAPI` facet.')
    }
    return api
}
