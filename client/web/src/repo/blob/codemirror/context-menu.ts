import {
    countColumn,
    EditorSelection,
    Extension,
    Facet,
    Line,
    SelectionRange,
    StateEffect,
    StateField,
} from '@codemirror/state'
import { EditorView, hoverTooltip, KeyBinding, keymap, showTooltip, Tooltip, TooltipView } from '@codemirror/view'
import { Remote } from 'comlink'
import * as H from 'history'

import { TextDocumentPositionParameters } from '@sourcegraph/client-api'
import { renderMarkdown } from '@sourcegraph/common'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { Occurrence, Position, Range } from '@sourcegraph/shared/src/codeintel/scip'
import { parseRepoURI, toPrettyBlobURL, toURIWithPath } from '@sourcegraph/shared/src/util/url'

import { BlobInfo } from '../Blob'

import { HighlightIndex, syntaxHighlight } from './highlight'
import { shouldScrollIntoView } from './linenumbers'
import { isInteractiveOccurrence } from './tokens-as-links'

import styles from './context-menu.module.scss'

function occurrenceAtPosition(
    view: EditorView,
    position: Position
): { occurrence: Occurrence; position: Position } | undefined {
    const table = view.state.facet(syntaxHighlight)
    for (
        let index = table.lineIndex[position.line];
        index !== undefined &&
        index < table.occurrences.length &&
        table.occurrences[index].range.start.line === position.line;
        index++
    ) {
        const occurrence = table.occurrences[index]
        if (occurrence.range.contains(position)) {
            return { occurrence, position }
        }
    }
    return
}

function closestOccurrence(
    line: number,
    table: HighlightIndex,
    position: Position,
    includeOccurrence?: (occurrence: Occurrence) => boolean
): Occurrence | undefined {
    const candidates: [Occurrence, number][] = []
    let index = table.lineIndex[line] ?? -1
    for (
        ;
        index >= 0 && index < table.occurrences.length && table.occurrences[index].range.start.line === line;
        index++
    ) {
        const occurrence = table.occurrences[index]
        if (!isInteractiveOccurrence(occurrence)) {
            continue
        }
        if (includeOccurrence && !includeOccurrence(occurrence)) {
            continue
        }
        candidates.push([occurrence, occurrence.range.characterDistance(position)])
    }
    candidates.sort(([, a], [, b]) => a - b)
    if (candidates.length > 0) {
        return candidates[0][0]
    }
    return undefined
}

function occurrenceAtEvent(
    view: EditorView,
    event: MouseEvent
): { occurrence: Occurrence; position: Position; coords: Coordinates } | undefined {
    const atEvent = positionAtEvent(view, event)
    if (!atEvent) {
        return
    }
    const { position, coords } = atEvent
    const occurrence = occurrenceAtPosition(view, position)
    if (!occurrence) {
        return
    }
    return { ...occurrence, coords }
}
function goToDefinitionAtOccurrence(
    view: EditorView,
    blobInfo: BlobInfo,
    history: H.History,
    codeintel: Remote<FlatExtensionHostAPI>,
    position: Position,
    occurrence: Occurrence,
    coords: Coordinates
): Promise<() => void> {
    const fromCache = definitionCache.get(occurrence)
    if (fromCache) {
        return fromCache
    }
    const uri = toURIWithPath(blobInfo)
    const promise = goToDefinition(view, history, codeintel, { position, textDocument: { uri } }, coords)
    definitionCache.set(occurrence, promise)
    return promise
}

function goToDefinitionAtEvent(
    view: EditorView,
    event: MouseEvent,
    blobInfo: BlobInfo,
    history: H.History,
    codeintel: Remote<FlatExtensionHostAPI>
): Promise<() => void> {
    const atEvent = occurrenceAtEvent(view, event)
    if (!atEvent) {
        return Promise.resolve(() => {})
    }
    const { occurrence, position, coords } = atEvent
    return goToDefinitionAtOccurrence(view, blobInfo, history, codeintel, position, occurrence, coords)
}

