/**
 * An experimental implementation of the Blob view using CodeMirror
 */

import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'

import { openSearchPanel } from '@codemirror/search'
import { Compartment, EditorState, Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { isEqual } from 'lodash'
import { createPath, useLocation, useNavigate } from 'react-router-dom-v5-compat'

import {
    addLineRangeQueryParameter,
    formatSearchParameters,
    toPositionOrRangeQueryParameter,
} from '@sourcegraph/common'
import { editorHeight, useCodeMirror } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { Shortcut } from '@sourcegraph/shared/src/react-shortcuts'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'
import { useLocalStorage } from '@sourcegraph/wildcard'

import { useExperimentalFeatures } from '../../stores'

import { BlobInfo, BlobProps, updateBrowserHistoryIfChanged } from './Blob'
import { blobPropsFacet } from './codemirror'
import { createBlameDecorationsExtension } from './codemirror/blame-decorations'
import { codeFoldingExtension } from './codemirror/code-folding'
import { syntaxHighlight } from './codemirror/highlight'
import { pin, updatePin } from './codemirror/hovercard'
import { selectableLineNumbers, SelectedLineRange, selectLines } from './codemirror/linenumbers'
import { lockFirstVisibleLine } from './codemirror/lock-line'
import { navigateToLineOnAnyClickExtension } from './codemirror/navigate-to-any-line-on-click'
import { occurrenceAtPosition, positionAtCmPosition } from './codemirror/occurrence-utils'
import { search } from './codemirror/search'
import { sourcegraphExtensions } from './codemirror/sourcegraph-extensions'
import { selectOccurrence } from './codemirror/token-selection/code-intel-tooltips'
import { tokenSelectionExtension } from './codemirror/token-selection/extension'
import { selectionFromLocation } from './codemirror/token-selection/selections'
import { tokensAsLinks } from './codemirror/tokens-as-links'
import { isValidLineRange } from './codemirror/utils'
import { setBlobEditView } from './use-blob-store'

const staticExtensions: Extension = [
    EditorState.readOnly.of(true),
    EditorView.editable.of(false),
    EditorView.contentAttributes.of({
        // This is required to make the blob view focusable and to make
        // triggering the in-document search (see below) work when Mod-f is
        // pressed
        tabindex: '0',
        // CodeMirror defaults to role="textbox" which does not produce the
        // desired screenreader behavior we want for this component.
        // See https://github.com/sourcegraph/sourcegraph/issues/43375
        role: 'generic',
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
]

// Compartments are used to reconfigure some parts of the editor without
// affecting others.

// Compartment to update various smaller settings
const settingsCompartment = new Compartment()
// Compartment to update blame decorations
const blameDecorationsCompartment = new Compartment()
// Compartment for propagating component props
const blobPropsCompartment = new Compartment()
// Compartment for line wrapping.
const wrapCodeCompartment = new Compartment()

export const Blob: React.FunctionComponent<BlobProps> = props => {
    const {
        className,
        wrapCode,
        isLightTheme,
        ariaLabel,
        role,
        extensionsController,
        isBlameVisible,
        blameHunks,
        enableLinkDrivenCodeNavigation,
        enableSelectionDrivenCodeNavigation,

        // Reference panel specific props
        navigateToLineOnAnyClick,

        overrideBrowserSearchKeybinding,
        'data-testid': dataTestId,
    } = props

    const location = useLocation()
    const navigate = useNavigate()
    const [useFileSearch, setUseFileSearch] = useLocalStorage('blob.overrideBrowserFindOnPage', true)

    const [container, setContainer] = useState<HTMLDivElement | null>(null)
    // This is used to avoid reinitializing the editor when new locations in the
    // same file are opened inside the reference panel.
    const blobInfo = useDistinctBlob(props.blobInfo)
    const position = useMemo(() => parseQueryAndHash(location.search, location.hash), [location.search, location.hash])
    const hasPin = useMemo(() => urlIsPinned(location.search), [location.search])

    const blobProps = useMemo(() => blobPropsFacet.of(props), [props])

    const themeSettings = useMemo(() => EditorView.darkTheme.of(isLightTheme === false), [isLightTheme])
    const wrapCodeSettings = useMemo<Extension>(() => (wrapCode ? EditorView.lineWrapping : []), [wrapCode])

    const blameDecorations = useMemo(
        () => createBlameDecorationsExtension(!!isBlameVisible, blameHunks, isLightTheme),
        [isBlameVisible, blameHunks, isLightTheme]
    )

    const preloadGoToDefinition = useExperimentalFeatures(features => features.preloadGoToDefinition ?? false)

    // Keep history and location in a ref so that we can use the latest value in
    // the onSelection callback without having to recreate it and having to
    // reconfigure the editor extensions
    const navigateRef = useRef(navigate)
    navigateRef.current = navigate
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
                    createPath({
                        ...locationRef.current,
                        search: formatSearchParameters(newSearchParameters),
                    })
                )
            } else {
                updateBrowserHistoryIfChanged(navigateRef.current, locationRef.current, newSearchParameters)
            }
        },
        [customHistoryAction]
    )
    const extensions = useMemo(
        () => [
            // Log uncaught errors that happen in callbacks that we pass to
            // CodeMirror. Without this exception sink, exceptions get silently
            // ignored making it difficult to debug issues caused by uncaught
            // exceptions.
            // eslint-disable-next-line no-console
            EditorView.exceptionSink.of(exception => console.log(exception)),
            staticExtensions,
            selectableLineNumbers({
                onSelection,
                initialSelection: position.line !== undefined ? position : null,
                navigateToLineOnAnyClick: navigateToLineOnAnyClick ?? false,
                enableSelectionDrivenCodeNavigation,
            }),
            enableSelectionDrivenCodeNavigation ? tokenSelectionExtension() : [],
            enableLinkDrivenCodeNavigation
                ? tokensAsLinks({ navigate: navigateRef.current, blobInfo, preloadGoToDefinition })
                : [],
            syntaxHighlight.of(blobInfo),
            pin.init(() => (hasPin ? position : null)),
            extensionsController !== null && !navigateToLineOnAnyClick
                ? sourcegraphExtensions({
                      blobInfo,
                      initialSelection: position,
                      extensionsController,
                      enableSelectionDrivenCodeNavigation,
                  })
                : [],
            blobPropsCompartment.of(blobProps),
            blameDecorationsCompartment.of(blameDecorations),
            navigateToLineOnAnyClick ? navigateToLineOnAnyClickExtension : [],
            settingsCompartment.of(themeSettings),
            wrapCodeCompartment.of(wrapCodeSettings),
            search({
                // useFileSearch is not a dependency because the search
                // extension manages its own state. This is just the initial
                // value
                overrideBrowserFindInPageShortcut: useFileSearch,
                onOverrideBrowserFindInPageToggle: setUseFileSearch,
            }),
            codeFoldingExtension(),
        ],
        // A couple of values are not dependencies (blameDecorations, blobProps,
        // hasPin, position and settings) because those are updated in effects
        // further below. However, they are still needed here because we need to
        // set initial values when we re-initialize the editor.
        // eslint-disable-next-line react-hooks/exhaustive-deps
        [onSelection, blobInfo, extensionsController]
    )

    const editorRef = useRef<EditorView>()
    const editor = useCodeMirror(container, blobInfo.content, extensions, {
        updateValueOnChange: false,
        updateOnExtensionChange: false,
    })
    editorRef.current = editor

    // Sync editor store with global Zustand store API
    useEffect(() => setBlobEditView(editor ?? null), [editor])

    // Reconfigure editor when blobInfo or core extensions changed
    useEffect(() => {
        if (editor) {
            // We use setState here instead of dispatching a transaction because
            // the new document has nothing to do with the previous one and so
            // any existing state should be discarded.
            const state = EditorState.create({ doc: blobInfo.content, extensions })
            editor.setState(state)

            if (!enableSelectionDrivenCodeNavigation) {
                return
            }

            // Sync editor selection/focus with the URL so that triggering
            // `history.goBack/goForward()` works similar to the "Go back"
            // command in VS Code.
            const { selection } = selectionFromLocation(editor, locationRef.current)
            if (selection) {
                const position = positionAtCmPosition(editor, selection.from)
                const occurrence = occurrenceAtPosition(editor.state, position)
                if (occurrence) {
                    selectOccurrence(editor, occurrence)
                    // Automatically focus the content DOM to enable keyboard
                    // navigation. Without this automatic focus, users need to click
                    // on the blob view with the mouse.
                    // NOTE: this focus statment does not seem to have an effect
                    // when using macOS VoiceOver.
                    editor.contentDOM.focus({ preventScroll: true })
                }
            }
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

    // Update blame decorations
    useLayoutEffect(() => {
        if (editor) {
            const effects = [blameDecorationsCompartment.reconfigure(blameDecorations), ...lockFirstVisibleLine(editor)]
            editor.dispatch({ effects })
        }
        // editor is not provided because this should only be triggered after the
        // editor was created (i.e. not on first render)
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [blameDecorations])

    // Update theme
    useEffect(() => {
        if (editor) {
            editor.dispatch({ effects: settingsCompartment.reconfigure(themeSettings) })
        }
        // editor is not provided because this should only be triggered after the
        // editor was created (i.e. not on first render)
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [themeSettings])

    // Update line wrapping
    useEffect(() => {
        if (editor) {
            const effects = [wrapCodeCompartment.reconfigure(wrapCodeSettings), ...lockFirstVisibleLine(editor)]
            editor.dispatch({ effects })
        }
        // editor is not provided because this should only be triggered after the
        // editor was created (i.e. not on first render)
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [wrapCodeSettings])

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
            {overrideBrowserSearchKeybinding && useFileSearch && (
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
