import { Extension, RangeSetBuilder, StateEffect, StateField, Text } from '@codemirror/state'
import { Decoration, EditorView, hoverTooltip, Tooltip } from '@codemirror/view'
import { upperFirst } from 'lodash'
import { createRoot } from 'react-dom/client'
import { Subscription } from 'rxjs'
import { takeWhile } from 'rxjs/operators'

import { DocumentHighlight } from '@sourcegraph/codeintellify'
import { isErrorLike } from '@sourcegraph/common'
import { Position } from '@sourcegraph/extension-api-types'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import hoverOverlayStyle from '@sourcegraph/shared/src/hover/HoverOverlay.module.scss'
import { HoverOverlayBaseProps } from '@sourcegraph/shared/src/hover/HoverOverlay.types'
import hoverOverlayContentsStyle from '@sourcegraph/shared/src/hover/HoverOverlayContents.module.scss'
import { HoverOverlayContent } from '@sourcegraph/shared/src/hover/HoverOverlayContents/HoverOverlayContent'
import { Alert, Card, WildcardThemeContext } from '@sourcegraph/wildcard'

import { getDocumentHighlights, getHover } from '../../../backend/features'
import { BlobInfo } from '../Blob'

import { blobField } from './highlight'

import webHoverOverlayStyle from '../../../components/WebHoverOverlay/WebHoverOverlay.module.scss'

const HOVER_TIMEOUT = 50

export function codeIntel(extensionsController: ExtensionsControllerProps): Extension {
    return [documentHighlights(extensionsController), hover(extensionsController)]
}

class DocumentHighlights {
    constructor(public highlights: DocumentHighlight[], public position: number) {}
    /**
     * Returns true if there currently is a highlight at the provided position.
     */
    public isActive(position: Position): boolean {
        for (const {
            range: { start, end },
        } of this.highlights) {
            if (
                position.line >= start.line + 1 &&
                position.line <= end.line + 1 &&
                position.character >= start.character &&
                position.character <= end.character
            ) {
                return true
            }
        }
        return false
    }
}

/**
 * Show document highlights.
 */
function documentHighlights(extensionsController: ExtensionsControllerProps): Extension {
    const highlightDecoration = Decoration.mark({ class: 'sourcegraph-document-highlight' })
    const setHighlights = StateEffect.define<DocumentHighlights | null>()
    const highlightsState = StateField.define<DocumentHighlights | null>({
        create() {
            return null
        },

        update(value, transaction) {
            if (transaction.docChanged) {
                // We don't have to worry about mapping highlights across
                // document changes since a change means a new document is
                // loaded.
                return null
            }

            for (const effect of transaction.effects) {
                if (effect.is(setHighlights)) {
                    return effect.value
                }
            }

            return value
        },
    })

    // Keeps track of the currently hovered position. Extension requests are
    // throttled and this helps to get the latest position (or null if the last
    // hover wasn't over a word).
    let lastHoverPosition: {
        position: Position
        offset: number
    } | null = null
    // getDocumentHighlights doesn't seem to complete so we need to unsubscribe
    // manually.
    let lastSubscription: Subscription | null = null

    const scheduleQuery = throttle((options, view: EditorView, blobInfo: BlobInfo) => {
        if (lastSubscription) {
            lastSubscription.unsubscribe()
        }

        if (lastHoverPosition) {
            const { position, offset } = lastHoverPosition

            lastSubscription = getDocumentHighlights(
                {
                    position,
                    commitID: blobInfo.commitID,
                    filePath: blobInfo.filePath,
                    repoName: blobInfo.repoName,
                },
                extensionsController
            )
                .pipe(takeWhile(() => offset === lastHoverPosition?.offset && !options.isCanceled))
                .subscribe(highlights => {
                    if (highlights.length === 0 && !view.state.field(highlightsState)) {
                        // No need to schedule a transaction if the state is already
                        // empty anyway.
                        return
                    }
                    view.dispatch({
                        effects: setHighlights.of(
                            highlights.length > 0 ? new DocumentHighlights(highlights, offset) : null
                        ),
                    })
                })
        }
    }, HOVER_TIMEOUT)

    function clearHighlights(view: EditorView): void {
        lastHoverPosition = null

        if (view.state.field(highlightsState)) {
            view.dispatch({ effects: setHighlights.of(null) })
        }
    }

    return [
        highlightsState,
        EditorView.domEventHandlers({
            mousemove(event: MouseEvent, view: EditorView) {
                const offset = view.posAtCoords(event)
                if (offset === null) {
                    clearHighlights(view)
                    return
                }

                // Do not query if there isn't a word at this position
                if (!view.state.wordAt(offset)) {
                    clearHighlights(view)
                    return
                }

                const blobInfo = view.state.field(blobField)
                if (!blobInfo) {
                    clearHighlights(view)
                    return
                }

                const line = view.state.doc.lineAt(offset)
                const position = {
                    line: line.number,
                    character: Math.max(offset - line.from, 1),
                }

                {
                    // Do not query if we already have highlighting information
                    // for the current position.
                    const currentHighlights = view.state.field(highlightsState)
                    if (currentHighlights?.isActive(position)) {
                        lastHoverPosition = null
                        return
                    }
                }

                lastHoverPosition = { position, offset }
                scheduleQuery(view, blobInfo)
            },
        }),
        EditorView.decorations.from(highlightsState, documentHighlights => {
            let decorations = Decoration.none

            return view => {
                if (documentHighlights) {
                    const builder = new RangeSetBuilder<Decoration>()

                    // Most of the time number of highlights is small and close
                    // together so it's likely ok to iterate over all them and
                    // not just the ones in the current viewport.
                    for (const highlight of documentHighlights.highlights) {
                        builder.add(
                            positionToOffset(view.state.doc, highlight.range.start),
                            positionToOffset(view.state.doc, highlight.range.end),
                            highlightDecoration
                        )
                    }

                    decorations = builder.finish()
                } else {
                    decorations = Decoration.none
                }

                return decorations
            }
        }),
    ]
}

