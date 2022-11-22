import { Extension, StateField } from '@codemirror/state'
import { EditorView, hoverTooltip, keymap, logException } from '@codemirror/view'
import { Remote } from 'comlink'
import * as H from 'history'

import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { Occurrence, Range } from '@sourcegraph/shared/src/codeintel/scip'
import { toURIWithPath } from '@sourcegraph/shared/src/util/url'

import { BlobInfo } from '../Blob'

import { isInteractiveOccurrence } from './tokens-as-links'
import { cmdClickFacet } from './cmd-click'
import { hasFindImplementationsSupport } from '@sourcegraph/shared/src/codeintel/api'
import { Spinner } from './Spinner'
import { occurrenceAtEvent, positionAtEvent } from './positions'
import { getHoverTooltip, hoverCache, hoverField, setHoverEffect } from './textdocument-hover'
import {
    blobInfoFacet,
    codeintelFacet,
    historyFacet,
    selectionsFacet,
    textDocumentImplemenationSupport,
    uriFacet,
} from './textdocument-facets'
import { tokenSelectionKeyBindings } from './token-selection-keymap'
import { selectedOccurrence, selectOccurrence, selectRange, tokenSelectionTheme } from './textdocument-selections'
import { definitionCache, goToDefinitionAtEvent } from './textdocument-definition'

export function tokenSelectionExtension(
    codeintel: Remote<FlatExtensionHostAPI>,
    blobInfo: BlobInfo,
    history: H.History,
    selections: Map<string, Range>
): Extension {
    return [
        cmdClickFacet.of(false),
        tokenSelectionTheme,
        selectedOccurrence.of(null),
        historyFacet.of(history),
        codeintelFacet.of(codeintel),
        // definitionCache,
        hoverCache,
        hoverTooltip((view, position) => getHoverTooltip(view, position), { hoverTime: 300, hideOnChange: true }),
        hoverField,
        uriFacet.of(toURIWithPath(blobInfo)),
        textDocumentImplemenationSupport.of(hasFindImplementationsSupport(blobInfo.mode)),
        selectionsFacet.of(selections),
        blobInfoFacet.of(blobInfo),
        keymap.of(tokenSelectionKeyBindings),
        EditorView.domEventHandlers({
            mouseover(event, view) {
                logOnError(view, () =>
                    goToDefinitionAtEvent(view, event).then(
                        () => {},
                        () => {}
                    )
                )
            },
            dblclick(event, view) {
                logOnError(view, () => {
                    const atEvent = positionAtEvent(view, event)
                    if (!atEvent) {
                        return
                    }
                    const {
                        position: { line },
                    } = atEvent
                    // Select the entire line
                    selectRange(view, Range.fromNumbers(line, 0, line, Number.MAX_VALUE))
                })
            },
            click(event, view) {
                logOnError(view, () => {
                    event.preventDefault()
                    view.dispatch({ effects: setHoverEffect.of(null) })
                    const atEvent = occurrenceAtEvent(view, event)
                    if (atEvent && isInteractiveOccurrence(atEvent.occurrence)) {
                        selectOccurrence(view, atEvent.occurrence)
                    }
                    if (!event.metaKey) {
                        return
                    }
                    const spinner = new Spinner(view, view.posAtCoords({ x: event.clientX, y: event.clientY }))
                    goToDefinitionAtEvent(view, event)
                        .then(
                            action => action(),
                            () => {}
                        )
                        .finally(() => spinner.stop())
                })
            },
        }),
    ]
}

function logOnError(view: EditorView, handler: () => void): void {
    try {
        handler()
    } catch (error) {
        logException(view.state, error)
    }
}
