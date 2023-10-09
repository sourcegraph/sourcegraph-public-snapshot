import { type Extension, StateEffect, StateField } from '@codemirror/state'
import type { EditorView } from '@codemirror/view'
import { createPath } from 'react-router-dom'

import type { TextDocumentPositionParameters } from '@sourcegraph/client-api'
import type { Location } from '@sourcegraph/extension-api-types'
import { type Occurrence, Position, Range } from '@sourcegraph/shared/src/codeintel/scip'
import { type BlobViewState, parseRepoURI, toPrettyBlobURL, toURIWithPath } from '@sourcegraph/shared/src/util/url'

import { blobPropsFacet } from '..'
import { getCodeIntelAPI } from '../codeintel'
import { syntaxHighlight } from '../highlight'
import {
    isInteractiveOccurrence,
    occurrenceAtMouseEvent,
    occurrenceAtPosition,
    OccurrenceMap,
} from '../occurrence-utils'
import { LoadingTooltip } from '../tooltips/LoadingTooltip'
import { showTemporaryTooltip } from '../tooltips/TemporaryTooltip'
import { preciseOffsetAtCoords } from '../utils'

import { getCodeIntelTooltipState, selectOccurrence, setFocusedOccurrenceTooltip } from './code-intel-tooltips'
import { isModifierKey } from './modifier-click'

// The minimum number of milliseconds that must elapse before we handle a "Go to
// definition request".  The motivation to impose a minimum latency on this
// action is to give the user feedback that something happened if they rapidly
// trigger "Go to definition" from the same location and the destination token
// is already visible in the viewport.  Without this minimum latency, the user
// gets no feedback that the destination is visible.  With this latency, the
// source token (where the user clicks) gets briefly focused before the focus
// moves back to the destination token. This small wiggle in the focus state
// makes it easier to find the destination token.
const MINIMUM_LATENCY_MILLIS = 20

export interface DefinitionResult {
    /**
     * @param position The position where the request was sent. Used to determine the position of tooltips.
     * @param startTime The start time (via `Date.now()`) when the request was sent.
     */
    handler: (position: Position, startTime: number) => void
    url?: string
    locations: Location[]
    atTheDefinition?: boolean
}

const setDefinitionEffect = StateEffect.define<OccurrenceMap<string>>()
export const definitionUrlField = StateField.define<OccurrenceMap<string>>({
    create: () => new OccurrenceMap(new Map(), 'empty-definition'),
    update(value, transaction) {
        for (const effect of transaction.effects) {
            if (effect.is(setDefinitionEffect)) {
                value = effect.value
            }
        }
        return value
    },
})

export const definitionCache = StateField.define<Map<Occurrence, Promise<DefinitionResult>>>({
    create: () => new Map(),
    update: value => value,
})

export function definitionExtension(): Extension {
    return [definitionCache, definitionUrlField]
}

export function preloadDefinition(view: EditorView, occurrence: Occurrence): void {
    if (!view.state.field(definitionCache).has(occurrence)) {
        goToDefinitionAtOccurrence(view, occurrence).then(
            () => {},
            () => {}
        )
    }
}

export function goToDefinitionOnMouseEvent(
    view: EditorView,
    event: MouseEvent,
    options?: { isLongClick?: boolean }
): void {
    const atEvent = occurrenceAtMouseEvent(view, event)
    if (!atEvent) {
        return
    }
    if (isInteractiveOccurrence(atEvent.occurrence)) {
        selectOccurrence(view, atEvent.occurrence, true)

        // Ensure editor remains focused for the keyboard navigation to work
        view.contentDOM.focus()
    }
    if (!isModifierKey(event) && !options?.isLongClick) {
        return
    }

    const offset = preciseOffsetAtCoords(view, { x: event.clientX, y: event.clientY })
    if (offset === null) {
        return
    }

    view.dispatch({ effects: setFocusedOccurrenceTooltip.of(new LoadingTooltip(offset)) })
    const startTime = Date.now()
    goToDefinitionAtOccurrence(view, atEvent.occurrence)
        .then(
            ({ handler }) => handler(atEvent.position, startTime),
            () => {}
        )
        .finally(() => {
            // close loading tooltip if any
            const current = getCodeIntelTooltipState(view, 'focus')
            if (current?.tooltip instanceof LoadingTooltip && current?.occurrence === atEvent.occurrence) {
                view.dispatch({ effects: setFocusedOccurrenceTooltip.of(null) })
            }
        })
}

export function goToDefinitionAtOccurrence(view: EditorView, occurrence: Occurrence): Promise<DefinitionResult> {
    const cache = view.state.field(definitionCache)
    const fromCache = cache.get(occurrence)
    if (fromCache) {
        return fromCache
    }
    const uri = toURIWithPath(view.state.facet(blobPropsFacet).blobInfo)
    const params: TextDocumentPositionParameters = {
        position: occurrence.range.start,
        textDocument: { uri },
    }
    const promise = goToDefinition(view, occurrence, params)
    promise.then(
        ({ locations, url }) => {
            if (locations.length > 0) {
                const definitions = view.state.field(definitionUrlField)
                view.dispatch({ effects: setDefinitionEffect.of(definitions.withOccurrence(occurrence, url)) })
            }
        },
        () => {}
    )
    cache.set(occurrence, promise)
    return promise
}

