/**
 * An implementation of the Blob view using CodeMirror
 */

import { type MutableRefObject, useCallback, useEffect, useMemo, useRef, useState, RefObject } from 'react'

import { openSearchPanel } from '@codemirror/search'
import { EditorState, type Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { isEqual } from 'lodash'
import { createRoot } from 'react-dom/client'
import { createPath, type NavigateFunction, useLocation, useNavigate, type Location } from 'react-router-dom'

import { NoopEditor } from '@sourcegraph/cody-shared/dist/editor'
import {
    addLineRangeQueryParameter,
    formatSearchParameters,
    toPositionOrRangeQueryParameter,
} from '@sourcegraph/common'
import { getOrCreateCodeIntelAPI, type CodeIntelAPI } from '@sourcegraph/shared/src/codeintel/api'
import { editorHeight, useCodeMirror, useCompartment } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import type { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { useKeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts/useKeyboardShortcut'
import type { PlatformContext, PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Shortcut } from '@sourcegraph/shared/src/react-shortcuts'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    type AbsoluteRepoFile,
    type ModeSpec,
    parseQueryAndHash,
    toPrettyBlobURL,
} from '@sourcegraph/shared/src/util/url'
import { useLocalStorage } from '@sourcegraph/wildcard'

import { CodeMirrorEditor } from '../../cody/components/CodeMirrorEditor'
import { isCodyEnabled } from '../../cody/isCodyEnabled'
import { useCodySidebar } from '../../cody/sidebar/Provider'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import type { ExternalLinkFields, Scalars } from '../../graphql-operations'
import type { BlameHunkData } from '../blame/useBlameHunks'
import type { HoverThresholdProps } from '../RepoContainer'

import { BlameDecoration } from './BlameDecoration'
import { blobPropsFacet } from './codemirror'
import { blameData, showBlame } from './codemirror/blame-decorations'
import { codeFoldingExtension } from './codemirror/code-folding'
import { hideEmptyLastLine } from './codemirror/eof'
import { syntaxHighlight } from './codemirror/highlight'
import { selectableLineNumbers, type SelectedLineRange, selectLines } from './codemirror/linenumbers'
import { linkify } from './codemirror/links'
import { lockFirstVisibleLine } from './codemirror/lock-line'
import { navigateToLineOnAnyClickExtension } from './codemirror/navigate-to-any-line-on-click'
import { scipSnapshot } from './codemirror/scip-snapshot'
import { search } from './codemirror/search'
import { sourcegraphExtensions } from './codemirror/sourcegraph-extensions'
import { CodeIntelAPIAdapter } from './codemirror/codeintel/api'
import { createCodeIntelExtension } from './codemirror/codeintel/extension'
import { pinnedLocation } from './codemirror/codeintel/pin'
import { syncOccurrencesWithURL } from './codemirror/codeintel/selections'
import { codyWidgetExtension } from './codemirror/tooltips/CodyTooltip'
import { HovercardView } from './codemirror/tooltips/HovercardView'
import { locationToURL } from './codemirror/utils'
import { setBlobEditView } from './use-blob-store'

/**
 * The minimum number of milliseconds that must elapse before we handle a "Go to
 * definition request".  The motivation to impose a minimum latency on this
 * action is to give the user feedback that something happened if they rapidly
 * trigger "Go to definition" from the same location and the destination token
 * is already visible in the viewport.  Without this minimum latency, the user
 * gets no feedback that the destination is visible.  With this latency, the
 * source token (where the user clicks) gets briefly focused before the focus
 * moves back to the destination token. This small wiggle in the focus state
 * makes it easier to find the destination token.
 */
const MINIMUM_GO_TO_DEF_LATENCY_MILLIS = 20

// Logical grouping of props that are only used by the CodeMirror blob view
// implementation.
interface CodeMirrorBlobProps {
    overrideBrowserSearchKeybinding?: boolean
}

export interface BlobProps
    extends SettingsCascadeProps,
        PlatformContextProps,
        TelemetryProps,
        HoverThresholdProps,
        ExtensionsControllerProps,
        CodeMirrorBlobProps {
    className: string
    wrapCode: boolean
    /** The current text document to be rendered and provided to extensions */
    blobInfo: BlobInfo
    'data-testid'?: string

    // When navigateToLineOnAnyClick=true, the code intel popover is disabled
    // and clicking on any line should navigate to that specific line.
    navigateToLineOnAnyClick?: boolean

    // If set, nav is called when a user clicks on a token highlighted by
    // WebHoverOverlay
    nav?: (url: string) => void
    role?: string
    ariaLabel?: string

    supportsFindImplementations?: boolean

    isBlameVisible?: boolean
    blameHunks?: BlameHunkData

    activeURL?: string
}

export interface BlobPropsFacet extends BlobProps {
    navigate: NavigateFunction
    location: Location
}

export interface BlobInfo extends AbsoluteRepoFile, ModeSpec {
    /** The raw content of the blob. */
    content: string

    /** LSIF syntax-highlighting data */
    lsif?: string

    /** If present, the file is stored in Git LFS (large file storage). */
    lfs?: { byteSize: Scalars['BigInt'] } | null

    /** External URLs for the file */
    externalURLs?: ExternalLinkFields[]

    snapshotData?: { offset: number; data: string; additional: string[] | null }[] | null
}

const staticExtensions: Extension = [
    // Log uncaught errors that happen in callbacks that we pass to
    // CodeMirror. Without this exception sink, exceptions get silently
    // ignored making it difficult to debug issues caused by uncaught
    // exceptions.
    // eslint-disable-next-line no-console
    EditorView.exceptionSink.of(exception => console.log(exception)),
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
        '.cm-content:focus-visible': {
            outline: 'none',
            boxShadow: 'none',
        },
        '.cm-line': {
            paddingLeft: '0',
        },
        '.selected-line': {
            backgroundColor: 'var(--code-selection-bg)',

            '&:focus': {
                boxShadow: 'none',
            },
        },
        '.highlighted-line': {
            backgroundColor: 'var(--code-selection-bg)',
        },
    }),
    hideEmptyLastLine,
    linkify,
    hoverCardConstructor.of(HovercardView),
]

