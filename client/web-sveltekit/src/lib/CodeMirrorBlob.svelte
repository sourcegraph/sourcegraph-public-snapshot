<script lang="ts" context="module">
    import { HovercardView } from '$lib/repo/HovercardView'

    export interface BlobInfo {
        /**
         * Name of the repository this file belongs to.
         */
        repoName: string
        /**
         * The commit OID of the currently viewed commit.
         */
        commitID: string
        /**
         * Human readable version of the current commit (e.g. branch name).
         */
        revision: string
        /**
         * The path of the file relative to the repository root.
         */
        filePath: string
        /**
         * The content of the file.
         */
        content: string
        /**
         * The language of the file.
         */
        languages: string[]
    }

    const defaultTheme = EditorView.theme({
        '&': {
            height: '100%',
            color: 'var(--color-code)',
        },
        '&.cm-focused': {
            outline: 'none',
        },
        '.cm-scroller': {
            lineHeight: '1rem',
            fontFamily: 'var(--code-font-family)',
            fontSize: 'var(--code-font-size)',
            overflow: 'auto',
        },
        '.cm-content': {
            paddingBottom: '1.5rem',
            '&:focus-visible': {
                outline: 'none',
                boxShadow: 'none',
            },
        },
        '.cm-panels': {
            '&-top': {
                borderBottom: '1px solid var(--border-color)',
            },
            backgroundColor: 'transparent',
        },
        '.cm-gutters': {
            'background-color': 'var(--code-bg)',
            border: 'none',
            color: 'var(--line-number-color)',
        },
        '.cm-gutterElement': {
            lineHeight: '1.54',
            minWidth: '40px !important',

            '&:hover': {
                color: 'var(--text-body)',
            },
        },
        '.cm-lineNumbers .cm-gutterElement': {
            padding: '0 1.5ex',
        },
        '.cm-line': {
            lineHeight: '1.54',
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
        '.sourcegraph-document-highlight': {
            backgroundColor: 'var(--secondary)',
        },
        '.selection-highlight': {
            backgroundColor: 'var(--mark-bg)',
        },
        '.cm-tooltip': {
            border: 'none',
            // For nice rounded corners in hover cards
            borderRadius: 'var(--border-radius)',
        },
    })

    const staticExtensions: Extension = [
        // Log uncaught errors that happen in callbacks that we pass to
        // CodeMirror. Without this exception sink, exceptions get silently
        // ignored making it difficult to debug issues caused by uncaught
        // exceptions.
        // eslint-disable-next-line no-console
        EditorView.exceptionSink.of(exception => console.log(exception)),
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
        defaultTheme,
        linkify,
        hideEmptyLastLine,
    ]
</script>

<script lang="ts">
    import '$lib/highlight.scss'

    import { openSearchPanel } from '@codemirror/search'
    import { EditorState, type Extension } from '@codemirror/state'
    import { EditorView } from '@codemirror/view'
    import { createEventDispatcher, onMount } from 'svelte'

    import { Occurrence, Range as SCIPRange, Position } from '@sourcegraph/shared/src/codeintel/scip'
    import { codeGraphData as codeGraphDataFacet } from '@sourcegraph/web/src/repo/blob/codemirror/codeintel/occurrences'

    import { browser } from '$app/environment'
    import { goto } from '$app/navigation'
    import type { LineOrPositionOrRange } from '$lib/common'
    import type { CodeIntelAPI } from '$lib/shared'
    import {
        selectableLineNumbers,
        syntaxHighlight,
        type SelectedLineRange,
        setSelectedLines,
        isValidLineRange,
        linkify,
        createCodeIntelExtension,
        syncSelection,
        showBlame as showBlameColumn,
        blameData as blameDataFacet,
        type BlameHunkData,
        lockFirstVisibleLine,
        temporaryTooltip,
        hideEmptyLastLine,
        search,
    } from '$lib/web'

    import BlameDecoration from './blame/BlameDecoration.svelte'
    import { SearchPanel, keyboardShortcut } from './codemirror/inline-search'
    import { type Range, staticHighlights } from './codemirror/static-highlights'
    import {
        createCompartments,
        restoreScrollSnapshot,
        type ExtensionType,
        type ScrollSnapshot,
        getScrollSnapshot as getScrollSnapshot_internal,
    } from './codemirror/utils'
    import type { CodeGraphData } from './graphql-types'
    import { registerHotkey } from './Hotkey'
    import { goToDefinition, openImplementations, openReferences } from './repo/blob'
    import { createLocalWritable } from './stores'

    export let blobInfo: BlobInfo
    export let highlights: string
    export let codeGraphData: CodeGraphData[] = []
    export let wrapLines: boolean = false
    export let selectedLines: LineOrPositionOrRange | null = null
    export let codeIntelAPI: CodeIntelAPI | null
    export let staticHighlightRanges: Range[] = []
    export let onCopy: () => void = () => {}
    /**
     * The initial scroll position when the editor is first mounted.
     * Changing the value afterwards has no effect.
     */
    export let initialScrollPosition: ScrollSnapshot | null = null

    export let showBlame: boolean = false
    export let blameData: BlameHunkData | undefined = undefined

    export function getScrollSnapshot(): ScrollSnapshot | null {
        return view ? getScrollSnapshot_internal(view) : null
    }

    const dispatch = createEventDispatcher<{ selectline: SelectedLineRange }>()
    const extensionsCompartment = createCompartments({
        selectableLineNumbers: null,
        syntaxHighlighting: null,
        lineWrapping: null,
        temporaryTooltip,
        codeIntelExtension: null,
        staticExtensions,
        staticHighlightExtension: null,
        blameDataExtension: null,
        blameColumnExtension: null,
        searchExtension: null,
        codeGraph: null,
    })
    const useFileSearch = createLocalWritable('blob.overrideBrowserFindOnPage', true)
    registerHotkey({
        keys: keyboardShortcut,
        handler(event) {
            if ($useFileSearch && view) {
                event.preventDefault()
                openSearchPanel(view)
            }
            // fall back to browser's find in page
        },
        allowDefault: true,
    })

    let container: HTMLDivElement | null = null
    let view: EditorView | undefined = undefined

    $: documentInfo = {
        repoName: blobInfo.repoName,
        commitID: blobInfo.commitID,
        revision: blobInfo.revision,
        filePath: blobInfo.filePath,
        languages: blobInfo.languages,
    }
    $: codeIntelExtension = codeIntelAPI
        ? createCodeIntelExtension({
              api: {
                  api: codeIntelAPI,
                  documentInfo: documentInfo,
                  goToDefinition: (view, definition, options) =>
                      goToDefinition(documentInfo, view, definition, options),
                  openReferences,
                  openImplementations,
                  createTooltipView: options => new HovercardView(options.view, options.token, options.hovercardData),
              },
              // TODO(fkling): Support tooltip pinning
              pin: {},
              navigate: to => {
                  if (typeof to === 'number') {
                      if (to > 0) {
                          history.forward()
                      } else {
                          history.back()
                      }
                  } else {
                      goto(to.toString())
                  }
              },
          })
        : null
    $: lineWrapping = wrapLines ? EditorView.lineWrapping : null
    $: syntaxHighlighting = highlights ? syntaxHighlight.of({ content: blobInfo.content, lsif: highlights }) : null
    $: codeGraph = codeGraphDataFacet.of(
        codeGraphData?.map(datum => ({
            provenance: datum.provenance,
            occurrences:
                datum.occurrences?.nodes.map(
                    occ =>
                        new Occurrence(
                            new SCIPRange(
                                new Position(occ.range.start.line, occ.range.start.character),
                                new Position(occ.range.end.line, occ.range.end.character)
                            ),
                            undefined,
                            occ.symbol ?? undefined,
                            undefined // TODO: how to convert to numberic roles?
                        )
                ) ?? [],
        })) ?? []
    )
    $: staticHighlightExtension = staticHighlights(staticHighlightRanges)
    $: searchExtension = search({
        overrideBrowserFindInPageShortcut: $useFileSearch,
        onOverrideBrowserFindInPageToggle(enabled) {
            useFileSearch.set(enabled)
        },
        createPanel(options) {
            return new SearchPanel(options)
        },
    })

    $: blameColumnExtension = showBlame
        ? showBlameColumn({
              createBlameDecoration(target, props) {
                  const decoration = new BlameDecoration({ target, props })
                  return {
                      destroy() {
                          decoration.$destroy()
                      },
                  }
              },
          })
        : null
    $: blameDataExtension = blameDataFacet(blameData)

    // Reinitialize the editor when its content changes. Update only the extensions when they change.
    $: update(view => {
        // blameColumnExtension is omitted here. It's updated separately below because we need to
        // apply additional effects when it changes (but only when it changes).
        const extensions: Partial<ExtensionType<typeof extensionsCompartment>> = {
            codeIntelExtension,
            lineWrapping,
            syntaxHighlighting,
            codeGraph,
            staticHighlightExtension,
            blameDataExtension,
            searchExtension,
        }
        if (view.state.sliceDoc() !== blobInfo.content) {
            view.setState(createEditorState(blobInfo, extensions))
        } else {
            extensionsCompartment.update(view, extensions)
        }
    })

    // Show/hide the blame column and ensure that the style changes do not change the scroll position
    $: update(view => {
        extensionsCompartment.update(view, { blameColumnExtension }, ...lockFirstVisibleLine(view))
    })

    // Update the selected lines. This will scroll the selected lines into view. Also set the editor's
    // selection (essentially the cursor position) to the selected lines. This is necessary in case the
    // selected range references a symbol.
    $: update(view => {
        view.dispatch({
            effects: setSelectedLines.of(
                selectedLines?.line && isValidLineRange(selectedLines, view.state.doc) ? selectedLines : null
            ),
        })
        if (selectedLines) {
            syncSelection(view, selectedLines)
        }
    })

    onMount(() => {
        if (container) {
            view = new EditorView({
                // On first render initialize all extensions
                state: createEditorState(blobInfo, {
                    codeIntelExtension,
                    lineWrapping,
                    syntaxHighlighting,
                    codeGraph,
                    staticHighlightExtension,
                    blameDataExtension,
                    blameColumnExtension,
                    searchExtension,
                }),
                parent: container,
            })
            if (selectedLines) {
                syncSelection(view, selectedLines)
            }
            if (initialScrollPosition) {
                restoreScrollSnapshot(view, initialScrollPosition)
            }
        }
        return () => {
            view?.destroy()
        }
    })

    // Helper function to update the editor state whithout depending on the view variable
    // (those updates should only run on subsequent updates)
    function update(updater: (view: EditorView) => void) {
        if (view) {
            updater(view)
        }
    }

    function createEditorState(blobInfo: BlobInfo, extensions: Partial<ExtensionType<typeof extensionsCompartment>>) {
        return EditorState.create({
            doc: blobInfo.content,
            extensions: extensionsCompartment.init({
                selectableLineNumbers: selectableLineNumbers({
                    onSelection(range) {
                        dispatch('selectline', range)
                    },
                    initialSelection: selectedLines?.line === undefined ? null : selectedLines,
                    // We don't want to scroll the selected line into view when a scroll position is explicitly set.
                    skipInitialScrollIntoView: initialScrollPosition !== null,
                }),
                ...extensions,
            }),
            selection: {
                anchor: 0,
            },
        })
    }
</script>

{#if browser}
    <div bind:this={container} class="root test-editor" data-editor="codemirror6" on:copy={onCopy} />
{:else}
    <div class="root">
        <pre>{blobInfo.content}</pre>
    </div>
{/if}

<style lang="scss">
    .root {
        --blame-decoration-width: 300px;
        --blame-recency-width: 4px;

        height: 100%;
    }
    pre {
        margin: 0;
    }
</style>
