/**
 * An experimental implementation of the Blob view using CodeMirror
 */

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { openSearchPanel } from '@codemirror/search'
import { Compartment, EditorState, Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { isEqual } from 'lodash'

import {
    addLineRangeQueryParameter,
    formatSearchParameters,
    toPositionOrRangeQueryParameter,
} from '@sourcegraph/common'
import { editorHeight, useCodeMirror } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { Shortcut } from '@sourcegraph/shared/src/react-shortcuts'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'

import { BlobInfo, BlobProps, updateBrowserHistoryIfChanged } from './Blob'
import { blobPropsFacet } from './codemirror'
import { showGitBlameDecorations } from './codemirror/blame-decorations'
import { syntaxHighlight } from './codemirror/highlight'
import { pin, updatePin } from './codemirror/hovercard'
import { selectLines, selectableLineNumbers, SelectedLineRange } from './codemirror/linenumbers'
import { search } from './codemirror/search'
import { sourcegraphExtensions } from './codemirror/sourcegraph-extensions'
import { tokensAsLinks } from './codemirror/tokens-as-links'
import { isValidLineRange } from './codemirror/utils'

const staticExtensions: Extension = [
    EditorState.readOnly.of(true),
    EditorView.editable.of(false),
    EditorView.contentAttributes.of({
        // This is required to make the blob view focusable and to make
        // triggering the in-document search (see below) work when Mod-f is
        // pressed
        tabindex: '0',
    }),
    editorHeight({ height: '100%' }),
    EditorView.theme({
        '&': {
            backgroundColor: 'var(--code-bg)',
        },
        '.cm-scroller': {
            fontFamily: 'var(--code-font-family)',
            fontSize: 'var(--code-font-size)',
            lineHeight: '1rem',
        },
        '.cm-gutters': {
            backgroundColor: 'var(--code-bg)',
            borderRight: 'initial',
        },
        '.cm-line': {
            paddingLeft: '1rem',
        },
        '.selected-line': {
            backgroundColor: 'var(--code-selection-bg)',
        },
        '.selected-line:focus': {
            boxShadow: 'none',
        },
        '.highlighted-line': {
            backgroundColor: 'var(--code-selection-bg)',
        },
    }),
    // Note that these only work out-of-the-box because the editor is
    // *focusable* by setting `tab-index: 0`.
    search,
]

// Compartments are used to reconfigure some parts of the editor without
// affecting others.

// Compartment to update various smaller settings
const settingsCompartment = new Compartment()
// Compartment to update blame information
const blameDecorationsCompartment = new Compartment()
// Compartment for propagating component props
const blobPropsCompartment = new Compartment()

export const Blob: React.FunctionComponent<BlobProps> = props => {
    const {
        className,
        wrapCode,
        isLightTheme,
        ariaLabel,
        role,
        extensionsController,
        location,
        history,
        blameHunks,
        tokenKeyboardNavigation,

        // Reference panel specific props
        disableStatusBar,
        disableDecorations,
        navigateToLineOnAnyClick,

        overrideBrowserSearchKeybinding,
        'data-testid': dataTestId,
    } = props

    const [container, setContainer] = useState<HTMLDivElement | null>(null)
    // This is used to avoid reinitializing the editor when new locations in the
    // same file are opened inside the reference panel.
    const blobInfo = useDistinctBlob(props.blobInfo)
    const position = useMemo(() => parseQueryAndHash(location.search, location.hash), [location.search, location.hash])
    const hasPin = useMemo(() => urlIsPinned(location.search), [location.search])

    const blobProps = useMemo(() => blobPropsFacet.of(props), [props])

    const settings = useMemo(
        () => [wrapCode ? EditorView.lineWrapping : [], EditorView.darkTheme.of(isLightTheme === false)],
        [wrapCode, isLightTheme]
    )

    const blameDecorations = useMemo(() => (blameHunks ? [showGitBlameDecorations.of(blameHunks)] : []), [blameHunks])

    const tokenLinks = useMemo(() => {
        if (!blobInfo.stencil) {
            return []
        }

        return blobInfo.stencil.map(range => ({
            range,
            url: `?${toPositionOrRangeQueryParameter({
                position: { line: range.start.line + 1, character: range.start.character + 1 },
            })}#tab=references`,
        }))
    }, [blobInfo.stencil])

    // Keep history and location in a ref so that we can use the latest value in
    // the onSelection callback without having to recreate it and having to
    // reconfigure the editor extensions
    const historyRef = useRef(history)
    historyRef.current = history
    const locationRef = useRef(location)
    locationRef.current = location

    const customHistoryAction = props.nav
    const onSelection = useCallback(
        (range: SelectedLineRange) => {
            const parameters = new URLSearchParams(locationRef.current.search)
            parameters.delete('popover')

            let query: string | undefined

            if (range?.line !== range?.endLine && range?.endLine) {
                query = toPositionOrRangeQueryParameter({
                    range: {
                        start: { line: range.line },
                        end: { line: range.endLine },
                    },
                })
            } else if (range?.line) {
                query = toPositionOrRangeQueryParameter({ position: { line: range.line } })
            }

            const newSearchParameters = addLineRangeQueryParameter(parameters, query)
            if (customHistoryAction) {
                customHistoryAction(
                    historyRef.current.createHref({
                        ...locationRef.current,
                        search: formatSearchParameters(newSearchParameters),
                    })
                )
            } else {
                updateBrowserHistoryIfChanged(historyRef.current, locationRef.current, newSearchParameters)
            }
        },
        [customHistoryAction]
    )

    const extensions = useMemo(
        () => [
            staticExtensions,
            selectableLineNumbers({
                onSelection,
                initialSelection: position.line !== undefined ? position : null,
                navigateToLineOnAnyClick: navigateToLineOnAnyClick ?? false,
            }),
            tokenKeyboardNavigation ? tokensAsLinks.of({ history, links: tokenLinks }) : [],
            syntaxHighlight.of(blobInfo),
            pin.init(() => (hasPin ? position : null)),
            extensionsController !== null && !navigateToLineOnAnyClick
                ? sourcegraphExtensions({
                      blobInfo,
                      initialSelection: position,
                      extensionsController,
                      disableStatusBar,
                      disableDecorations,
                  })
                : [],
            blobPropsCompartment.of(blobProps),
            blameDecorationsCompartment.of(blameDecorations),
            settingsCompartment.of(settings),
        ],
        // A couple of values are not dependencies (blameDecorations, blobProps,
        // hasPin, position and settings) because those are updated in effects
        // further below. However they are still needed here because we need to
        // set initial values when we re-initialize the editor.
        // eslint-disable-next-line react-hooks/exhaustive-deps
        [onSelection, blobInfo, extensionsController, disableStatusBar, disableDecorations, tokenLinks]
    )

    const editorRef = useRef<EditorView>()
    const editor = useCodeMirror(container, blobInfo.content, extensions, {
        updateValueOnChange: false,
        updateOnExtensionChange: false,
    })
    editorRef.current = editor

    // Reconfigure editor when blobInfo or core extensions changed
    useEffect(() => {
        if (editor) {
            // We use setState here instead of dispatching a transaction because
            // the new document has nothing to do with the previous one and so
            // any existing state should be discarded.
            editor.setState(
                EditorState.create({
                    doc: blobInfo.content,
                    extensions,
                })
            )
        }
        // editor is not provided because this should only be triggered after the
        // editor was created (i.e. not on first render)
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [blobInfo, extensions])

    // Propagate props changes to extensions
    useEffect(() => {
        if (editor) {
            editor.dispatch({ effects: blobPropsCompartment.reconfigure(blobProps) })
        }
        // editor is not provided because this should only be triggered after the
        // editor was created (i.e. not on first render)
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [blobProps])

    // Update blame information
    useEffect(() => {
        if (editor) {
            editor.dispatch({ effects: blameDecorationsCompartment.reconfigure(blameDecorations) })
        }
        // editor is not provided because this should only be triggered after the
        // editor was created (i.e. not on first render)
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [blameDecorations])

    // Update settings
    useEffect(() => {
        if (editor) {
            editor.dispatch({ effects: settingsCompartment.reconfigure(settings) })
        }
        // editor is not provided because this should only be triggered after the
        // editor was created (i.e. not on first render)
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [settings])

    // Update selected lines when URL changes
    useEffect(() => {
        if (editor) {
            selectLines(editor, position.line ? position : null)
        }
        // editor is not provided because this should only be triggered after the
        // editor was created (i.e. not on first render)
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [position])

    // Update pinned hovercard range
    useEffect(() => {
        if (editor && (!hasPin || (position.line && isValidLineRange(position, editor.state.doc)))) {
            // Only update range if position is valid inside the document.
            updatePin(editor, hasPin ? position : null)
        }
        // editor is not provided because this should only be triggered after the
        // editor was created (i.e. not on first render)
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [position, hasPin])

    const openSearch = useCallback(() => {
        if (editorRef.current) {
            openSearchPanel(editorRef.current)
        }
    }, [])

    return (
        <>
            <div
                ref={setContainer}
                aria-label={ariaLabel}
                role={role}
                data-testid={dataTestId}
                className={`${className} overflow-hidden test-editor`}
                data-editor="codemirror6"
            />
            {overrideBrowserSearchKeybinding && (
                <Shortcut ordered={['f']} held={['Mod']} onMatch={openSearch} ignoreInput={true} />
            )}
        </>
    )
}

/**
 * Returns true when the URL indicates that the hovercard at the URL position
 * should be shown on load (the hovercard is "pinned").
 */
function urlIsPinned(search: string): boolean {
    return new URLSearchParams(search).get('popover') === 'pinned'
}

/**
 * Helper hook to prevent resetting the editor view if the blob contents hasn't
 * changed.
 */
function useDistinctBlob(blobInfo: BlobInfo): BlobInfo {
    const blobRef = useRef(blobInfo)
    return useMemo(() => {
        if (!isEqual(blobRef.current, blobInfo)) {
            blobRef.current = blobInfo
        }
        return blobRef.current
    }, [blobInfo])
}