function positionAtEvent(view: EditorView, event: MouseEvent): { position: Position; coords: Coordinates } | undefined {
    const coords: Coordinates = {
        x: event.clientX,
        y: event.clientY,
    }
    const position = view.posAtCoords(coords)
    if (position === null) {
        return
    }
    event.preventDefault()
    return { position: scipPositionAtCodemirrorPosition(view, position), coords }
}

function scipPositionAtCodemirrorPosition(view: EditorView, position: number): Position {
    const cmLine = view.state.doc.lineAt(position)
    const line = cmLine.number - 1
    const character = position - cmLine.from
    return new Position(line, character)
}

const definitionCache = new Map<Occurrence, Promise<() => void>>()
const hoverCache = new Map<Occurrence, Promise<string>>()

// HACK: we store the editor view in a global variable so that we can access it
// from global keydown/keyup event handlers even when the editor is not focused.
// The `keydown` handler in EditorView.domEventHandler doesn't capture events
// when the editor is out of focus.
let globalViewHack: EditorView | undefined
const globalEventHandler = (event: KeyboardEvent): void => {
    if (!globalViewHack) {
        return
    }
    if (event.metaKey) {
        globalViewHack.contentDOM.classList.add(styles.clickable)
    } else {
        globalViewHack.contentDOM.classList.remove(styles.clickable)
    }
}

const scrollLineIntoView = (view: EditorView, line: Line): boolean => {
    if (shouldScrollIntoView(view, { line: line.number })) {
        view.dispatch({
            effects: EditorView.scrollIntoView(line.from, { y: 'nearest' }),
        })
        return true
    }
    return false
}

export const rangeToSelection = (view: EditorView, range: Range): SelectionRange => {
    const startLine = view.state.doc.line(range.start.line + 1)
    const endLine = view.state.doc.line(range.end.line + 1)
    const start = startLine.from + range.start.character
    const end = Math.min(endLine.from + range.end.character, endLine.to)
    return EditorSelection.range(start, end)
}
export const uriFacet = Facet.define<string, string>({
    combine: props => props[0],
})
export const selectionsFacet = Facet.define<Map<string, Range>, Map<string, Range>>({
    combine: props => props[0],
})
export const blobInfoFacet = Facet.define<BlobInfo, BlobInfo>({
    combine: props => props[0],
})

export const selectOccurrence = (
    codeintel: Remote<FlatExtensionHostAPI>,
    history: H.History,
    view: EditorView,
    occurrence: Occurrence
): void => {
    const blobInfo = view.state.facet(blobInfoFacet)
    hoverAtOccurrence(codeintel, blobInfo, occurrence).then(
        () => {},
        () => {}
    )
    const cmLine = view.state.doc.line(occurrence.range.start.line + 1)
    const rect = view.coordsAtPos(cmLine.from + occurrence.range.start.character + 1)
    const coords: Coordinates = rect ? { x: rect.left, y: rect.top } : { x: 0, y: 0 }
    goToDefinitionAtOccurrence(view, blobInfo, history, codeintel, occurrence.range.start, occurrence, coords).then(
        () => {},
        () => {}
    )
    const url = toPrettyBlobURL({ ...blobInfo, range: occurrence.range.asOneBased() })
    history.replace(url)
    selectRange(view, occurrence.range)
}
export const selectRange = (view: EditorView, range: Range): void => {
    const selection = rangeToSelection(view, range)
    const uri = view.state.facet(uriFacet)
    const selections = view.state.facet(selectionsFacet)
    if (selections) {
        selections.set(uri, range)
    }
    view.dispatch({ selection })

    const lineAbove = view.state.doc.line(Math.min(view.state.doc.lines, range.start.line + 3))
    if (scrollLineIntoView(view, lineAbove)) {
        return
    }
    const lineBelow = view.state.doc.line(Math.max(1, range.end.line - 1))
    scrollLineIntoView(view, lineBelow)
}

