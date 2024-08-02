<script lang="ts" context="module">
    // TODO(fkling): Add support for missing features
    //  - History more
    //  - Default context support
    //  - Global keyboard shortcut

    import { EditorSelection, EditorState, Prec, type Extension } from '@codemirror/state'
    import { EditorView } from '@codemirror/view'

    import { goto, invalidate } from '$app/navigation'
    import {
        type Option,
        type Action,
        applyAction,
        modeScope,
        queryDiagnostic,
        overrideContextOnPaste,
        filterDecoration,
        tokenInfo,
        placeholder,
        showWhenEmptyWithoutContext,
        suggestions,
        searchHistoryExtension,
        onModeChange,
        setMode,
    } from '$lib/branded'
    import { query, type DocumentInput } from '$lib/graphql'
    import { SearchPatternType } from '$lib/graphql-operations'
    import Icon from '$lib/Icon.svelte'
    import BaseCodeMirrorQueryInput from '$lib/search/BaseQueryInput.svelte'
    import { user, settings } from '$lib/stores'
    import Tooltip from '$lib/Tooltip.svelte'
    import { createSuggestionsSource } from '$lib/web'

    import { type QueryStateStore, getQueryURL, QueryState } from '../state'

    import { createRecentSearchesStore } from './recentSearches'
    import Suggestions from './Suggestions.svelte'

    const placeholderText = 'Search for code or files...'

    // Creates extensions that don't depend on props
    const position0 = EditorSelection.single(0)
    const staticExtensions: Extension = [
        placeholder(placeholderText, showWhenEmptyWithoutContext),
        EditorState.transactionFilter.of(transaction => {
            // This is a hacky way to "fix" the cursor position when the input receives
            // focus by clicking outside of it in Chrome.
            // Debugging has revealed that in such a case the transaction has a user event
            // 'select', the new selection is set to `0` and 'scrollIntoView' is 'false'.
            // This is different from other events that change the cursor position:
            // - Clicking on text inside the input (whether focused or not) will be a 'select.pointer'
            //   user event.
            // - Moving the cursor with arrow keys will be a 'select' user event but will also set
            //   'scrollIntoView' to 'true'
            // - Entering new characters will be of user type 'input'
            // - Selecting a text range will be of user type 'select.pointer'
            // - Tabbing to the input seems to only trigger a 'select' user event transaction when
            //   the user clicked outside the input (also only in Chrome, this transaction doesn't
            //   occur in Firefox)

            if (
                !transaction.isUserEvent('select.pointer') &&
                transaction.isUserEvent('select') &&
                !transaction.scrollIntoView &&
                transaction.selection?.eq(position0)
            ) {
                return [transaction, { selection: EditorSelection.single(transaction.newDoc.length) }]
            }
            return transaction
        }),
        modeScope([queryDiagnostic(), overrideContextOnPaste], [null]),
        Prec.low([modeScope([tokenInfo(), filterDecoration], [null])]),
        EditorView.theme({
            '&': {
                flex: 1,
                backgroundColor: 'var(--input-bg)',
                borderRadius: 'var(--border-radius)',
                borderColor: 'var(--border-color)',
                // To ensure that the input doesn't overflow the parent
                maxWidth: '100%',
            },
            '&.cm-editor.cm-focused': {
                outline: 'none',
            },
            '.cm-scroller': {
                overflowX: 'hidden',
                lineHeight: '1.6',
            },
            '.cm-content': {
                paddingLeft: '0.25rem',
            },
            '.cm-content.focus-visible': {
                boxShadow: 'none',
            },
            '.sg-decorated-token-hover': {
                borderRadius: '3px',
            },
            '.sg-query-filter-placeholder': {
                color: 'var(--text-muted)',
                fontStyle: 'italic',
            },
        }),
    ]

    async function graphqlQuery<T, V extends Record<string, any>>(request: DocumentInput, variables: V) {
        const result = await query<T, V>(request, variables)
        // This is a hack to make urlq work with the API that createSuggestionsSource expects
        return result.data ?? ({} as any)
    }

    export const enum Style {
        /**
         * Reduced paddings and margins.
         */
        Compact = 1 << 0,
        /**
         * No border around the input.
         */
        NoBorder = 1 << 1,
    }
</script>