/**
 * Show hovercard.
 */
function hover(extensionsController: ExtensionsControllerProps): Extension {
    return [
        // hoverTooltip takes care of only calling the source (and processing
        // its return value) when necessary.
        hoverTooltip(
            (view, position) => {
                const blobInfo = view.state.field(blobField)
                if (!blobInfo) {
                    return null
                }

                const line = view.state.doc.lineAt(position)
                const column = Math.max(position - line.from, 1)

                return new Promise(resolve => {
                    const subscription = getHover(
                        {
                            position: { line: line.number, character: column },
                            commitID: blobInfo.commitID,
                            filePath: blobInfo.filePath,
                            repoName: blobInfo.repoName,
                        },
                        extensionsController
                    ).subscribe(({ isLoading, result }) => {
                        if (isLoading === false) {
                            // It looks like the observable returned by getHover
                            // never completes so we unsubscribe manually
                            subscription.unsubscribe()

                            if (!result) {
                                return resolve(null)
                            }

                            // Try to align the tooltip with the token start,
                            // falling back to CodeMirror's logic to find a word
                            // boundary or the cursor position.
                            let start = position
                            let end = position

                            if (result.range) {
                                start = positionToOffset(view.state.doc, result.range.start)
                                end = positionToOffset(view.state.doc, result.range.end)
                            } else {
                                const word = view.state.wordAt(position)
                                if (word) {
                                    start = word.from
                                    end = word.from
                                }
                            }

                            resolve(reactTooltip(start, end, <Hovercard hoverOrError={result} />))
                        }
                    })
                })
            },
            {
                hoverTime: HOVER_TIMEOUT,
            }
        ),
        EditorView.theme({
            '.cm-code-intel-hovercard': {
                fontFamily: 'sans-serif',
                minWidth: '10rem',
            },
            '.cm-code-intel-contents': {
                maxHeight: '25rem',
            },
            '.cm-code-intel-content': {
                maxHeight: '25rem',
            },
            '.cm-tooltip': {
                border: 'initial',
                backgroundColor: 'initial',
            },
        }),
    ]
}

/**
 * Converts line/character positions to document offsets.
 */
function positionToOffset(textDocument: Text, position: Position): number {
    // Position seems 0 based
    return textDocument.line(position.line + 1).from + position.character
}

/**
 * Helper function to generate a Tooltip with React content. No idea whether
 * that's a good way to do it.
 */
function reactTooltip(start: number, end: number, element: React.ReactElement): Tooltip | null {
    return {
        above: true,
        pos: start,
        end,
        create() {
            const container = document.createElement('div')
            const root = createRoot(container)
            root.render(element)
            return {
                dom: container,
                overlap: true,
            }
        },
    }
}

/**
 * A simple replication of the hovercard for the old blob view.
 */
const Hovercard: React.FunctionComponent<Pick<HoverOverlayBaseProps, 'hoverOrError'>> = ({ hoverOrError }) => {
    if (isErrorLike(hoverOrError)) {
        return <Alert className={hoverOverlayStyle.hoverError}>{upperFirst(hoverOrError.message)}</Alert>
    }

    if (hoverOrError === undefined || hoverOrError === 'loading') {
        return null
    }

    if (hoverOrError === null || (hoverOrError.contents.length === 0 && hoverOrError.alerts?.length)) {
        return (
            // Show some content to give the close button space and communicate to the user we couldn't find a hover.
            <small className={hoverOverlayStyle.hoverEmpty}>No hover information available.</small>
        )
    }

    return (
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            <Card
                className={`${hoverOverlayStyle.card} ${webHoverOverlayStyle.webHoverOverlay} cm-code-intel-hovercard`}
            >
                <div className={`${hoverOverlayContentsStyle.hoverOverlayContents} cm-code-intel-contents`}>
                    {hoverOrError.contents.map((content, index) => (
                        <HoverOverlayContent
                            key={index}
                            index={index}
                            content={content}
                            aggregatedBadges={hoverOrError.aggregatedBadges}
                            contentClassName="cm-code-intel-content"
                        />
                    ))}
                </div>
            </Card>
        </WildcardThemeContext.Provider>
    )
}

/**
 * This is a trailing throttle implementation that allows the callback to query
 * whether it was canceled or not (i.e. whether or not the function was called
 * again). This can be useful if the callback performs asynchronous work.
 */
function throttle<P extends unknown[]>(
    callback: ({ isCanceled }: { isCanceled: boolean }, ...args: P) => void,
    timeout: number
): (...args: P) => void {
    let running = false
    let lastTimeCalled = 0
    let lastArguments: P

    return (...args) => {
        lastArguments = args

        if (!running) {
            running = true
            const timeCalled = (lastTimeCalled = Date.now())
            setTimeout(() => {
                running = false
                callback(
                    {
                        get isCanceled() {
                            return timeCalled !== lastTimeCalled
                        },
                    },
                    ...lastArguments
                )
            }, timeout)
        }
    }
}