export function contextMenu(
    codeintel: Remote<FlatExtensionHostAPI> | undefined,
    blobInfo: BlobInfo,
    history: H.History,
    selections: Map<string, Range>
): Extension {
    if (!codeintel) {
        return []
    }
    document.removeEventListener('keydown', globalEventHandler)
    document.addEventListener('keydown', globalEventHandler)
    document.removeEventListener('keyup', globalEventHandler)
    document.addEventListener('keyup', globalEventHandler)
    const setHoverEffect = StateEffect.define<Tooltip | null>()
    const hoverField = StateField.define<Tooltip | null>({
        create: () => null,
        update(tooltip, transactions) {
            for (const effect of transactions.effects) {
                if (effect.is(setHoverEffect)) {
                    tooltip = effect.value
                }
            }
            return tooltip
        },

        provide: field => showTooltip.from(field),
    })

    const keybindings: readonly KeyBinding[] = [
        {
            key: 'Space',
            run(view) {
                const hover = view.state.field(hoverField)
                if (hover !== null) {
                    view.dispatch({ effects: setHoverEffect.of(null) })
                    return true
                }
                const position = view.state.selection.main.from + 1
                getHoverTooltip(codeintel, blobInfo, view, position).then(
                    value => view.dispatch({ effects: setHoverEffect.of(value) }),
                    () => {}
                )
                return true
            },
        },
        {
            key: 'Enter',
            run(view) {
                const position = scipPositionAtCodemirrorPosition(view, view.state.selection.main.from)
                const atEvent = occurrenceAtPosition(view, position)
                if (!atEvent) {
                    return false
                }
                const { occurrence } = atEvent
                const cmLine = view.state.doc.line(occurrence.range.start.line + 1)
                const cmPos = cmLine.from + occurrence.range.start.character
                const rect = view.coordsAtPos(cmPos)
                const coords: Coordinates = rect ? { x: rect.left, y: rect.top } : { x: 0, y: 0 }
                const spinner = new Spinner(coords)
                goToDefinitionAtOccurrence(view, blobInfo, history, codeintel, position, occurrence, coords)
                    .then(
                        action => action(),
                        () => {}
                    )
                    .finally(() => spinner.stop())
                return true
            },
        },
        {
            key: 'Mod-ArrowRight',
            run() {
                history.goForward()
                return true
            },
        },
        {
            key: 'Mod-ArrowLeft',
            run() {
                history.goBack()
                return true
            },
        },
        {
            key: 'ArrowLeft',
            run(view) {
                view.dispatch({ effects: setHoverEffect.of(null) })
                const position = scipPositionAtCodemirrorPosition(view, view.state.selection.main.from)
                const table = view.state.facet(syntaxHighlight)
                const line = position.line
                const occurrence = closestOccurrence(line, table, position, occurrence =>
                    occurrence.range.start.isSmaller(position)
                )
                if (occurrence) {
                    selectOccurrence(codeintel, history, view, occurrence)
                }
                return true
            },
        },
        {
            key: 'ArrowRight',
            run(view) {
                view.dispatch({ effects: setHoverEffect.of(null) })
                const position = scipPositionAtCodemirrorPosition(view, view.state.selection.main.from)
                const table = view.state.facet(syntaxHighlight)
                const line = position.line
                const occurrence = closestOccurrence(line, table, position, occurrence =>
                    occurrence.range.start.isGreater(position)
                )
                if (occurrence) {
                    selectOccurrence(codeintel, history, view, occurrence)
                }
                return true
            },
        },
        {
            key: 'ArrowDown',
            run(view) {
                view.dispatch({ effects: setHoverEffect.of(null) })
                const position = scipPositionAtCodemirrorPosition(view, view.state.selection.main.from)
                const table = view.state.facet(syntaxHighlight)
                for (let line = position.line + 1; line < table.lineIndex.length; line++) {
                    const occurrence = closestOccurrence(line, table, position)
                    if (occurrence) {
                        selectOccurrence(codeintel, history, view, occurrence)
                        return true
                    }
                }
                return true
            },
        },
        {
            key: 'ArrowUp',
            run(view) {
                view.dispatch({ effects: setHoverEffect.of(null) })
                const position = scipPositionAtCodemirrorPosition(view, view.state.selection.main.from)
                const table = view.state.facet(syntaxHighlight)
                for (let line = position.line - 1; line >= 0; line--) {
                    const occurrence = closestOccurrence(line, table, position)
                    if (occurrence) {
                        selectOccurrence(codeintel, history, view, occurrence)
                        return true
                    }
                }
                return true
            },
        },
        {
            key: 'Escape',
            run(view) {
                view.dispatch({ effects: setHoverEffect.of(null) })
                view.contentDOM.blur()
                return true
            },
        },
    ]
    return [
        hoverTooltip((view, position) => getHoverTooltip(codeintel, blobInfo, view, position), {
            hoverTime: 100,
            // Hiding the tooltip when the document changes replicates
            // Monaco's behavior and also "feels right" because it removes
            // "clutter" from the input.
            hideOnChange: true,
        }),
        hoverField,
        uriFacet.of(toURIWithPath(blobInfo)),
        selectionsFacet.of(selections),
        blobInfoFacet.of(blobInfo),
        keymap.of(
            keybindings.flatMap(keybinding => {
                if (keybinding.key === 'ArrowUp') {
                    return [keybinding, { ...keybinding, key: 'k' }]
                }
                if (keybinding.key === 'ArrowDown') {
                    return [keybinding, { ...keybinding, key: 'j' }]
                }
                if (keybinding.key === 'ArrowLeft') {
                    return [keybinding, { ...keybinding, key: 'h' }, { ...keybinding, key: 'Shift-Tab' }]
                }
                if (keybinding.key === 'ArrowRight') {
                    return [keybinding, { ...keybinding, key: 'l' }, { ...keybinding, key: 'Tab' }]
                }
                return [keybinding]
            })
        ),
        EditorView.domEventHandlers({
            mouseover(event, view) {
                globalViewHack = view

                if (!codeintel) {
                    return
                }
                // toggleClickableClass(view, event.metaKey)
                goToDefinitionAtEvent(view, event, blobInfo, history, codeintel).then(
                    () => {},
                    () => {}
                )
            },
            dblclick(event, view) {
                if (!codeintel) {
                    return
                }
                const atEvent = positionAtEvent(view, event)
                if (!atEvent) {
                    return
                }
                const {
                    position: { line },
                } = atEvent
                // Select the entire line
                selectRange(view, Range.fromNumbers(line, 0, line, Number.MAX_VALUE))
                return true
            },
            click(event, view) {
                event.preventDefault()
                view.dispatch({ effects: setHoverEffect.of(null) })
                const atEvent = occurrenceAtEvent(view, event)
                if (atEvent && isInteractiveOccurrence(atEvent.occurrence)) {
                    selectOccurrence(codeintel, history, view, atEvent.occurrence)
                }
                if (!event.metaKey) {
                    return
                }
                const spinner = new Spinner({
                    x: event.clientX,
                    y: event.clientY,
                })
                goToDefinitionAtEvent(view, event, blobInfo, history, codeintel)
                    .then(
                        action => action(),
                        () => {}
                    )
                    .finally(() => spinner.stop())
            },
            contextmenu(event, view) {
                if (event.shiftKey) {
                    return
                }
                if (!codeintel) {
                    return
                }
                const atEvent = positionAtEvent(view, event)
                if (!atEvent) {
                    return
                }
                const atEvent2 = occurrenceAtEvent(view, event)
                if (atEvent2 && isInteractiveOccurrence(atEvent2.occurrence)) {
                    selectOccurrence(codeintel, history, view, atEvent2.occurrence)
                }
                const definitionAction = goToDefinitionAtEvent(view, event, blobInfo, history, codeintel)
                const { coords } = atEvent
                const menu = document.createElement('div')
                const definition = document.createElement('div')
                definition.innerHTML = 'Go to definition'
                definition.classList.add('codeintel-contextmenu-item')
                definition.classList.add('codeintel-contextmenu-item-action')

                definition.addEventListener('click', event => {
                    event.preventDefault()
                    const spinner = new Spinner(coords)
                    definitionAction
                        .then(
                            action => action(),
                            () => {}
                        )
                        .finally(() => spinner.stop())
                })
                menu.append(definition)

                const references = document.createElement('div')
                references.innerHTML = 'Find references'
                references.classList.add('codeintel-contextmenu-item')
                references.classList.add('codeintel-contextmenu-item-action')
                menu.append(references)

                const browserMenu = document.createElement('div')
                browserMenu.innerHTML = 'Browser context menu shift+right-click'
                browserMenu.classList.add('codeintel-contextmenu-item')
                menu.append(browserMenu)
                showHackyTooltip(view, menu, coords)
            },
        }),
    ]
}