async function goToDefinition(
    view: EditorView,
    occurrence: Occurrence,
    params: TextDocumentPositionParameters
): Promise<DefinitionResult> {
    const api = getCodeIntelAPI(view.state)
    const definition = await api.getDefinition(params, {
        referenceOccurrence: occurrence,
        documentOccurrences: view.state.facet(syntaxHighlight).occurrences,
    })
    const locationFrom: Location = { uri: params.textDocument.uri, range: occurrence.range }

    if (definition.length === 0) {
        return {
            handler: position => showTemporaryTooltip(view, 'No definition found', position, 2000, { arrow: true }),
            locations: [],
        }
    }

    for (const location of definition) {
        if (location.uri === params.textDocument.uri && location.range && location.range) {
            const {
                start: { line: startLine, character: startCharacter },
                end: { line: endLine, character: endCharacter },
            } = location.range
            const resultRange = Range.fromNumbers(startLine, startCharacter, endLine, endCharacter)
            const requestPosition = new Position(params.position.line, params.position.character)

            if (resultRange.contains(requestPosition)) {
                const refPanelURL = locationToURL(view, locationFrom, 'references')
                return {
                    url: refPanelURL,
                    atTheDefinition: true,
                    handler: position => {
                        showTemporaryTooltip(view, 'You are at the definition', position, 2000, { arrow: true })
                        const { navigate } = view.state.facet(blobPropsFacet)
                        if (refPanelURL) {
                            navigate(refPanelURL, { replace: true })
                        }
                    },
                    locations: definition,
                }
            }
        }
    }

    if (definition.length === 1) {
        const destination = definition[0]
        const hrefTo = locationToURL(view, destination)
        const { range, uri } = definition[0]

        if (hrefTo && range) {
            const handler: () => void = () => {
                interface DefinitionState {
                    // The destination URL if we trigger `history.goBack()`.  We use this state
                    // to avoid inserting redundant 'A->B->A->B' entries when the user triggers
                    // "go to definition" twice in a row from the same location.
                    previousURL?: string
                }

                const { location, navigate } = view.state.facet(blobPropsFacet)
                const locationState = location.state as DefinitionState

                const hrefFrom = locationToURL(view, locationFrom)
                // Don't push URLs into the history if the last goto-def
                // action was from the same URL same as this action. This
                // happens when the user repeatedly triggers goto-def, which
                // is easy to do when the definition URL is close to
                // where the action got triggered.
                const shouldPushHistory = locationState?.previousURL !== hrefFrom
                if (hrefFrom && shouldPushHistory && createPath(location) !== hrefFrom) {
                    navigate(hrefFrom)
                }
                if (uri === params.textDocument.uri) {
                    const definitionOccurrence = occurrenceAtPosition(
                        view.state,
                        new Position(range.start.line, range.start.character)
                    )
                    if (definitionOccurrence) {
                        selectOccurrence(view, definitionOccurrence)
                    }
                }
                if (shouldPushHistory) {
                    navigate(hrefTo, { state: { previousURL: hrefFrom } })
                }
            }
            return {
                locations: definition,
                url: hrefTo,
                handler: (_position, startTime) => {
                    const elapsed = Date.now() - startTime
                    if (elapsed < MINIMUM_LATENCY_MILLIS) {
                        setTimeout(handler, MINIMUM_LATENCY_MILLIS - elapsed)
                    } else {
                        handler()
                    }
                },
            }
        }
    }
    // Linking to the reference panel is a temporary workaround until we
    // implement a component to resolve ambiguous results inside the blob
    // view similar to how VS Code "Peek definition" works like.
    const refPanelURL = locationToURL(view, locationFrom, 'def')

    return {
        locations: definition,
        url: refPanelURL,
        handler: () => {
            if (refPanelURL) {
                view.state.facet(blobPropsFacet).navigate(refPanelURL)
            } else {
                // Should not happen but we handle this case because
                // `locationToURL` potentially returns undefined.
                showTemporaryTooltip(view, 'Multiple definitions found', params.position, 2000)
            }
        },
    }
}

function locationToURL(view: EditorView, location: Location, viewState?: BlobViewState): string | undefined {
    const { range, uri } = location
    const { filePath, repoName, revision: locationRevision } = parseRepoURI(uri)
    const { blobInfo } = view.state.facet(blobPropsFacet)

    // Try to preserve the non-canonical revision (branch name or empty revision)
    // when possible.
    const preserveNonCanonicalRevision =
        blobInfo.repoName === repoName && // Destination location is within the same repo.
        blobInfo.revision !== blobInfo.commitID && // Current URL is non-canonical/non-commit revision.
        blobInfo.commitID === locationRevision // Destination revision is the as as current revision.
    const revision = preserveNonCanonicalRevision ? blobInfo.revision : locationRevision

    if (filePath && range) {
        return toPrettyBlobURL({
            repoName,
            revision,
            filePath,
            position: { line: range.start.line + 1, character: range.start.character + 1 },
            range: location.range ? Range.fromExtensions(location.range).withIncrementedValues() : undefined,
            viewState,
        })
    }
    return undefined
}
