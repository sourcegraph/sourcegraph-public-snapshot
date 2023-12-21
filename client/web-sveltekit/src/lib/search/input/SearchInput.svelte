<script lang="ts" context="module">
    // TODO(fkling): Add support for missing features
    //  - History more
    //  - Default context support
    //  - Global keyboard shortcut
    import { mdiCodeBrackets, mdiFormatLetterCase, mdiRegex } from '@mdi/js'

    import { goto, invalidate } from '$app/navigation'
    import { SearchPatternType } from '$lib/graphql-operations'
    import Icon from '$lib/Icon.svelte'
    import Tooltip from '$lib/Tooltip.svelte'

    import { submitSearch, type QueryStateStore } from '../state'
    import BaseCodeMirrorQueryInput from '$lib/search/BaseQueryInput.svelte'
    import { createSuggestionsSource } from '$lib/web'
    import { gql, query } from '$lib/graphql'
    import Suggestions from './Suggestions.svelte'
    import { user } from '$lib/stores'
    import SmartSearchToggleButton from './SmartSearchToggleButton.svelte'

    import { EditorSelection, EditorState, Prec, type Extension } from '@codemirror/state'
    import { EditorView } from '@codemirror/view'
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
    } from '$lib/branded'

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
                minWidth: 0,
                marginRight: '0.5rem',
            },
            '&.cm-editor.cm-focused': {
                outline: 'none',
            },
            '.cm-scroller': {
                overflowX: 'hidden',
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

    function graphqlQuery<T, V extends Record<string, any>>(request: string, variables: V) {
        return query<T, V>(gql(request), variables)
    }
</script>

<script lang="ts">
    export let queryState: QueryStateStore
    export let showSmartSearchButton = false

    export function focus() {
        input?.focus()
    }

    const popoverID = 'main-search'

    let input: BaseCodeMirrorQueryInput
    let editor: EditorView | null = null
    let mode = ''
    let suggestionsPaddingTop = 0
    let suggestionsUI: Extension = []

    $: patternType = $queryState.patternType
    $: regularExpressionEnabled = patternType === SearchPatternType.regexp
    $: structuralEnabled = patternType === SearchPatternType.structural
    $: extension = [
        suggestions({
            id: popoverID,
            source: createSuggestionsSource({
                valueType: patternType === SearchPatternType.newStandardRC1 ? 'glob' : 'regex',
                graphqlQuery,
                authenticatedUser: $user,
                isSourcegraphDotCom: false,
            }),
            navigate: goto,
        }),
        suggestionsUI,
        staticExtensions,
    ]

    function setOrUnsetPatternType(patternType: SearchPatternType): void {
        queryState.setPatternType(currentPatternType =>
            currentPatternType === patternType ? SearchPatternType.standard : patternType
        )
    }

    async function handleSubmit(event: Event) {
        event.preventDefault()
        const currentQueryState = $queryState
        await invalidate(`query:${$queryState.query}--${$queryState.caseSensitive}`)
        submitSearch(currentQueryState)
    }

    function selectOption(event: { detail: { option: Option; action: Action } }): void {
        if (editor) {
            applyAction(editor, event.detail.action, event.detail.option, 'mouse')
            window.requestAnimationFrame(() => editor?.focus())
        }
    }
</script>

<form
    bind:clientHeight={suggestionsPaddingTop}
    class="search-box"
    action="/search"
    method="get"
    on:submit={handleSubmit}
>
    <input class="hidden" value={$queryState.query} name="q" />
    <div class="focus-container">
        <div class="mode-switcher" />
        <BaseCodeMirrorQueryInput
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
            <Tooltip tooltip={`${$queryState.caseSensitive ? 'Disable' : 'Enable'} case sensitivity`}>
                <button
                    class="toggle icon"
                    type="button"
                    class:active={$queryState.caseSensitive}
                    on:click={() => queryState.setCaseSensitive(caseSensitive => !caseSensitive)}
                >
                    <Icon svgPath={mdiFormatLetterCase} inline />
                </button>
            </Tooltip>
            <Tooltip tooltip="{regularExpressionEnabled ? 'Disable' : 'Enable'} regular expression">
                <button
                    class="toggle icon"
                    type="button"
                    class:active={regularExpressionEnabled}
                    on:click={() => setOrUnsetPatternType(SearchPatternType.regexp)}
                >
                    <Icon svgPath={mdiRegex} inline />
                </button>
            </Tooltip>
            <Tooltip tooltip="{structuralEnabled ? 'Disable' : 'Enable'} structural search">
                <button
                    class="toggle icon"
                    type="button"
                    class:active={structuralEnabled}
                    on:click={() => setOrUnsetPatternType(SearchPatternType.structural)}
                >
                    <Icon svgPath={mdiCodeBrackets} inline />
                </button>
            </Tooltip>
            {#if showSmartSearchButton}
                <span class="divider" />
                <SmartSearchToggleButton {queryState} />
            {/if}
        {/if}
    </div>
    <div class="suggestions" style:padding-top="{suggestionsPaddingTop}px">
        <Suggestions bind:this={suggestionsUI} on:select={selectOption} />
    </div>
</form>

<style lang="scss">
    @use '$lib/breakpoints';

    form {
        width: 100%;
        position: relative;
        // Necessary to ensure that the search input (especially the suggestions) are rendered above sticky headers
        // in the search results page ("position: sticky" creates a new stacking context).
        z-index: 1;
        padding: 0.75rem;

        &:focus-within {
            .suggestions {
                display: block;
            }
        }
    }

    .hidden {
        display: none;
    }

    .focus-container {
        flex: 1;
        display: flex;
        align-items: center;
        min-height: 32px;
        padding: 0 0.25rem;
        border-radius: 4px;
        border: 1px solid var(--border-color-2);
        background-color: var(--input-bg);
        position: relative;
        z-index: 1;

        &:focus-within {
            @media (--sm-breakpoint-up) {
                outline: 2px solid var(--primary-2);
                outline-offset: 0;
                border-color: var(--primary-2);
            }
        }

        @media (--xs-breakpoint-down) {
            flex-direction: column;
            align-items: start;
            padding: 0.5rem;
            gap: 0.5rem;
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

    button.toggle {
        width: 1.5rem;
        height: 1.5rem;
        cursor: pointer;
        border-radius: var(--border-radius);
        display: flex;
        align-items: center;
        justify-content: center;

        &.active {
            background-color: var(--primary);
            color: var(--light-text);
        }

        :global(svg) {
            transform: scale(1.172);
        }
    }

    .divider {
        width: 1px;
        height: 1rem;
        background-color: var(--border-color-2);
        margin: 0 0.5rem;
    }

    button.icon {
        padding: 0;
        margin: 0;
        border: 0;
        background-color: transparent;
        cursor: pointer;
    }

    .popover-content {
        input {
            margin-left: 0;
        }

        label {
            max-width: 17rem;
            display: flex;
            cursor: pointer;
            padding: 0.5rem 1rem;
            border-top: 1px solid var(--border-color);
        }
    }
</style>
