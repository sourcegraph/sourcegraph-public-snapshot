<svelte:options immutable={true} />

<script context="module" lang="ts">
    interface Config {
        patternType: SearchPatternType
        interpretComments: boolean
        isLightTheme: boolean
        placeholder: string
        onChange: (querySate: QueryState) => void
        suggestionsContainer: HTMLDivElement | null
    }

    function resizeObserver(node: HTMLElement) {
        const resizeObserver = new ResizeObserver(entries => {
            node.dispatchEvent(new CustomEvent('resize'))
        })

        resizeObserver.observe(node)

        return {
            destroy() {
                resizeObserver.unobserve(node)
            },
        }
    }
</script>

<script lang="ts">
    import { defaultKeymap, historyKeymap, history } from '@codemirror/commands'

    import { Compartment, EditorState, Prec } from '@codemirror/state'
    import { EditorView, keymap, placeholder as placeholderExtension } from '@codemirror/view'

    import { QueryChangeSource, QueryState, SearchPatternType } from '@sourcegraph/search'
    import { onDestroy } from 'svelte'
    import { parseInputAsQuery } from './codemirror/parsedQuery'
    import { suggestions } from './codemirror/suggestions'
    import { querySyntaxHighlighting } from './codemirror/syntax-highlighting'

    export let isLightTheme: boolean
    export let patternType: SearchPatternType
    export let interpretComments: boolean
    export let queryState: QueryState
    export let onChange: (queryState: QueryState) => void
    export let placeholder = ''

    let editor: EditorView | null = null
    let container: HTMLDivElement | null = null
    let suggestionsContainer: HTMLDivElement | null = null

    // For simplicity we will recompute all extensions when input changes.
    const extensionsCompartment = new Compartment()

    function configureEditor(
        parent: HTMLDivElement,
        { patternType, interpretComments, isLightTheme, placeholder, onChange, suggestionsContainer }: Config
    ) {
        const extensions = [
            EditorView.darkTheme.of(isLightTheme === false),
            parseInputAsQuery({ patternType, interpretComments }),
            EditorView.updateListener.of(update => {
                if (update.docChanged) {
                    onChange({
                        query: update.state.sliceDoc(),
                        changeSource: QueryChangeSource.userInput,
                    })
                }
            }),
        ]

        if (placeholder) {
            // Passing a DOM element instead of a string makes the CodeMirror
            // extension set aria-hidden="true" on the placeholder, which is
            // what we want.
            const element = document.createElement('span')
            element.append(document.createTextNode(placeholder))
            extensions.push(placeholderExtension(element))
        }

        if (suggestionsContainer) {
            extensions.push(suggestions(suggestionsContainer, []))
        }

        if (!editor) {
            editor = new EditorView({
                state: EditorState.create({
                    doc: queryState.query,
                    extensions: [
                        keymap.of(historyKeymap),
                        keymap.of(defaultKeymap),
                        history(),
                        Prec.low([querySyntaxHighlighting]),
                        extensionsCompartment.of(extensions),
                        EditorView.theme({
                            '&': {
                                flex: 1,
                                backgroundColor: 'var(--input-bg)',
                                borderRadius: 'var(--border-radius)',
                                borderColor: 'var(--border-color)',
                            },
                            '&.cm-editor.cm-focused': {
                                outline: 'none',
                            },
                            '.cm-content': {
                                caretColor: 'var(--search-query-text-color)',
                                fontFamily: 'var(--code-font-family)',
                                fontSize: 'var(--code-font-size)',
                                color: 'var(--search-query-text-color)',
                            },
                        }),
                    ],
                }),
                parent,
            })
        } else {
            editor.dispatch({ effects: extensionsCompartment.reconfigure(extensions) })
        }
    }

    // Helper function to create reactive statements without referencing editor
    // directly so that those blocks are only executed when non-editor
    // dependencies change
    function getEditor() {
        return editor
    }

    onDestroy(() => {
        editor?.destroy()
    })

    // Update editor configuration whenever one of these props changes
    $: if (container) {
        configureEditor(container, {
            patternType,
            interpretComments,
            isLightTheme,
            placeholder,
            onChange,
            suggestionsContainer,
        })
    }

    // Update editor content whenever query state changes
    $: {
        const editor = getEditor()
        if (editor && queryState.changeSource !== QueryChangeSource.userInput) {
            editor.dispatch({ changes: { from: 0, to: editor.state.doc.length, insert: queryState.query } })
        }
    }

    $: hasValue = queryState.query.length > 0
    let height: number
    function onResize(event: Event) {
        height = (event.target as HTMLElement).clientHeight
    }
</script>

<div class="container">
    <div class="spacer" style="height: {height}px" />
    <div class="root">
        <div class="focus-container" use:resizeObserver on:resize={onResize}>
            <div bind:this={container} style="display: contents" />
            {#if hasValue}
                <button type="button" on:click={() => onChange({ query: '' })}>X</button>
            {/if}
        </div>
        <div bind:this={suggestionsContainer} class="suggestions" />
    </div>
</div>

<style lang="scss">
    .container {
        flex: 1;
        position: relative;

        .spacer {
            margin: 0.5rem;
        }
    }

    .root {
        position: absolute;
        left: 0;
        right: 0;
        top: 0;
        padding: 0.5rem;
        border-radius: var(--border-radius);
        z-index: 100;

        &:focus-within {
            background-color: var(--color-bg-1);
            box-shadow: var(--box-shadow);
        }
    }

    .focus-container {
        width: 100%;
        display: flex;
        background-color: var(--color-bg-1);
        border-radius: var(--border-radius);

        &:focus-within {
            outline: 2px solid var(--primary-2);
        }
    }

    button {
        padding: 0.125rem 0.25rem;
        margin: 0;
        border: 0;
        background-color: transparent;
    }

    .suggestions {
        position: relative;
    }
</style>