interface Coordinates {
    x: number
    y: number
}

function showHackyTooltip(view: EditorView, element: HTMLElement, coords: Coordinates, clearTimeout?: number): void {
    const tooltip = document.createElement('div')
    tooltip.classList.add('codeintel-tooltip')
    tooltip.style.left = `${coords.x}px`
    tooltip.style.top = `${coords.y}px`
    tooltip.append(element)
    document.body.append(tooltip)
    let counter = 0
    const tooltipCloseListener = (): void => {
        counter += 1
        if (counter === 1) {
            return
        }
        tooltip.remove()
        document.removeEventListener('click', tooltipCloseListener)
        document.removeEventListener('contextmenu', tooltipCloseListener)
    }
    document.addEventListener('contextmenu', tooltipCloseListener)
    document.addEventListener('click', tooltipCloseListener)
    if (clearTimeout) {
        setTimeout(() => {
            tooltipCloseListener()
            tooltipCloseListener()
        }, clearTimeout)
    }
    // TODO: register up/down arrows

    // Measure and reposition after rendering first version
    requestAnimationFrame(() => {
        tooltip.style.left = `${coords.x}px`
        tooltip.style.top = `${top}px`
    })
}

async function goToDefinition(
    view: EditorView,
    history: H.History,
    codeintel: Remote<FlatExtensionHostAPI>,
    params: TextDocumentPositionParameters,
    coords: Coordinates
): Promise<() => void> {
    const definition = await codeintel.getDefinition(params)

    const result = await wrapRemoteObservable(definition).toPromise()
    if (result.isLoading) {
        return () => {}
    }
    if (result.result.length === 0) {
        return () => {
            const element = document.createElement('div')
            element.textContent = 'No definition found'
            element.style.color = 'white'
            element.style.backgroundColor = 'deepskyblue'
            showHackyTooltip(view, element, coords, 2000)
        }
    }
    for (const location of result.result) {
        if (location.uri === params.textDocument.uri && location.range && location.range) {
            const requestPosition = new Position(params.position.line, params.position.character)
            const {
                start: { line: startLine, character: startCharacter },
                end: { line: endLine, character: endCharacter },
            } = location.range
            const resultRange = Range.fromNumbers(startLine, startCharacter, endLine, endCharacter)
            if (resultRange.contains(requestPosition)) {
                return () => {
                    const element = document.createElement('div')
                    element.textContent = 'You are at the definition'
                    element.style.color = 'white'
                    element.style.backgroundColor = 'deepskyblue'
                    showHackyTooltip(view, element, coords, 2000)
                }
            }
        }
    }
    if (result.result.length === 1) {
        const { range, uri } = result.result[0]
        const { filePath, repoName, revision } = parseRepoURI(uri)
        if (filePath && range) {
            return () => {
                const selectionRange = Range.fromNumbers(
                    range.start.line,
                    range.start.character,
                    range.end.line,
                    range.end.character
                )
                if (uri === params.textDocument.uri) {
                    const atEvent = occurrenceAtPosition(view, selectionRange.start)
                    if (atEvent) {
                        selectOccurrence(codeintel, history, view, atEvent.occurrence)
                    } else {
                        selectRange(view, selectionRange)
                    }
                } else {
                    const selections = view.state.facet(selectionsFacet)
                    selections.set(uri, selectionRange)
                    const href = toPrettyBlobURL({
                        repoName,
                        revision,
                        filePath,
                        position: { line: range.start.line + 1, character: range.start.character + 1 },
                    })
                    history.push(href)
                }
            }
        }
    }
    return () => {
        // TODO: something more useful than opening ref panel
        const element = document.createElement('div')
        element.textContent = 'FIXME: Multiple definitions found'
        element.style.color = 'white'
        element.style.backgroundColor = 'deepskyblue'
        showHackyTooltip(view, element, coords, 2000)
    }
}

