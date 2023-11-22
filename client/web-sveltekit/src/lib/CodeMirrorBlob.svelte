<script lang="ts" context="module">
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
    import { createEventDispatcher } from 'svelte'

    import { browser } from '$app/environment'
    import {
        blobPropsFacet,
        selectableLineNumbers,
        syntaxHighlight,
        type SelectedLineRange,
        setSelectedLines,
        isValidLineRange,
    } from '$lib/web'
    import type { BlobFileFields } from '$lib/repo/api/blob'

    export let blob: BlobFileFields
    export let highlights: string
    export let wrapLines: boolean = false
    export let selectedLines: SelectedLineRange | null = null

    const dispatch = createEventDispatcher<{ selectline: SelectedLineRange }>()

    let editor: EditorView
    let container: HTMLDivElement | null = null

    function createEditor(container: HTMLDivElement): EditorView {
        const extensions = [
            // @ts-ignore - ugly (temporary?) hack to avoid issues with existing extension (selectableLineNumbers)
            blobPropsFacet.of({}),
            staticExtensions,
            selectableLineNumbers({
                onSelection(range) {
                    dispatch('selectline', range)
                },
                initialSelection: selectedLines,
                navigateToLineOnAnyClick: false,
            }),
            miscSettingsCompartment.of(configureMiscSettings({ wrapLines })),
            shCompartment.of(configureSyntaxHighlighting(blob.content, highlights)),
        ]

        const view = new EditorView({
            state: EditorState.create({ doc: blob.content, extensions }),
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
    $: updateExtensions(shCompartment.reconfigure(configureSyntaxHighlighting(blob.content, highlights)))
    // Update line wrapping
    $: updateExtensions(miscSettingsCompartment.reconfigure(configureMiscSettings({ wrapLines })))
    // Update selected line
    $: updateSelectedLines(selectedLines)

    $: if (editor && editor?.state.sliceDoc() !== blob.content) {
        editor.dispatch({
            changes: { from: 0, to: editor.state.doc.length, insert: blob.content },
        })
    }

    $: if (container && !editor) {
        editor = createEditor(container)
    }
</script>

{#if browser}
    <div bind:this={container} class="root test-editor" data-editor="codemirror6" />
{:else}
    <div class="root">
        <pre>{blob.content}</pre>
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