export const CodeMirrorBlob: React.FunctionComponent<BlobProps> = props => {
    const {
        className,
        wrapCode,
        ariaLabel,
        role,
        extensionsController,
        isBlameVisible,
        blameHunks,

        // Reference panel specific props
        navigateToLineOnAnyClick,

        overrideBrowserSearchKeybinding,
        'data-testid': dataTestId,
    } = props

    const navigate = useNavigate()
    const location = useLocation()

    const [enableBlobPageSwitchAreasShortcuts] = useFeatureFlag('blob-page-switch-areas-shortcuts')
    const focusCodeEditorShortcut = useKeyboardShortcut('focusCodeEditor')

    const [useFileSearch, setUseFileSearch] = useLocalStorage('blob.overrideBrowserFindOnPage', true)

    const containerRef = useRef<HTMLDivElement | null>(null)
    // This is used to avoid reinitializing the editor when new locations in the
    // same file are opened inside the reference panel.
    const blobInfo = useDistinctBlob(props.blobInfo)
    const position = useMemo(() => {
        // When an activeURL is passed, it takes presedence over the react
        // router location API.
        //
        // This is needed to support the reference panel
        if (props.activeURL) {
            const url = new URL(props.activeURL, window.location.href)
            return parseQueryAndHash(url.search, url.hash)
        }
        return parseQueryAndHash(location.search, location.hash)
    }, [props.activeURL, location.search, location.hash])
    const hasPin = useMemo(() => urlIsPinned(location.search), [location.search])

    // Keep history and location in a ref so that we can use the latest value in
    // the onSelection callback without having to recreate it and having to
    // reconfigure the editor extensions
    const navigateRef = useMutableValue(navigate)
    const locationRef = useMutableValue(location)

    const navigateOnClick = useMemo(
        () =>
            navigateToLineOnAnyClick
                ? (line: number) =>
                      navigate(
                          toPrettyBlobURL({
                              repoName: blobInfo.repoName,
                              filePath: blobInfo.filePath,
                              revision: blobInfo.revision,
                              commitID: blobInfo.commitID,
                              position: { line, character: 0 },
                          })
                      )
                : undefined,
        [navigateToLineOnAnyClick, navigate, blobInfo.repoName, blobInfo.filePath, blobInfo.revision, blobInfo.commitID]
    )

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

    // Added fallback to take care of ReferencesPanel/Simple storybook
    const { setEditorScope } = useCodySidebar()

    const editorRef = useRef<EditorView | null>(null)

<<<<<<< HEAD
    const blameDecorations = useBlameDecoration(editorRef, { visible: !!isBlameVisible, blameHunks })
    const blobProps = useCompartment(
        editorRef,
        useMemo(
            () =>
                blobPropsFacet.of({
                    ...props,
                    navigate,
                    location,
                }),
            [props, navigate, location]
        )
    )
    const wrapCodeSettings = useCompartment(
        editorRef,
        useMemo<Extension>(() => (wrapCode ? EditorView.lineWrapping : []), [wrapCode])
    )
    const codeIntelExtension = useCodeIntelExtension(
        props.platformContext,
        { repoName: blobInfo.repoName, filePath: blobInfo.filePath, commitID: blobInfo.commitID },
        blobInfo.mode
    )
    const pinnedTooltip = useCompartment(
        editorRef,
        useMemo(() => pinnedLocation.of(hasPin ? position : null), [hasPin, position])
    )
    const extensions = useMemo(
        () => [
            staticExtensions,
            selectableLineNumbers({
                onSelection,
                initialSelection: position.line !== undefined ? position : null,
                onClick: navigateOnClick,
            }),
            scipSnapshot(blobInfo.content, blobInfo.snapshotData),
            codeFoldingExtension(),
            isCodyEnabled()
                ? codyWidgetExtension(
                      editorRef.current
                          ? new CodeMirrorEditor({
                                view: editorRef.current,
                                repo: props.blobInfo.repoName,
                                revision: props.blobInfo.revision,
                                filename: props.blobInfo.filePath,
                                content: props.blobInfo.content,
                            })
                          : undefined
                  )
                : [],
            search({
                // useFileSearch is not a dependency because the search
                // extension manages its own state. This is just the initial
                // value
                overrideBrowserFindInPageShortcut: useFileSearch,
                onOverrideBrowserFindInPageToggle: setUseFileSearch,
                navigate,
            }),
            pinnedTooltip,
            navigateToLineOnAnyClick ? navigateToLineOnAnyClickExtension(navigate) : codeIntelExtension,
            syntaxHighlight.of(blobInfo),
            extensionsController !== null && !navigateToLineOnAnyClick
                ? sourcegraphExtensions({
                      blobInfo,
                      initialSelection: position,
                      extensionsController,
                  })
                : [],
            blobProps,
            blameDecorations,
            wrapCodeSettings,
            search({
                // useFileSearch is not a dependency because the search
                // extension manages its own state. This is just the initial
                // value
                overrideBrowserFindInPageShortcut: useFileSearch,
                onOverrideBrowserFindInPageToggle: setUseFileSearch,
                navigate,
            }),
        ],
        // A couple of values are not dependencies (hasPin and position) because those are updated in effects
        // further below. However, they are still needed here because we need to
        // set initial values when we re-initialize the editor.
        // eslint-disable-next-line react-hooks/exhaustive-deps
        [
            onSelection,
            navigate,
            blobInfo,
            extensionsController,
            isCodyEnabled,
            codeIntelExtension,
            editorRef.current,
            blameDecorations,
            pinnedTooltip,
        ]
    )

    // Reconfigure editor when blobInfo or core extensions changed
    useEffect(() => {
        const editor = editorRef.current
        if (editor) {
            // We use setState here instead of dispatching a transaction because
            // the new document has nothing to do with the previous one and so
            // any existing state should be discarded.
            const state = EditorState.create({ doc: blobInfo.content, extensions })
            editor.setState(state)

            if (navigateToLineOnAnyClick) {
                /**
                 * `navigateToLineOnAnyClick` is `true` when CodeMirrorBlob is rendered in the references panel.
                 * We don't need code intel and keyboard navigation in the references panel blob: https://github.com/sourcegraph/sourcegraph/pull/41615.
                 */
                return
            }

            // Sync editor selection/focus with the URL so that triggering
            // `history.goBack/goForward()` works similar to the "Go back"
            // command in VS Code.
            syncOccurrencesWithURL(locationRef.current, editor)
            // Automatically focus the content DOM to enable keyboard
            // navigation. Without this automatic focus, users need to click
            // on the blob view with the mouse.
            // NOTE: this focus statment does not seem to have an effect
            // when using macOS VoiceOver.
            editor.contentDOM.focus({ preventScroll: true })
        }
    }, [blobInfo, extensions, navigateToLineOnAnyClick, locationRef])

    // Update selected lines when URL changes
    useEffect(() => {
        const editor = editorRef.current
        if (editor) {
            selectLines(editor, position.line ? position : null)
        }
    }, [position])

    // Listens to location changes and update editor selection accordingly.
    useEffect(() => {
        if (editorRef.current) {
            syncOccurrencesWithURL(location, editorRef.current)
        }
    }, [location])

    useCodeMirror(
        editorRef,
        containerRef,
        // We update the value ourselves
        // eslint-disable-next-line react-hooks/exhaustive-deps
        useMemo(() => blobInfo.content, []),
        // We update extensions ourselves
        // eslint-disable-next-line react-hooks/exhaustive-deps
        useMemo(() => extensions, [])
    )

    // Sync editor store with global Zustand store API
    useEffect(() => setBlobEditView(editorRef.current ?? null), [])

    const openSearch = useCallback(() => {
        if (editorRef.current) {
            openSearchPanel(editorRef.current)
        }
    }, [])

    // Sync the currently viewed document with the editor zustand store. This is used for features
    // like Cody to know what file and range a user is looking at.
    useEffect(() => {
        const view = editorRef.current
        setEditorScope(
            new CodeMirrorEditor(
                view
                    ? {
                          view,
                          repo: props.blobInfo.repoName,
                          revision: props.blobInfo.revision,
                          filename: props.blobInfo.filePath,
                          content: props.blobInfo.content,
                      }
                    : undefined
            )
        )
        return () => setEditorScope(new NoopEditor())
    }, [
        props.blobInfo.content,
        props.blobInfo.filePath,
        props.blobInfo.repoName,
        props.blobInfo.revision,
        setEditorScope,
    ])

    return (
        <>
            <div
                ref={containerRef}
                aria-label={ariaLabel}
                role={role}
                data-testid={dataTestId}
                className={`${className} overflow-hidden test-editor`}
                data-editor="codemirror6"
            />
            {overrideBrowserSearchKeybinding && useFileSearch && (
                <Shortcut ordered={['f']} held={['Mod']} onMatch={openSearch} ignoreInput={true} />
            )}
            {enableBlobPageSwitchAreasShortcuts &&
                focusCodeEditorShortcut?.keybindings.map((keybinding, index) => (
                    <Shortcut
                        key={index}
                        {...keybinding}
                        allowDefault={true}
                        onMatch={() => {
                            editorRef.current?.contentDOM.focus()
                        }}
                    />
                ))}
        </>
    )
}