class Spinner {
    private spinner: HTMLElement
    constructor(coords: Coordinates) {
        this.spinner = document.createElement('div')
        this.spinner.textContent = 'loading...'
        this.spinner.style.backgroundColor = 'white'
        this.spinner.style.position = 'fixed'
        this.spinner.style.top = `${coords.y}px`
        this.spinner.style.left = `${coords.x}px`
        document.body.append(this.spinner)
    }
    public stop(): void {
        this.spinner.remove()
    }
}

async function getHoverTooltip(
    codeintel: Remote<FlatExtensionHostAPI> | undefined,
    blobInfo: BlobInfo,
    view: EditorView,
    pos: number
): Promise<Tooltip | null> {
    if (!codeintel) {
        return null
    }
    const cmLine = view.state.doc.lineAt(pos)
    const line = cmLine.number - 1
    const character = countColumn(cmLine.text, 1, pos - cmLine.from) - 1
    const atEvent = occurrenceAtPosition(view, new Position(line, character))
    if (!atEvent) {
        return null
    }
    const contents = await hoverAtOccurrence(codeintel, blobInfo, atEvent.occurrence)
    return markdownTooltip(pos, contents)
}

function hoverAtOccurrence(
    codeintel: Remote<FlatExtensionHostAPI>,
    blobInfo: BlobInfo,
    occurrence: Occurrence
): Promise<string> {
    const fromCache = hoverCache.get(occurrence)
    if (fromCache) {
        return fromCache
    }
    const uri = toURIWithPath(blobInfo)
    const contents = hoverRequest(codeintel, {
        position: { line: occurrence.range.start.line, character: occurrence.range.start.character + 1 },
        textDocument: { uri },
    })
    hoverCache.set(occurrence, contents)
    return contents
}

async function hoverRequest(
    codeintel: Remote<FlatExtensionHostAPI>,
    params: TextDocumentPositionParameters
): Promise<string> {
    const hover = await codeintel.getHover(params)

    const result = await wrapRemoteObservable(hover).toPromise()
    const contents =
        result === undefined || result.isLoading || result.result === null || result.result.contents.length === 0
            ? 'No hover information'
            : result.result.contents
                  .map(({ value }) => value)
                  .join('\n\n----\n\n')
                  .trimEnd()
    return contents
}

function markdownTooltip(pos: number, contents: string): Tooltip {
    return {
        pos,
        end: pos,
        create(): TooltipView {
            const dom = document.createElement('div')
            const markdown = renderMarkdown(contents)
            dom.innerHTML = markdown
            return { dom }
        },
    }
}