<script lang="ts">
    import { registerHotkey } from '$lib/Hotkey'

    export let autoFocus = false
    export let style: Style | undefined = undefined
    export let queryState: QueryStateStore
    export let onSubmit: (state: QueryState) => void = () => {}
    export let extension: Extension = []

    export function focus() {
        input?.focus()
    }

    const popoverID = 'main-search'
    const recentSearches = createRecentSearchesStore()

    let input: BaseCodeMirrorQueryInput
    let editor: EditorView | null = null
    let mode = ''
    let suggestionsPaddingTop = 0
    let suggestionsUI: Extension = []

    // When autofocus is set we only show suggestions when the user has interacted with the input
    let userHasInteracted = !autoFocus
    const hasInteractedExtension = EditorView.updateListener.of(update => {
        if (!userHasInteracted) {
            if (
                update.transactions.some(
                    tr => tr.isUserEvent('select') || tr.isUserEvent('input') || tr.isUserEvent('delete')
                )
            ) {
                userHasInteracted = true
            }
        }
    })
    const searchHistory = searchHistoryExtension({
        mode: {
            name: 'History',
            placeholder: 'Filter history',
        },
        source: () => $recentSearches ?? [],
        submitQuery: (query, view) => {
            void submitQuery($queryState.setQuery(query))
            view.contentDOM.blur()
        },
    })

    registerHotkey({
        keys: { key: '/' },
        handler: () => {
            // If the search input doesn't have focus, focus it
            // and disallow `/` symbol populate the input value
            focus()
            return false
        },
    })

    $: regularExpressionEnabled = $queryState.patternType === SearchPatternType.regexp
    $: structuralEnabled = $queryState.patternType === SearchPatternType.structural
    $: suggestionsExtension = suggestions({
        id: popoverID,
        source: createSuggestionsSource({
            graphqlQuery,
            authenticatedUser: $user,
            isSourcegraphDotCom: false,
        }),
        navigate: goto,
    })

    $: extension = [
        extension,
        onModeChange((_view, newMode) => (mode = newMode ?? '')),
        hasInteractedExtension,
        suggestionsExtension,
        suggestionsUI,
        searchHistory,
        staticExtensions,
    ]

    function setOrUnsetPatternType(patternType: SearchPatternType): void {
        queryState.setPatternType(currentPatternType =>
            currentPatternType === patternType ? getUnselectedPatternType() : patternType
        )
    }

    // When a toggle is unset, we revert back to the default pattern type. However, if the default pattern type
    // is regexp, we should revert to keyword instead (otherwise it's not possible to disable the toggle).
    function getUnselectedPatternType(): SearchPatternType {
        const defaultPatternType =
            ($settings?.['search.defaultPatternType'] as SearchPatternType) ?? SearchPatternType.keyword
        return defaultPatternType === SearchPatternType.regexp ? SearchPatternType.keyword : defaultPatternType
    }

    async function submitQuery(state: QueryState): Promise<void> {
        const url = getQueryURL(state)
        // This ensures that the same query can be resubmitted from the search input. Without
        // this, SvelteKit will not re-run the loader because the URL hasn't changed.
        await invalidate(`search:${url}`)
        void goto(url)

        // Reset interaction state since after success submit we should hide
        // suggestions UI but still keep focus on input, after user interacts with
        // search input again we show suggestion panel
        userHasInteracted = false
    }

    async function handleSubmit(event: Event) {
        event.preventDefault()
        if (!mode) {
            // Only submit query if you are not in history mode
            onSubmit($queryState)
            void submitQuery($queryState)
        }
    }

    function selectOption(event: { detail: { option: Option; action: Action } }): void {
        if (editor) {
            applyAction(editor, event.detail.action, event.detail.option, 'mouse')
            window.requestAnimationFrame(() => editor?.focus())
        }
    }

    function toggleMode() {
        if (editor) {
            setMode(editor, currentMode => (currentMode === 'History' ? null : 'History'))
            editor.focus()
            // This ensures that history suggestions are shown after the button was pressed,
            // before the user has interacted with the input in any other way.
            userHasInteracted = true
        }
    }
</script>

<form
    method="get"
    action="/search"
    class="search-box"
    class:compact={style && style & Style.Compact}
    class:no-border={style && style & Style.NoBorder}
    on:submit={handleSubmit}
    bind:clientHeight={suggestionsPaddingTop}
