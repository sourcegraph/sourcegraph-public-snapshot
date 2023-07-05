<script lang="ts">
    import '@sourcegraph/wildcard/src/global-styles/highlight.scss'

    import { Compartment, EditorState, StateEffect, type Extension, type TransactionSpec } from '@codemirror/state'
    import { Decoration, EditorView, lineNumbers } from '@codemirror/view'

    import { browser } from '$app/environment'
    import type { BlobFileFields } from '$lib/graphql-operations'
    import { syntaxHighlight } from '$lib/web'

    export let blob: BlobFileFields
    export let highlights: string
    export let wrapLines: boolean = false
    export let focusedLine: number|undefined = undefined
    export let extension: Extension = []

    let editor: EditorView
    let container: HTMLDivElement | null = null

    const shCompartment = new Compartment()
    const miscSettingsCompartment = new Compartment()
    const focusedLineCompartment = new Compartment()
    const externalExtensionsCompartement = new Compartment()

    function createEditor(container: HTMLDivElement): EditorView {
        const extensions = [
            lineNumbers(),
            EditorView.editable.of(false),
            miscSettingsCompartment.of(configureMiscSettings({ wrapLines })),
            shCompartment.of(configureSyntaxHighlighting(blob.content, highlights)),
            focusedLineCompartment.of(configureFocusedLine(focusedLine)),
            externalExtensionsCompartement.of(extension),
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

    const lineDecoration = Decoration.line({class: 'focused-line'})

    function configureFocusedLine(line?: number): Extension {
        return [line !== undefined ? EditorView.decorations.of(editor => Decoration.set(lineDecoration.range(editor.state.doc.line(line).from))) : []]
    }

    function updateExtensions(effects: StateEffect<unknown>[]) {
        if (editor) {
            editor.dispatch({ effects })
        }
    }

    $: updateExtensions([shCompartment.reconfigure(configureSyntaxHighlighting(blob.content, highlights))])
    $: updateExtensions([miscSettingsCompartment.reconfigure(configureMiscSettings({ wrapLines }))])
    $: updateExtensions([externalExtensionsCompartement.reconfigure(extension)])

    $: if (editor && editor?.state.sliceDoc() !== blob.content) {
        editor.dispatch({
            changes: { from: 0, to: editor.state.doc.length, insert: blob.content },
        })
    }

    $: if (container && !editor) {
        editor = createEditor(container)
    }

    $: if (editor && focusedLine !== undefined) {
        editor.dispatch({
            effects: [EditorView.scrollIntoView(editor.state.doc.line(focusedLine).from, {y: "center"}), focusedLineCompartment.reconfigure(configureFocusedLine(focusedLine))],
        })
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

        :global(.focused-line) {
            background-color: var(--color-bg-2);
        }
    }
    pre {
        margin: 0;
    }
</style>
