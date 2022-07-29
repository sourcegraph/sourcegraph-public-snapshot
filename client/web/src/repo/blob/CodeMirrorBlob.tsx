/**
 * An experimental implementation of the Blob view using CodeMirror
 */

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { EditorState, Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'

import { addLineRangeQueryParameter, toPositionOrRangeQueryParameter } from '@sourcegraph/common'
import { editorHeight, useCodeMirror, useCompartment } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'

import { enableExtensionsDecorationsColumnViewFromSettings } from '../../util/settings'

import { blameDecorationType, BlobProps, updateBrowserHistoryIfChanged } from './Blob'
import { locationField, updateLocation } from './codemirror'
import {
    enableExtensionsDecorationsColumnView as enableColumnView,
    showTextDocumentDecorations,
} from './codemirror/decorations'
import { syntaxHighlight } from './codemirror/highlight'
import { selectLines, selectableLineNumbers, SelectedLineRange } from './codemirror/linenumbers'
import { sourcegraphExtensions } from './codemirror/sourcegraph-extensions'

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
]

export const Blob: React.FunctionComponent<BlobProps> = ({
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
}) => {
    const [container, setContainer] = useState<HTMLDivElement | null>(null)

    const enableExtensionsDecorationsColumnView = enableExtensionsDecorationsColumnViewFromSettings(settingsCascade)

    const settings = useMemo(
        () => [
            wrapCode ? EditorView.lineWrapping : [],
            EditorView.darkTheme.of(isLightTheme === false),
            locationField.init(() => location),
            // Force column view if blameDecorations is set
            enableColumnView.of(!!blameDecorations || enableExtensionsDecorationsColumnView),
            blameDecorations ? showTextDocumentDecorations.of([[blameDecorationType, blameDecorations]]) : [],
        ],
        [wrapCode, isLightTheme, location, enableExtensionsDecorationsColumnView, blameDecorations]
    )
    const [settingsCompartment, updateSettingsCompartment] = useCompartment(settings)

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

    const extensions = useMemo(
        () => [
            selectableLineNumbers({ onSelection }),
            staticExtensions,
            settingsCompartment,
            syntaxHighlight.of(blobInfo),
            sourcegraphExtensions({ blobInfo, extensionsController }),
        ],
        [settingsCompartment, onSelection, blobInfo, extensionsController]
    )

    const editor = useCodeMirror(container, blobInfo.content, extensions, {
        updateValueOnChange: false,
        updateOnExtensionChange: false,
    })

    useEffect(() => {
        if (editor) {
            updateLocation(editor, location)
        }
        // editor is not provided because this should only be triggered after the
        // editor was created (i.e. not on first render)
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [location])

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
    const position = useMemo(() => parseQueryAndHash(location.search, location.hash), [location.search, location.hash])
    useEffect(() => {
        if (editor) {
            // This check is necessary because at the moment the position
            // information is updated before the file content, meaning it's
            // possible that the currently loaded document has fewer lines.
            if (!position?.line || editor.state.doc.lines >= (position.endLine ?? position.line)) {
                selectLines(editor, position.line ? position : null)
            }
        }
        // blobInfo isn't used but we need to trigger the line selection and focus
        // logic whenever the content changes
    }, [editor, position, blobInfo])

    return <div ref={setContainer} aria-label={ariaLabel} role={role} className={`${className} overflow-hidden`} />
}
