import { Extension, Facet } from '@codemirror/state'
import { EditorView, hoverTooltip, repositionTooltips, ViewUpdate } from '@codemirror/view'
import { createRoot, Root } from 'react-dom/client'

import { addLineRangeQueryParameter, isErrorLike, toPositionOrRangeQueryParameter } from '@sourcegraph/common'
import { HoverOverlayBaseProps } from '@sourcegraph/shared/src/hover/HoverOverlay.types'

import { Position } from '@sourcegraph/extension-api-types'
import { WebHoverOverlay, WebHoverOverlayProps } from '../../../components/WebHoverOverlay'
import { blobPropsFacet } from '.'
import { combineLatest, Observable, Subject } from 'rxjs'
import { startWith } from 'rxjs/operators'
import { BlobProps, updateBrowserHistoryIfChanged } from '../Blob'
import { Container } from './react-interop'
import webOverlayStyles from '../../../components/WebHoverOverlay/WebHoverOverlay.module.scss'

type HovercardSource = (
    view: EditorView,
    position: Position
) => Observable<Pick<HoverOverlayBaseProps, 'hoverOrError' | 'actionsOrError'>>

const HOVER_TIMEOUT = 50

const hoverCardTheme = EditorView.theme({
    '.cm-code-intel-hovercard': {
        // Without this all text in the hovercard is monospace
        fontFamily: 'sans-serif',
    },
    [`.${webOverlayStyles.webHoverOverlay}`]: {
        // This is normally "position: 'absolute'". CodeMirror does the
        // positioning. Without this CodeMirror thinks the hover content is
        // empty.
        position: 'initial !important',
    },
    '.cm-tooltip': {
        // Reset CodeMirror's default style
        border: 'initial',
        backgroundColor: 'initial',
    },
})

// WebHoverOverlay requires to be passed an element representing the currently
// hovered token.  Since we don't have/want that for CodeMirror we are passing a
// dummy element.
const dummyHoveredElement = document.createElement('span')
// WebHoverOverlay expects to be passed the overlay position. Since CodeMirror
// positions the element we always use the same value.
const dummyOverlayPosition = { left: 0, bottom: 0 }

/**
 * Facet with which an extension can provide a hovercard source. For simplicity
 * only one source can be provided, others are ignored (in practice there is
 * only one source at the moment anyway).
 */
export const hovercardSource = Facet.define<HovercardSource, HovercardSource>({
    combine: sources => sources[0],
    enables: facet => hovercard(facet),
})

/**
 * Registers an extension to show a hovercard
 */
export function hovercard(facet: Facet<HovercardSource, HovercardSource>): Extension {
    return [
        // hoverTooltip takes care of only calling the source (and processing
        // its return value) when necessary.
        hoverTooltip(
            async (view, offset) => {
                const line = view.state.doc.lineAt(offset)

                // Find enclosing word/token. This is necessary otherwise
                // the extension host cannot determine whether we are the
                // definition or not.
                let start = offset
                let end = offset

                {
                    const word = view.state.wordAt(offset)
                    if (word) {
                        start = word.from
                        end = word.to
                    }
                }

                const character = Math.max(start - line.from + 1, 1)
                const position = { character, line: line.number }
                const nextRoot = new Subject<Root>()
                const nextProps = new Subject<BlobProps>()

                combineLatest([
                    nextRoot,
                    view.state.facet(facet)(view, position),
                    nextProps.pipe(startWith(view.state.facet(blobPropsFacet))),
                ]).subscribe(([root, { hoverOrError, actionsOrError }, props]) => {
                    let hoverContext = {
                        commitID: props.blobInfo.commitID,
                        filePath: props.blobInfo.filePath,
                        repoName: props.blobInfo.repoName,
                        revision: props.blobInfo.revision,
                    }

                    let hoveredToken: WebHoverOverlayProps['hoveredToken'] = {
                        ...hoverContext,
                        ...position,
                    }

                    if (
                        hoverOrError &&
                        hoverOrError !== 'loading' &&
                        !isErrorLike(hoverOrError) &&
                        hoverOrError.range
                    ) {
                        hoveredToken = {
                            ...hoveredToken,
                            line: hoverOrError.range.start.line + 1,
                            character: hoverOrError.range.start.character + 1,
                        }
                    }

                    if (hoverOrError) {
                        root.render(
                            <Container onRender={() => repositionTooltips(view)} history={props.history}>
                                <div className="cm-code-intel-hovercard">
                                    <WebHoverOverlay
                                        // Blob props
                                        location={props.location}
                                        onHoverShown={props.onHoverShown}
                                        isLightTheme={props.isLightTheme}
                                        platformContext={props.platformContext}
                                        settingsCascade={props.settingsCascade}
                                        telemetryService={props.telemetryService}
                                        extensionsController={props.extensionsController}
                                        nav={props.nav ?? (url => props.history.push(url))}
                                        // Hover props
                                        actionsOrError={actionsOrError}
                                        hoverOrError={hoverOrError}
                                        // CodeMirror handles the positioning but a
                                        // non-nullable value must be passed for the
                                        // hovercard to render
                                        overlayPosition={dummyOverlayPosition}
                                        hoveredToken={hoveredToken}
                                        hoveredTokenElement={dummyHoveredElement}
                                        pinOptions={{
                                            showCloseButton: true,
                                            onCloseButtonClick: () => {
                                                const parameters = new URLSearchParams(props.location.search)
                                                parameters.delete('popover')

                                                updateBrowserHistoryIfChanged(props.history, props.location, parameters)
                                            },
                                            onCopyLinkButtonClick: async () => {
                                                const range = { start: position, end: position }
                                                const context = { position, range }
                                                const search = new URLSearchParams(location.search)
                                                search.set('popover', 'pinned')
                                                updateBrowserHistoryIfChanged(
                                                    props.history,
                                                    props.location,
                                                    addLineRangeQueryParameter(
                                                        search,
                                                        toPositionOrRangeQueryParameter(context)
                                                    )
                                                )
                                                await navigator.clipboard.writeText(window.location.href)
                                            },
                                        }}
                                    />
                                </div>
                            </Container>
                        )
                    } else {
                        root.render([])
                    }
                })

                return {
                    pos: start,
                    end,
                    above: true,
                    create() {
                        const container = document.createElement('div')
                        return {
                            dom: container,
                            overlap: true,
                            mount() {
                                nextRoot.next(createRoot(container))
                            },
                            update(viewUpdate: ViewUpdate) {
                                if (
                                    viewUpdate.startState.facet(blobPropsFacet) !==
                                    viewUpdate.state.facet(blobPropsFacet)
                                ) {
                                    nextProps.next(viewUpdate.state.facet(blobPropsFacet))
                                }
                            },
                        }
                    },
                }
            },
            {
                hoverTime: HOVER_TIMEOUT,
            }
        ),
        hoverCardTheme,
    ]
}