function useCodeIntelExtension(
    context: PlatformContext,
    documentInfo: { repoName: string; filePath: string; commitID: string },
    mode: string
): Extension {
    const navigate = useNavigate()
    const location = useLocation()
    const locationRef = useRef(location)
    const [api, setApi] = useState<CodeIntelAPI | null>(null)

    useEffect(() => {
        locationRef.current = location
    }, [location])

    useEffect(() => {
        let ignore = false
        getOrCreateCodeIntelAPI(context).then(api => {
            if (!ignore) {
                setApi(api)
            }
        })
        return () => {
            ignore = true
        }
    }, [context])

    return useMemo(
        () =>
            api
                ? createCodeIntelExtension({
                      api: new CodeIntelAPIAdapter({
                          api,
                          documentInfo,
                          mode,
                          createTooltipView: ({ view, token, hovercardData }) =>
                              new HovercardView(view, token, hovercardData),
                          goToDefinition(view, definition) {
                              switch (definition.type) {
                                  case 'none':
                                      break
                                  case 'at-definition': {
                                      // TODO: show temporary tooltip
                                      // showTemporaryTooltip(view, 'You are at the definition', position, 2000, { arrow: true })

                                      // Open reference panel
                                      navigate(locationToURL(documentInfo, definition.from, 'references'), {
                                          replace: true,
                                      })
                                      break
                                  }
                                  case 'single': {
                                      interface DefinitionState {
                                          // The destination URL if we trigger `history.goBack()`.  We use this state
                                          // to avoid inserting redundant 'A->B->A->B' entries when the user triggers
                                          // "go to definition" twice in a row from the same location.
                                          previousURL?: string
                                      }

                                      const locationState = locationRef.current.state as DefinitionState
                                      const hrefFrom = locationToURL(documentInfo, definition.from)
                                      // Don't push URLs into the history if the last goto-def
                                      // action was from the same URL same as this action. This
                                      // happens when the user repeatedly triggers goto-def, which
                                      // is easy to do when the definition URL is close to
                                      // where the action got triggered.
                                      const shouldPushHistory = locationState?.previousURL !== hrefFrom
                                      // Add browser history entry for reference location. This allows users
                                      // to easily jump back to the location they triggered 'go to definition'
                                      // from.
                                      navigate(hrefFrom, {
                                          replace: !shouldPushHistory || createPath(locationRef.current) === hrefFrom,
                                      })

                                      setTimeout(() => {
                                          navigate(locationToURL(documentInfo, definition.destination), {
                                              replace: !shouldPushHistory,
                                              state: { previousURL: hrefFrom },
                                          })
                                      }, MINIMUM_GO_TO_DEF_LATENCY_MILLIS)
                                      break
                                  }
                                  case 'multiple': {
                                      // Linking to the reference panel is a temporary workaround until we
                                      // implement a component to resolve ambiguous results inside the blob
                                      // view similar to how VS Code "Peek definition" works like.
                                      navigate(locationToURL(documentInfo, definition.destination, 'def'))
                                      break
                                  }
                              }
                          },
                      }),
                      navigate,
                  })
                : [],
        [documentInfo.repoName, documentInfo.filePath, documentInfo.commitID, mode, api, navigate, locationRef]
    )
}

