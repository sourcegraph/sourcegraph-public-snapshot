/**
 * An experimental implementation of the Blob view using CodeMirror
 */

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { search, searchKeymap } from '@codemirror/search'
import { EditorState, Extension } from '@codemirror/state'
import { EditorView, keymap } from '@codemirror/view'

import { addLineRangeQueryParameter, toPositionOrRangeQueryParameter } from '@sourcegraph/common'
import {
    createUpdateableField,
    editorHeight,
    useCodeMirror,
    useCompartment,
} from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { parseQueryAndHash, toURIWithPath } from '@sourcegraph/shared/src/util/url'

import { enableExtensionsDecorationsColumnViewFromSettings } from '../../util/settings'

import { blameDecorationType, BlobInfo, BlobProps, updateBrowserHistoryIfChanged } from './Blob'
import {
    enableExtensionsDecorationsColumnView as enableColumnView,
    showTextDocumentDecorations,
} from './codemirror/decorations'
import { syntaxHighlight } from './codemirror/highlight'
import { selectLines, selectableLineNumbers, SelectedLineRange } from './codemirror/linenumbers'
import { sourcegraphExtensions } from './codemirror/sourcegraph-extensions'
import { blobPropsFacet, hovercardRangeFromPin } from './codemirror'
import { hovercardRanges } from './codemirror/hovercard'

const staticExtensions: Extension = [
    // Using EditorState.readOnly instead of EditorView.editable allows us to
    // focus the editor and placing a text cursor
    EditorState.readOnly.of(true),
    editorHeight({ height: '100%' }),
    EditorView.theme({
        '&': {
            fontFamily: 'var(--code-font-family)',
            fontSize: 'var(--code-font-size)',
            backgroundColor: 'var(--code-bg)',
        },
        '.selected-line': {
            backgroundColor: 'var(--code-selection-bg)',
        },
        '.cm-gutters': {
            backgroundColor: 'initial',
            borderRight: 'initial',
        },
    }),
    // Note that these only work out-of-the-box because the editor is
    // *focusable* but read-only (see EditorState.readOnly above).
    search({ top: true }),
    keymap.of(searchKeymap),
]

export const Blob: React.FunctionComponent<BlobProps> = props => {
    const {
        className,
        blobInfo,
        wrapCode,
        isLightTheme,
        ariaLabel,
        role,
        extensionsController,
        settingsCascade,
        location,
        history,
        blameDecorations,

        // These props don't have to be supported yet because the CodeMirror blob
        // view is only used on the blob page where these are always true
        // disableStatusBar
        // disableDecorations
    } = props

    const [container, setContainer] = useState<HTMLDivElement | null>(null)
    const position = useMemo(() => parseQueryAndHash(location.search, location.hash), [location.search, location.hash])
    const blobIsLoading = useBlobIsLoading(blobInfo, location.pathname)

    const enableExtensionsDecorationsColumnView = enableExtensionsDecorationsColumnViewFromSettings(settingsCascade)

    const settings = useMemo(
        () => [
            wrapCode ? EditorView.lineWrapping : [],
            EditorView.darkTheme.of(isLightTheme === false),
            blameDecorations
                ? [
                      // Force column view if blameDecorations is set
                      enableColumnView.of(true),
                      showTextDocumentDecorations.of([[blameDecorationType, blameDecorations]]),
                  ]
                : [],
            enableColumnView.of(enableExtensionsDecorationsColumnView),
        ],
        [wrapCode, isLightTheme, location, enableExtensionsDecorationsColumnView, blameDecorations]
    )
    const [settingsCompartment, updateSettingsCompartment] = useCompartment(settings)

    // Used to render pinned hovercards
    const [pinnedRangeField, updatePinnedRangeField] = useMemo(() => {
        return createUpdateableField(urlIsPinned(location.search) ? position : null, field =>
            hovercardRangeFromPin(field)
        )
    }, [blobInfo])

    // Keep history and location in a ref so that we can use the latest value in
    // the onSelection callback without having to recreate it and having to
    // reconfigure the editor extensions
    const historyRef = useRef(history)
    historyRef.current = history
    const locationRef = useRef(location)
    locationRef.current = location

    const onSelection = useCallback((range: SelectedLineRange) => {
        const parameters = new URLSearchParams(locationRef.current.search)
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

        updateBrowserHistoryIfChanged(
            historyRef.current,
            locationRef.current,
            addLineRangeQueryParameter(parameters, query)
        )
    }, [])

    const [propsField, updatePropsField] = useMemo(
        () => createUpdateableField(props, field => blobPropsFacet.from(field)),
        // Should only be executed on first render
        // eslint-disable-next-line react-hooks/exhaustive-deps
        []
    )
    const extensions = useMemo(
        () => [
            selectableLineNumbers({ onSelection }),
            staticExtensions,
            settingsCompartment,
            syntaxHighlight.of(blobInfo),
            sourcegraphExtensions({ blobInfo, extensionsController }),
            propsField,
            pinnedRangeField,
        ],
        [propsField, settingsCompartment, onSelection, blobInfo, extensionsController]
    )

    const editor = useCodeMirror(container, blobInfo.content, extensions, {
        updateValueOnChange: false,
        updateOnExtensionChange: false,
    })

    useEffect(() => {
        if (editor) {
            updatePropsField(editor, props)
        }
        // editor is not provided because this should only be triggered after the
        // editor was created (i.e. not on first render)
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [props])

    useEffect(() => {
        if (editor) {
            updateSettingsCompartment(editor, settings)
        }
        // editor is not provided because this should only be triggered after the
        // editor was created (i.e. not on first render)
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [updateSettingsCompartment, settings])

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

    // Update selected lines when URL changes
    useEffect(() => {
        if (editor && !blobIsLoading) {
            selectLines(editor, position.line ? position : null)
        }
        // blobInfo isn't used but we need to trigger the line selection and focus
        // logic whenever the content changes
    }, [editor, position, blobIsLoading])

    // Update pinned hovercard range
    const hasPin = useMemo(() => urlIsPinned(location.search), [location.search])
    useEffect(() => {
        if (editor && !blobIsLoading) {
            updatePinnedRangeField(editor, hasPin ? position : null)
        }
        // blobInfo isn't used but we need to trigger the line selection and focus
        // logic whenever the content changes
    }, [position, hasPin, blobIsLoading])

    return <div ref={setContainer} aria-label={ariaLabel} role={role} className={`${className} overflow-hidden`} />
}

function urlIsPinned(search: string): boolean {
    return new URLSearchParams(search).get('popover') === 'pinned'
}

/**
 * Because location changes before new blob info is available we often apply
 * updates to the old document, which can be problematic or thorw errors. This
 * helper hook observers keeps track of blob info and location changes to
 * determine whether or not to apply updates.
 */
function useBlobIsLoading(blobInfo: BlobInfo, pathname: string): boolean {
    return pathname !== useMemo(() => pathname, [blobInfo])
}
