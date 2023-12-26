<script lang="ts" context="module">
    import type { BlobFileFields } from '$lib/repo/api/blob'
    import { HovercardView } from '$lib/repo/HovercardView'

    export interface BlobInfo extends BlobFileFields {
        commitID: string
        filePath: string
        repoName: string
        revision: string
    }

    const extensionsCompartment = new Compartment()

    const defaultTheme = EditorView.theme({
        '&': {
            width: '100%',
            'min-height': 0,
            color: 'var(--color-code)',
            flex: 1,
        },
        '.cm-scroller': {
            lineHeight: '1rem',
            fontFamily: 'var(--code-font-family)',
            fontSize: 'var(--code-font-size)',
        },
        '.cm-content:focus-visible': {
            outline: 'none',
            boxShadow: 'none',
        },
        '.cm-gutters': {
            'background-color': 'var(--code-bg)',
            border: 'none',
            color: 'var(--line-number-color)',
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
        '.sourcegraph-document-highlight': {
            backgroundColor: 'var(--secondary)',
        },
        '.selection-highlight': {
            backgroundColor: 'var(--mark-bg)',
        },
        '.cm-tooltip': {
            border: 'none',
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
    ]

    function configureSyntaxHighlighting(content: string, lsif: string): Extension {
        return lsif ? syntaxHighlight.of({ content, lsif }) : []
    }

    function configureMiscSettings({ wrapLines }: { wrapLines: boolean }): Extension {
        return [wrapLines ? EditorView.lineWrapping : []]
    }
</script>

<script lang="ts">
    import '$lib/highlight.scss'

    import { Compartment, EditorState, type Extension } from '@codemirror/state'
    import { EditorView } from '@codemirror/view'
    import { createEventDispatcher, onMount } from 'svelte'

    import { browser } from '$app/environment'
    import {
        selectableLineNumbers,
        syntaxHighlight,
        type SelectedLineRange,
        setSelectedLines,
        isValidLineRange,
        linkify,
        createCodeIntelExtension,
        syncSelection,
        temporaryTooltip,
    } from '$lib/web'
    import { goto } from '$app/navigation'
    import { type CodeIntelAPI } from '$lib/shared'
    import { goToDefinition, openImplementations, openReferences } from './repo/blob'
    import type { LineOrPositionOrRange } from '$lib/common'

    export let blobInfo: BlobInfo
    export let highlights: string
    export let wrapLines: boolean = false
    export let selectedLines: LineOrPositionOrRange | null = null
    export let codeIntelAPI: CodeIntelAPI

    const dispatch = createEventDispatcher<{ selectline: SelectedLineRange }>()

    let editor: EditorView
    let container: HTMLDivElement | null = null

    const lineNumbers = selectableLineNumbers({
        onSelection(range) {
            dispatch('selectline', range)
        },
        initialSelection: selectedLines?.line === undefined ? null : selectedLines,
    })

    $: documentInfo = {
        repoName: blobInfo.repoName,
        commitID: blobInfo.commitID,
        revision: blobInfo.revision,
        filePath: blobInfo.filePath,
        languages: blobInfo.languages,
    }
    $: codeIntelExtension = createCodeIntelExtension({
        api: {
            api: codeIntelAPI,
            documentInfo: documentInfo,
            goToDefinition: (view, definition, options) => goToDefinition(documentInfo, view, definition, options),
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
    $: settings = configureMiscSettings({ wrapLines })
    $: sh = configureSyntaxHighlighting(blobInfo.content, highlights)

    $: extensions = [sh, settings, lineNumbers, temporaryTooltip, codeIntelExtension, staticExtensions]

    function update(blobInfo: BlobInfo, extensions: Extension, range: LineOrPositionOrRange | null) {
        if (editor) {
            // TODO(fkling): Find a way to combine this into a single transaction.
            if (editor.state.sliceDoc() !== blobInfo.content) {
                editor.setState(
                    EditorState.create({ doc: blobInfo.content, extensions: extensionsCompartment.of(extensions) })
                )
            } else {
                editor.dispatch({ effects: [extensionsCompartment.reconfigure(extensions)] })
            }
            editor.dispatch({
                effects: setSelectedLines.of(range?.line && isValidLineRange(range, editor.state.doc) ? range : null),
            })

            if (range) {
                syncSelection(editor, range)
            }
        }
    }

    $: update(blobInfo, extensions, selectedLines)

    onMount(() => {
        if (container) {
            editor = new EditorView({
                state: EditorState.create({ doc: blobInfo.content, extensions: extensionsCompartment.of(extensions) }),
                parent: container,
            })
            if (selectedLines) {
                syncSelection(editor, selectedLines)
            }
        }
    })
</script>

{#if browser}
    <div bind:this={container} class="root test-editor" data-editor="codemirror6" />
{:else}
    <div class="root">
        <pre>{blobInfo.content}</pre>
    </div>
{/if}

<style lang="scss">
    .root {
        display: contents;
        overflow: hidden;
    }
    pre {
        margin: 0;
    }
</style>