/**
 * Create and update blame decorations.
 */
function useBlameDecoration(
    editorRef: RefObject<EditorView>,
    { visible, blameHunks, navigate }: { visible: boolean; blameHunks?: BlameHunkData; navigate: NavigateFunction }
): Extension {
    // Blame support is split into two compartments because we only want to trigger
    // `lockFirstVisibleLine` when blame is enabled, not when data is received
    // (this can cause the editor to scroll to a different line)
    const enabled = useCompartment(
        editorRef,
        useMemo(() => (visible ? enableBlame(navigate) : []), [visible, navigate]),
        lockFirstVisibleLine
    )

    const data = useCompartment(
        editorRef,
        useMemo(() => blameData(blameHunks), [blameHunks])
    )
    return useMemo(() => [enabled, data], [enabled, data])
}

/**
 * Create and update blame decorations.
 */
function useBlameDecoration(
    editorRef: RefObject<EditorView>,
    { visible, blameHunks }: { visible: boolean; blameHunks?: BlameHunkData }
): Extension {
    const navigate = useNavigate()

    // Blame support is split into two compartments because we only want to trigger
    // `lockFirstVisibleLine` when blame is enabled, not when data is received
    // (this can cause the editor to scroll to a different line)
    const enabled = useCompartment(
        editorRef,
        useMemo(
            () =>
                visible
                    ? showBlame({
                          createBlameDecoration(container, { line, hunk, onSelect, onDeselect, externalURLs }) {
                              const root = createRoot(container)
                              root.render(
                                  <BlameDecoration
                                      navigate={navigate}
                                      line={line ?? 0}
                                      blameHunk={hunk}
                                      onSelect={onSelect}
                                      onDeselect={onDeselect}
                                      externalURLs={externalURLs}
                                  />
                              )
                              return {
                                  destroy() {
                                      root.unmount()
                                  },
                              }
                          },
                      })
                    : [],
            [visible, navigate]
        ),
        lockFirstVisibleLine
    )

    const data = useCompartment(
        editorRef,
        useMemo(() => blameData(blameHunks), [blameHunks])
    )
    return useMemo(() => [enabled, data], [enabled, data])
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

function useMutableValue<T>(value: T): Readonly<MutableRefObject<T>> {
    const valueRef = useRef(value)

    useEffect(() => {
        valueRef.current = value
    }, [value])

    return valueRef
}

/**
 * Adds an entry to the browser history only if new search parameters differ
 * from the current ones. This prevents adding a new entry when e.g. the user
 * clicks the same line multiple times.
 */
export function updateBrowserHistoryIfChanged(
    navigate: NavigateFunction,
    location: Location,
    newSearchParameters: URLSearchParams,
    /** If set to true replace the current history entry instead of adding a new one. */
    replace: boolean = false
): void {
    const currentSearchParameters = [...new URLSearchParams(location.search).entries()]

    // Update history if the number of search params changes or if any parameter
    // value changes. This will also work for file position changes, which are
    // encoded as parameter without a value. The old file position will be a
    // non-existing key in the new search parameters and thus return `null`
    // (whereas it returns an empty string in the current search parameters).
    const needsUpdate =
        currentSearchParameters.length !== [...newSearchParameters.keys()].length ||
        currentSearchParameters.some(([key, value]) => newSearchParameters.get(key) !== value)

    if (needsUpdate) {
        const entry = {
            ...location,
            search: formatSearchParameters(newSearchParameters),
        }

        navigate(entry, { replace })
    }
}
