<script lang="ts" context="module">
    import type { BlobFileFields } from '$lib/repo/api/blob'

    export interface BlobInfo extends BlobFileFields {
        commitID: string
        filePath: string
        repoName: string
        revision: string
    }

    const shCompartment = new Compartment()
    const miscSettingsCompartment = new Compartment()

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
        buildLinks.of(true),
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

    import { Compartment, EditorState, StateEffect, type Extension } from '@codemirror/state'
    import { EditorView } from '@codemirror/view'
    import { createEventDispatcher, onMount } from 'svelte'

    import { browser } from '$app/environment'
    import {
        blobPropsFacet,
        selectableLineNumbers,
        syntaxHighlight,
        type SelectedLineRange,
        setSelectedLines,
        isValidLineRange,
        codeIntelAPI as codeIntelAPIFacet,
        buildLinks,
    } from '$lib/web'
    import type { CodeIntelAPI } from '@sourcegraph/shared/src/codeintel/api'

    export let blobInfo: BlobInfo
    export let highlights: string
    export let wrapLines: boolean = false
    export let selectedLines: SelectedLineRange | null = null
    export let codeIntelAPI: CodeIntelAPI

    const dispatch = createEventDispatcher<{ selectline: SelectedLineRange }>()

    let editor: EditorView
    let container: HTMLDivElement | null = null

    function createEditor(container: HTMLDivElement): EditorView {
        const extensions = [
            // @ts-ignore - ugly (temporary?) hack to avoid issues with existing extension (selectableLineNumbers)
            blobPropsFacet.of({
                blobInfo,
            }),
            codeIntelAPIFacet.of(codeIntelAPI),
            staticExtensions,
            selectableLineNumbers({
                onSelection(range) {
                    dispatch('selectline', range)
                },
                initialSelection: selectedLines,
                navigateToLineOnAnyClick: false,
            }),
            miscSettingsCompartment.of(configureMiscSettings({ wrapLines })),
            shCompartment.of(configureSyntaxHighlighting(blobInfo.content, highlights)),
        ]

        const view = new EditorView({
            state: EditorState.create({ doc: blobInfo.content, extensions }),
            parent: container,
        })
        return view
    }

    function updateExtensions(effects: StateEffect<unknown> | readonly StateEffect<unknown>[]) {
        if (editor) {
            editor.dispatch({ effects })
        }
    }

    function updateSelectedLines(range: SelectedLineRange) {
        if (editor) {
            updateExtensions(setSelectedLines.of(range && isValidLineRange(range, editor.state.doc) ? range : null))
        }
    }

    // Update blob content and highlights
    $: updateExtensions(shCompartment.reconfigure(configureSyntaxHighlighting(blobInfo.content, highlights)))
    // Update line wrapping
    $: updateExtensions(miscSettingsCompartment.reconfigure(configureMiscSettings({ wrapLines })))
    // Update selected line
    $: updateSelectedLines(selectedLines)

    $: if (editor && editor?.state.sliceDoc() !== blobInfo.content) {
        editor.dispatch({
            changes: { from: 0, to: editor.state.doc.length, insert: blobInfo.content },
        })
    }

    onMount(() => {
        if (container) {
            editor = createEditor(container)
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
