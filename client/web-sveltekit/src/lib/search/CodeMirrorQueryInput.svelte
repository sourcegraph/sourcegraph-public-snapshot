<script lang="ts">
    import { closeCompletion } from '@codemirror/autocomplete'
    import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
    import { Compartment, EditorState, Prec } from '@codemirror/state'
    import { EditorView, keymap, placeholder as placeholderExtension } from '@codemirror/view'
    import { createEventDispatcher } from 'svelte'

    import { browser } from '$app/environment'
    import { goto } from '$app/navigation'
    import { createDefaultSuggestions, multiline, parseInputAsQuery, querySyntaxHighlighting } from '$lib/branded'
    import type { SearchPatternType } from '$lib/graphql-operations'
    import { fetchStreamSuggestions, QueryChangeSource, type QueryState } from '$lib/shared'

    import { defaultTheme } from './codemirror/theme'

    export let queryState: QueryState
    export let patternType: SearchPatternType
    export let interpretComments: boolean = false
    export let placeholder: string = ''
    export let autoFocus: boolean = false

    export function focus() {
        editor?.focus()
        editor?.dispatch({ selection: { anchor: editor.state.doc.length } })
    }

    const dispatch = createEventDispatcher<{ change: QueryState; submit: void }>()
    let container: HTMLDivElement | null = null
    let editor: EditorView | null = null

    let dynamicExtensions = new Compartment()

    interface ExtensionConfig {
        patternType: SearchPatternType
        interpretComments: boolean
        placeholder: string
    }

    function configureExtensions(config: ExtensionConfig) {
        if (!browser) {
            return []
        }
        const extensions = [
            parseInputAsQuery({ patternType: config.patternType, interpretComments: config.interpretComments }),
            createDefaultSuggestions({
                fetchSuggestions: query => fetchStreamSuggestions(query),
                isSourcegraphDotCom: false,
                navigate: url => goto(url.toString()),
            }),
        ]

        if (config.placeholder) {
            // Passing a DOM element instead of a string makes the CodeMirror
            // extension set aria-hidden="true" on the placeholder, which is
            // what we want.
            const element = document.createElement('span')
            element.append(document.createTextNode(placeholder))
            extensions.push(placeholderExtension(element))
        }

        return extensions
    }

    function updateExtensions(config: ExtensionConfig) {
        if (editor) {
            editor.dispatch({ effects: dynamicExtensions.reconfigure(configureExtensions(config)) })
        }
    }

    function createEditor(container: HTMLDivElement): EditorView {
        const extensions = [
            defaultTheme,
            dynamicExtensions.of(configureExtensions({ interpretComments, patternType, placeholder })),
            Prec.high(
                keymap.of([
                    {
                        key: 'Enter',
                        run(view) {
                            closeCompletion(view)
                            dispatch('submit')
                            return true
                        },
                    },
                ])
            ),
            multiline(false),
            EditorView.updateListener.of(update => {
                const { state } = update
                if (update.docChanged) {
                    dispatch('change', {
                        query: state.sliceDoc(),
                        changeSource: QueryChangeSource.userInput,
                    })
                }
                if (update.focusChanged && !update.view.hasFocus) {
                    closeCompletion(update.view)
                }
            }),
            keymap.of(historyKeymap),
            keymap.of(defaultKeymap),
            history(),
            // themeExtension.of(EditorView.darkTheme.of(isLightTheme === false)),
            // queryDiagnostic(),
            // The precedence of these extensions needs to be decreased
            // explicitly, otherwise the diagnostic indicators will be
            // hidden behind the highlight background color
            Prec.low([
                //tokenInfo(),
                //highlightFocusedFilter,
                // It baffels me but the syntax highlighting extension has
                // to come after the highlight current filter extension,
                // otherwise CodeMirror keeps steeling the focus.
                // See https://github.com/sourcegraph/sourcegraph/issues/38677
                querySyntaxHighlighting,
            ]),
        ]

        const view = new EditorView({
            state: EditorState.create({ doc: queryState.query, extensions }),
            parent: container,
        })
        return view
    }

    $: if (editor && editor.state.sliceDoc() !== queryState.query) {
        editor.dispatch({
            changes: { from: 0, to: editor.state.doc.length, insert: queryState.query },
        })
    }

    $: updateExtensions({ placeholder, patternType, interpretComments })

    $: if (container && !editor) {
        editor = createEditor(container)
        if (autoFocus) {
            window.requestAnimationFrame(() => editor!.focus())
        }
    }
</script>

{#if browser}
    <div bind:this={container} class="root test-query-input test-editor" data-editor="codemirror6" />
{:else}
    <div class="root">
        <input value={queryState.query} {placeholder} />
    </div>
{/if}

<style lang="scss">
    input {
        border: 0;
        font-family: var(--code-font-family);
        font-size: var(--code-font-size);
        width: 100%;
    }

    .root {
        flex: 1;
        box-sizing: border-box;
        background-color: var(--color-bg-1);
        min-width: 0;

        :global(.cm-editor) {
            // Codemirror shows a focus ring by default. Since we handle that
            // differently, disable it here.
            outline: none !important;

            :global(.cm-scroller) {
                // Codemirror shows a vertical scroll bar by default (when
                // overflowing). This disables it.
                overflow-x: hidden;
            }

            :global(.cm-content) {
                caret-color: var(--search-query-text-color);
                font-family: var(--code-font-family);
                font-size: var(--code-font-size);
                color: var(--search-query-text-color);
                // Disable default padding
                padding: 0;

                &:global(.focus-visible) {
                    box-shadow: none;
                }
            }
        }

        :global(.cm-line) {
            // Disable default padding
            padding: 0;
        }

        :global(.cm-placeholder) {
            // CodeMirror uses display: inline-block by default, but that causes
            // Chrome to render a larger cursor if the placeholder holder spans
            // multiple lines. Firefox doesn't have this problem (but
            // setting display: inline doesn't seem to have a negative effect
            // either)
            display: inline;
            // Once again, Chrome renders the placeholder differently than
            // Firefox. CodeMirror sets 'word-break: break-word' (which is
            // deprecated) and 'overflow-wrap: anywhere' but they don't seem to
            // have an effect in Chrome (at least not in this instance).
            // Setting 'word-break: break-all' explicitly makes appearances a
            // bit better for example queries with long tokens.
            word-break: break-all;
        }
        // .placeholder needs to explicilty have the same background color because it
        // appears to be placed outside of .focusedFilter rather than within it.
        :global(.placeholder),
        :global(.focusedFilter) {
            background-color: var(--gray-02);

            :global(.theme-dark) & {
                background-color: var(--gray-08);
            }
        }
    }
</style>
