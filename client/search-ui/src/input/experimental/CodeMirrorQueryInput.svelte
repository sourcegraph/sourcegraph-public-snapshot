<svelte:options immutable={true} />

<script lang="ts">
    import { defaultKeymap, historyKeymap, history as codemirrorHistory } from '@codemirror/commands'

    import { Compartment, EditorState, Extension, Prec } from '@codemirror/state'
    import { EditorView, keymap, placeholder as placeholderExtension } from '@codemirror/view'

    import { QueryChangeSource, QueryState, SearchPatternType } from '@sourcegraph/search'
    import { onDestroy } from 'svelte'
    import { parseInputAsQuery } from '../codemirror/parsedQuery'
    import { editorConfigFacet, Source, suggestions } from './suggestions'
    import { filterHighlight, querySyntaxHighlighting } from '../codemirror/syntax-highlighting'
    import { singleLine } from '../codemirror'
    import { mdiClose } from '@mdi/js'
    import Icon from './Icon.svelte'
    import { History } from 'history'

    export let isLightTheme: boolean
    export let patternType: SearchPatternType
    export let interpretComments: boolean
    export let queryState: QueryState
    export let onChange: (queryState: QueryState) => void
    export let onSubmit: (() => void) | undefined = undefined
    export let placeholder = ''
    export let suggestionSource: Source | undefined = undefined
    export let history: History

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

    let editor: EditorView | null = null
    let container: HTMLDivElement | null = null
    let suggestionsContainer: HTMLDivElement | null = null

    const popoverID = `searchinput-popover-${Math.floor(Math.random() * 2e6).toString(36)}`

    // Helper function to observe the current size of an element. This is used
    // to create an appropriately sized placeholder element.
    function resizeObserver(node: HTMLElement) {
        const resizeObserver = new ResizeObserver(() => {
            node.dispatchEvent(new CustomEvent('resize'))
        })

        resizeObserver.observe(node)

        return {
            destroy() {
                resizeObserver.unobserve(node)
            },
        }
    }

    // Helper function to update extensions dependent on props. Used when
    // creating the editor and to update it when the props change.
    function configureExtensions({
        patternType,
        interpretComments,
        isLightTheme,
        placeholder,
        onChange,
        onSubmit,
        suggestionsContainer,
        suggestionSource,
        history,
    }: Config): Extension {
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
                editorConfigFacet.of({ onSubmit }),
                Prec.high(
                    keymap.of([
                        {
                            key: 'Enter',
                            run() {
                                onSubmit?.()
                                return true
                            },
                        },
                        {
                            key: 'Mod-Enter',
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

        return extensions
    }

    // For simplicity we will recompute all extensions when input changes using
    // this ocmpartment
    const extensionsCompartment = new Compartment()

    function createEditor(parent: HTMLDivElement, extensions: Extension) {
        if (editor) {
            return
        }
        editor = new EditorView({
            state: EditorState.create({
                doc: queryState.query,
                extensions: [
                    EditorView.lineWrapping,
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
                    extensionsCompartment.of(extensions),
                ],
            }),
            parent,
        })
    }

    function updateEditor(extensions: Extension) {
        if (editor) {
            editor.dispatch({ effects: extensionsCompartment.reconfigure(extensions) })
        }
    }

    function updateValueIfNecessary(queryState: QueryState) {
        if (editor && queryState.changeSource !== QueryChangeSource.userInput) {
            editor.dispatch({ changes: { from: 0, to: editor.state.doc.length, insert: queryState.query } })
        }
    }

    onDestroy(() => {
        editor?.destroy()
    })

    // Used to set placeholder height to the same height as the input.
    let height: number
    function onResize(event: Event) {
        height = (event.target as HTMLElement).clientHeight
    }

    // Update editor content whenever query state changes
    $: updateValueIfNecessary(queryState)

    // Update editor configuration whenever one of these props changes
    $: updateEditor(
        configureExtensions({
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
    )

    // Create editor when container element is available
    $: if (container) {
        createEditor(
            container,
            configureExtensions({
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
        )
    }

    $: hasValue = queryState.query.length > 0

    function focus() {
        editor?.contentDOM.focus()
    }

    function handleGlobalShortcut(event: KeyboardEvent) {
        if (
            !event.defaultPrevented &&
            container &&
            event.target &&
            !container.contains(event.target as Node) &&
            event.key === '/'
        ) {
            focus()
            event.preventDefault()
        }
    }
</script>

<svelte:window on:keydown={handleGlobalShortcut} />

<div class="container">
    <div class="spacer" style="height: {height}px" />
    <div class="root">
        <div class="focus-container" use:resizeObserver on:resize={onResize}>
            <div bind:this={container} style="display: contents" />
            <!-- TODO: Consider making this a CodeMirror extension -->
            <button
                type="button"
                class:showWhenFocused={hasValue}
                on:click={() => {
                    console.log('clear')
                    onChange({ query: '' })
                }}><Icon path={mdiClose} /></button
            >
            <!-- A temporary solution for rendering a global shortcut button. Should probably be a CodeMirror extension too -->
            <button type="button" class="global-shortcut hideWhenFocused" on:click={focus}>/</button>
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
        margin: 12px;
        border: 1px solid var(--border-color-2);
        padding: 0 4px;
        min-height: 32px;
        align-items: center;

        &:focus-within {
            outline: 2px solid rgba(163, 208, 255, 1);
            outline-offset: 0px;
            border-color: var(--border-active-color);

            .hideWhenFocused {
                display: none;
            }

            .showWhenFocused {
                display: block;
            }
        }

        .global-shortcut {
            display: block;
            align-self: center;
            border: 1px solid var(--border-color-2);
            width: 24px;
        }
    }

    button {
        display: none;
        align-self: flex-start;
        padding: 0.125rem 0.25rem;
        margin: 2px;
        border: 0;
        background-color: transparent;
        border: 1px solid transparent;
        border-radius: 4px;

        &:focus {
            outline: 2px solid rgba(163, 208, 255, 1);
            outline-offset: 0px;
        }
    }

    .suggestions {
        position: relative;
        display: none;
    }
</style>
