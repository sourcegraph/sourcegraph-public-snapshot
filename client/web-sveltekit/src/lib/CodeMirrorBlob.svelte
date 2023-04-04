<script lang="ts">
    import '@sourcegraph/wildcard/src/global-styles/highlight.scss'

    import { Compartment, EditorState, StateEffect, type Extension } from '@codemirror/state'
    import { EditorView, lineNumbers } from '@codemirror/view'

    import { browser } from '$app/environment'
    import { syntaxHighlight } from '$lib/web'
    import type { BlobFileFields } from '$lib/graphql-operations'

    export let blob: BlobFileFields
    export let highlights: string
    export let wrapLines: boolean = false

    let editor: EditorView
    let container: HTMLDivElement | null = null

    const shCompartment = new Compartment()
    const miscSettingsCompartment = new Compartment()

    function createEditor(container: HTMLDivElement): EditorView {
        const extensions = [
            lineNumbers(),
            miscSettingsCompartment.of(configureMiscSettings({ wrapLines })),
            shCompartment.of(configureSyntaxHighlighting(blob.content, highlights)),
            EditorView.theme({
                '&': {
                    width: '100%',
                    'min-height': 0,
                    color: 'var(--color-code)',
                    flex: 1,
                },
                '.cm-scroller': {
                    overflow: 'auto',
                    'font-family': 'var(--code-font-family)',
                    'font-size': 'var(--code-font-size)',
                },
                '.cm-gutters': {
                    'background-color': 'var(--code-bg)',
                    border: 'none',
                    color: 'var(--line-number-color)',
                },
                '.cm-line': {
                    'line-height': '1rem',
                    'padding-left': '1rem',
                },
            }),
        ]

        const view = new EditorView({
            state: EditorState.create({ doc: blob.content, extensions }),
            parent: container,
        })
        return view
    }

    function configureSyntaxHighlighting(content: string, lsif: string): Extension {
        return lsif ? syntaxHighlight.of({ content, lsif }) : []
    }

    function configureMiscSettings({ wrapLines }: { wrapLines: boolean }): Extension {
        return [wrapLines ? EditorView.lineWrapping : []]
    }

    function updateExtensions(effects: StateEffect<unknown>[]) {
        if (editor) {
            editor.dispatch({ effects })
        }
    }

    $: updateExtensions([shCompartment.reconfigure(configureSyntaxHighlighting(blob.content, highlights))])
    $: updateExtensions([miscSettingsCompartment.reconfigure(configureMiscSettings({ wrapLines }))])

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