>
    <input class="hidden" value={$queryState.query} name="q" />
    <div class="focus-container" class:userHasInteracted>
        <div class="mode-switcher" class:active={!!mode}>
            <Tooltip tooltip="Recent searches">
                <button class="icon" type="button" on:click={toggleMode}>
                    <Icon icon={ILucideHistory} inline aria-hidden />
                    {#if mode}
                        <span>{mode}:</span>
                    {/if}
                </button>
            </Tooltip>
        </div>
        <BaseCodeMirrorQueryInput
            {autoFocus}
            bind:this={input}
            bind:view={editor}
            placeholder="Search for code or files"
            value={$queryState.query}
            on:change={event => queryState.setQuery(event.detail)}
            on:enter={handleSubmit}
            patternType={$queryState.patternType}
            interpretComments={false}
            {extension}
        />
        {#if !mode}
            <span class="actions">
                <Tooltip tooltip={`${$queryState.caseSensitive ? 'Disable' : 'Enable'} case sensitivity`}>
                    <button
                        class="toggle icon"
                        type="button"
                        class:active={$queryState.caseSensitive}
                        on:click={() => queryState.setCaseSensitive(caseSensitive => !caseSensitive)}
                    >
                        <Icon icon={IMdiFormatLetterCase} inline aria-hidden />
                    </button>
                </Tooltip>
                <Tooltip tooltip="{regularExpressionEnabled ? 'Disable' : 'Enable'} regular expression">
                    <button
                        class="toggle icon"
                        type="button"
                        title="regexp toggle"
                        class:active={regularExpressionEnabled}
                        on:click={() => setOrUnsetPatternType(SearchPatternType.regexp)}
                    >
                        <Icon icon={IMdiRegex} inline aria-hidden />
                    </button>
                </Tooltip>
                {#if structuralEnabled}
                    <Tooltip tooltip="Disable structural search">
                        <button
                            class="toggle icon"
                            type="button"
                            class:active={structuralEnabled}
                            on:click={() => setOrUnsetPatternType(SearchPatternType.structural)}
                        >
                            <Icon icon={ILucideBrackets} inline aria-hidden />
                        </button>
                    </Tooltip>
                {/if}
            </span>
        {/if}
    </div>
    <div class="suggestions" style:padding-top="{suggestionsPaddingTop}px">
        <Suggestions bind:this={suggestionsUI} on:select={selectOption} />
    </div>
</form>

<style lang="scss">
    form {
        isolation: isolate;
        width: 100%;
        position: relative;
        padding: 0.75rem;

        &:focus-within {
            .userHasInteracted + .suggestions {
                display: block;
            }
        }

        &.compact {
            padding: 0.25rem;
            margin: -0.25rem;
            width: calc(100% + 0.5rem);
        }
    }

    .hidden {
        display: none;
    }

    .focus-container {
        --gap: 0.25rem;

        flex: 1;
        min-height: 32px;
        padding: 0 0.25rem;
        border-radius: 4px;
        border: 1px solid var(--border-color-2);
        background-color: var(--input-bg);
        position: relative;
        // This is necessary to ensure that the input is shown above the suggestions container
        z-index: 1;

        display: grid;
        grid-template-columns: min-content 1fr auto;
        grid-template-areas: 'mode-switcher input actions';
        align-items: center;
        gap: var(--gap);

        .no-border & {
            border: none;
            border-radius: 0;
        }

        :global([data-query-input]) {
            grid-area: input;
        }

        &:focus-within {
            @media (--sm-breakpoint-up) {
                outline: 2px solid var(--primary-2);
                outline-offset: 0;
                border-color: var(--primary-2);
            }
        }

        @media (--mobile) {
            --gap: 0.5rem;

            grid-template-columns: min-content 1fr;
            grid-template-areas: 'input input' 'mode-switcher actions';
            padding: 0.5rem;

            :global(.cm-content) {
                white-space: break-spaces;
                word-break: break-word;
                overflow-wrap: anywhere;
                flex-shrink: 1;
            }

            :global([data-query-input]) {
                border-radius: 4px;
                border: 1px solid var(--border-color-2);
                background-color: var(--input-bg);
                padding: inherit;
                width: 100%;

                &:focus-within {
                    outline: 2px solid var(--primary-2);
                    outline-offset: 0;
                    border-color: var(--primary-2);
                }
            }
        }
    }

    button.icon {
        padding: 0;
        margin: 0;
        border: 0;
        background-color: transparent;
        cursor: pointer;
    }

    .actions {
        grid-area: actions;

        button.toggle {
            --icon-color: var(--text-body);

            width: 1.5rem;
            height: 1.5rem;
            cursor: pointer;
            border-radius: var(--border-radius);
            display: inline-flex;
            align-items: center;
            justify-content: center;

            &.active {
                background-color: var(--primary);
                color: var(--light-text);
                --icon-color: currentColor;
            }

            :global(svg) {
                transform: scale(1.172);
            }
        }
    }

    .mode-switcher {
        grid-area: mode-switcher;

        display: flex;
        align-items: center;
        padding-right: var(--gap);
        border-right: 1px solid var(--border-color-2);
        font-family: var(--code-font-family);
        font-size: var(--code-font-size);
        --color: var(--text-muted);

        button {
            --icon-color: currentColor;

            padding: 0.0625rem 0.125rem;
            color: var(--color);
            border-radius: var(--border-radius);
            font-size: 0.875rem;
            display: flex;
            align-items: center;
            gap: 0.25rem;

            &:hover {
                background-color: var(--color-bg-2);
            }

            span {
                font-family: var(--code-font-family);
                font-size: var(--code-font-size);
            }
        }

        &.active {
            --color: var(--logo-purple);
            padding: 0;
            border: 0;
        }

        @media (--xs-breakpoint-down) {
            border: 0;

            span {
                display: none;
            }
        }
    }

    .suggestions {
        display: none;
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        border-radius: 8px;
        background-color: var(--color-bg-1);
        box-shadow: 0 10px 50px rgba(0, 0, 0, 0.15);

        // Set a default paddings to the suggestion panel (see Suggestions.module.scss)
        --suggestions-padding: 0.75rem;

        :global(.theme-dark) & {
            box-shadow: 0 10px 60px rgba(0, 0, 0, 0.8);
        }
    }
</style>
