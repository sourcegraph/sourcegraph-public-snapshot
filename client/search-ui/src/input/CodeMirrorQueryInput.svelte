<svelte:options immutable={true} />

<script context="module" lang="ts">
    import { History } from 'history'

    interface Config {
        patternType: SearchPatternType
        interpretComments: boolean
        isLightTheme: boolean
        placeholder: string
        onChange: (querySate: QueryState) => void
        onSubmit?: () => void
        suggestionsContainer: HTMLDivElement | null
        suggestionSource?: Source
        history: History
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
    import { defaultKeymap, historyKeymap, history as codemirrorHistory } from '@codemirror/commands'

    import { Compartment, EditorState, Prec } from '@codemirror/state'
    import { EditorView, keymap, placeholder as placeholderExtension } from '@codemirror/view'

    import { QueryChangeSource, QueryState, SearchPatternType } from '@sourcegraph/search'
    import { onDestroy } from 'svelte'
    import { parseInputAsQuery } from './codemirror/parsedQuery'
    import { Source, suggestions } from './codemirror/suggestions'
    import { filterHighlight, querySyntaxHighlighting } from './codemirror/syntax-highlighting'
    import { singleLine } from './codemirror'
    import { mdiClose } from '@mdi/js'
    import Icon from './codemirror/Icon.svelte'

    export let isLightTheme: boolean
    export let patternType: SearchPatternType
    export let interpretComments: boolean
    export let queryState: QueryState
    export let onChange: (queryState: QueryState) => void
    export let onSubmit: (() => void) | undefined = undefined
    export let placeholder = ''
    export let suggestionSource: Source | undefined = undefined
    export let history: History

    let editor: EditorView | null = null
    let container: HTMLDivElement | null = null
    let suggestionsContainer: HTMLDivElement | null = null

    const popoverID = `searchinput-popover-${Math.floor(Math.random() * 2 ** 50)}`

    // For simplicity we will recompute all extensions when input changes.
    const extensionsCompartment = new Compartment()

    function configureEditor(
        parent: HTMLDivElement,
        {
            patternType,
            interpretComments,
            isLightTheme,
            placeholder,
            onChange,
            suggestionsContainer,
            suggestionSource,
            history,
        }: Config
    ) {
        const extensions = [
            singleLine,
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

        if (onSubmit) {
            extensions.push(
                Prec.high(
                    keymap.of([
                        {
                            key: 'Enter',
                            run() {
                                onSubmit?.()
                                return true
                            },
                        },
                    ])
                )
            )
        }

        if (suggestionSource && suggestionsContainer) {
            extensions.push(suggestions(popoverID, suggestionsContainer, suggestionSource, history))
        }

        if (!editor) {
            editor = new EditorView({
                state: EditorState.create({
                    doc: queryState.query,
                    extensions: [
                        EditorView.contentAttributes.of({
                            role: 'combobox',
                            'aria-controls': popoverID,
                            'aria-owns': popoverID,
                            'aria-haspopup': 'grid',
                        }),
                        keymap.of(historyKeymap),
                        keymap.of(defaultKeymap),
                        codemirrorHistory(),
                        Prec.low([querySyntaxHighlighting, filterHighlight]),
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
            onSubmit,
            suggestionsContainer,
            suggestionSource,
            history,
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

    // Used to set placeholder height to the same height as the input.
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
                <button type="button" on:click={() => onChange({ query: '' })}><Icon path={mdiClose} /></button>
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
            margin: 12px;
        }
    }

    .root {
        position: absolute;
        left: 0;
        right: 0;
        top: 0;
        border-radius: 8px;
        z-index: 100;

        &:focus-within {
            background-color: var(--color-bg-1);
            box-shadow: var(--box-shadow);

            .suggestions {
                display: block;
            }
        }
    }

    .focus-container {
        display: flex;
        background-color: var(--color-bg-1);
        border-radius: 4px;
        margin: 12px 12px 0 12px;
        border: 1px solid var(--border-color-2);
        padding: 0 4px;
        min-height: 32px;
        align-items: center;

        &:focus-within {
            outline: 2px solid rgba(163, 208, 255, 1);
            outline-offset: 0px;
            border-color: var(--border-active-color);
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
        display: none;
    }
</style>
