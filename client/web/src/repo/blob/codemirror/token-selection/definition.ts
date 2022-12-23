import { Facet, RangeSet, StateEffect, StateField } from '@codemirror/state'
import { Decoration, EditorView } from '@codemirror/view'
import * as H from 'history'

import { TextDocumentPositionParameters } from '@sourcegraph/client-api'
import { Location } from '@sourcegraph/extension-api-types'
import { getOrCreateCodeIntelAPI } from '@sourcegraph/shared/src/codeintel/api'
import { Occurrence, Position, Range } from '@sourcegraph/shared/src/codeintel/scip'
import { BlobViewState, parseRepoURI, toPrettyBlobURL, toURIWithPath } from '@sourcegraph/shared/src/util/url'

import { blobPropsFacet } from '..'
import { isInteractiveOccurrence, occurrenceAtMouseEvent, OccurrenceMap, rangeToCmSelection } from '../occurrence-utils'
import { LoadingTooltip } from '../tooltips/LoadingTooltip'
import { showTemporaryTooltip } from '../tooltips/TemporaryTooltip'
import { preciseOffsetAtCoords } from '../utils'

import { hoveredOccurrenceField } from './hover'
import { isModifierKey, isModifierKeyHeld } from './modifier-click'
import { selectOccurrence, selectRange } from './selections'

export interface DefinitionResult {
    handler: (position: Position) => void
    url?: string
    locations: Location[]
    atTheDefinition?: boolean
}
const definitionReady = Decoration.mark({
    class: 'cm-token-selection-definition-ready',
})
const setDefinitionEffect = StateEffect.define<OccurrenceMap<string>>()
const definitionUrlField = StateField.define<OccurrenceMap<string>>({
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

export const underlinedDefinitionFacet = Facet.define<unknown, unknown>({
    combine: props => props[0],
    enables: () => [
        definitionUrlField,
        EditorView.decorations.compute([definitionUrlField, hoveredOccurrenceField, isModifierKeyHeld], state => {
            const occ = state.field(hoveredOccurrenceField)
            const { value: url, hasOccurrence: hasDefinition } = state.field(definitionUrlField).get(occ)
            if (occ && state.field(isModifierKeyHeld) && hasDefinition) {
                const range = rangeToCmSelection(state, occ.range)
                if (range.from === range.to) {
                    return RangeSet.empty
                }
                if (url) {
                    // Insert an HTML link to support Context-menu>Open-link-in-new-tab
                    const definitionURL = Decoration.mark({
                        attributes: {
                            href: url,
                        },
                        tagName: 'a',
                    })
                    return RangeSet.of([definitionURL.range(range.from, range.to)])
                }
                return RangeSet.of([definitionReady.range(range.from, range.to)])
            }
            return RangeSet.empty
        }),
    ],
})

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
        selectOccurrence(view, atEvent.occurrence)
    }
    if (!isModifierKey(event) && !options?.isLongClick) {
        return
    }
    const spinner = new LoadingTooltip(view, preciseOffsetAtCoords(view, { x: event.clientX, y: event.clientY }))
    goToDefinitionAtOccurrence(view, atEvent.occurrence)
        .then(
            ({ handler }) => handler(atEvent.position),
            () => {}
        )
        .finally(() => spinner.stop())
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
    const api = await getOrCreateCodeIntelAPI(view.state.facet(blobPropsFacet).platformContext)
    const definition = await api.getDefinition(params).toPromise()
    if (definition.length === 0) {
        return {
            handler: position => showTemporaryTooltip(view, 'No definition found', position, 2000, { arrow: true }),
            locations: [],
        }
    }
    const locationFrom: Location = { uri: params.textDocument.uri, range: occurrence.range }
    for (const location of definition) {
        if (location.uri === params.textDocument.uri && location.range && location.range) {
            const requestPosition = new Position(params.position.line, params.position.character)
            const {
                start: { line: startLine, character: startCharacter },
                end: { line: endLine, character: endCharacter },
            } = location.range
            const resultRange = Range.fromNumbers(startLine, startCharacter, endLine, endCharacter)
            if (resultRange.contains(requestPosition)) {
                const refPanelURL = locationToURL(locationFrom, 'references')
                return {
                    url: refPanelURL,
                    atTheDefinition: true,
                    handler: position => {
                        showTemporaryTooltip(view, 'You are at the definition', position, 2000, { arrow: true })
                        const history = view.state.facet(blobPropsFacet).history
                        if (refPanelURL) {
                            history.replace(refPanelURL)
                        }
                    },
                    locations: definition,
                }
            }
        }
    }
    if (definition.length === 1) {
        const destination = definition[0]
        const hrefTo = locationToURL(destination)
        const { range, uri } = definition[0]
        if (hrefTo && range) {
            return {
                locations: definition,
                url: hrefTo,
                handler: () => {
                    interface DefinitionState {
                        // The destination URL if we trigger `history.goBack()`.  We use this state
                        // to avoid inserting redundant 'A->B->A->B' entries when the user triggers
                        // "go to definition" twice in a row from the same location.
                        previousURL?: string
                    }
                    const history = view.state.facet(blobPropsFacet).history as H.History<DefinitionState>
                    const selectionRange = Range.fromNumbers(
                        range.start.line,
                        range.start.character,
                        range.end.line,
                        range.end.character
                    )
                    const hrefFrom = locationToURL(locationFrom)
                    // Don't push URLs into the history if the last goto-def
                    // action was from the same URL same as this action. This
                    // happens when the user repeatedly triggers goto-def, which
                    // is easy to do when the the definition URL is close to
                    // where the action got triggered.
                    const shouldPushHistory = history.location.state?.previousURL !== hrefFrom
                    if (hrefFrom && shouldPushHistory && history.createHref(history.location) !== hrefFrom) {
                        history.push(hrefFrom)
                    }
                    if (uri === params.textDocument.uri) {
                        selectRange(view, selectionRange)
                    }
                    if (shouldPushHistory) {
                        history.push(hrefTo, { previousURL: hrefFrom })
                    }
                },
            }
        }
    }
    // Linking to the reference panel is a temporary workaround until we
    // implement a component to resolve ambiguous results inside the blob
    // view similar to how VS Code "Peek definition" works like.
    const refPanelURL = locationToURL(locationFrom, 'def')
    return {
        locations: definition,
        url: refPanelURL,
        handler: () => {
            if (refPanelURL) {
                const history = view.state.facet(blobPropsFacet).history
                history.push(refPanelURL)
            } else {
                // Should not happen but we handle this case because
                // `locationToURL` potentially returns undefined.
                showTemporaryTooltip(view, 'Multiple definitions found', params.position, 2000)
            }
        },
    }
}

function locationToURL(location: Location, viewState?: BlobViewState): string | undefined {
    const { range, uri } = location
    const { filePath, repoName, revision } = parseRepoURI(uri)
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
