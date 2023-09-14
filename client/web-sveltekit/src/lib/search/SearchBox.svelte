<script lang="ts">
    import { mdiClose, mdiCodeBrackets, mdiFormatLetterCase, mdiLightningBolt, mdiMagnify, mdiRegex } from '@mdi/js'

    import { invalidate } from '$app/navigation'
    import { SearchPatternType } from '$lib/graphql-operations'
    import Icon from '$lib/Icon.svelte'
    import Popover from '$lib/Popover.svelte'
    import Tooltip from '$lib/Tooltip.svelte'

    import CodeMirrorQueryInput from './CodeMirrorQueryInput.svelte'
    import { SearchMode, submitSearch, type QueryStateStore } from './state'

    export let queryState: QueryStateStore
    export let autoFocus = false

    export function focus() {
        input?.focus()
    }

    let input: CodeMirrorQueryInput

    $: regularExpressionEnabled = $queryState.patternType === SearchPatternType.regexp
    $: structuralEnabled = $queryState.patternType === SearchPatternType.structural
    $: smartEnabled = $queryState.searchMode === SearchMode.SmartSearch

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
</script>

<form class="search-box" action="/search" method="get" on:submit={handleSubmit}>
    <input class="hidden" value={$queryState.query} name="q" />
    <span class="context"
        ><span class="search-filter-keyword">context:</span><span>{$queryState.searchContext}</span></span
    >
    <span class="divider" />
    <CodeMirrorQueryInput
        bind:this={input}
        {autoFocus}
        placeholder="Search for code or files"
        queryState={$queryState}
        on:change={event => queryState.setQuery(event.detail.query)}
        on:submit={handleSubmit}
        patternType={$queryState.patternType}
    />
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
    <Tooltip tooltip={`${regularExpressionEnabled ? 'Disable' : 'Enable'} regular expression`}>
        <button
            class="toggle icon"
            type="button"
            class:active={regularExpressionEnabled}
            on:click={() => setOrUnsetPatternType(SearchPatternType.regexp)}
        >
            <Icon svgPath={mdiRegex} inline />
        </button>
    </Tooltip>
    <Tooltip tooltip={`${structuralEnabled ? 'Disable' : 'Enable'} structural search`}>
        <button
            class="toggle icon"
            type="button"
            class:active={structuralEnabled}
            on:click={() => setOrUnsetPatternType(SearchPatternType.structural)}
        >
            <Icon svgPath={mdiCodeBrackets} inline />
        </button>
    </Tooltip>
    <span class="divider" />
    <Popover let:registerTrigger let:toggle>
        <Tooltip tooltip="Smart search {smartEnabled ? 'enabled' : 'disabled'}">
            <button
                class="toggle icon"
                type="button"
                class:active={smartEnabled}
                on:click={() => toggle()}
                use:registerTrigger
            >
                <Icon svgPath={mdiLightningBolt} inline />
            </button>
        </Tooltip>
        <div slot="content" class="popover-content" let:toggle>
            {@const delayedClose = () => setTimeout(() => toggle(false), 100)}
            <div class="d-flex align-items-center px-3 py-2">
                <h4 class="m-0 mr-auto">SmartSearch</h4>
                <button class="icon" type="button" on:click={() => toggle(false)}>
                    <Icon svgPath={mdiClose} inline />
                </button>
            </div>
            <div>
                <label class="d-flex align-items-start">
                    <input
                        type="radio"
                        name="mode"
                        value="smart"
                        checked={smartEnabled}
                        on:click={() => {
                            queryState.setMode(SearchMode.SmartSearch)
                            delayedClose()
                        }}
                    />
                    <span class="d-flex flex-column ml-1">
                        <span>Enable</span>
                        <small class="text-muted"
                            >Suggest variations of your query to find more results that may relate.</small
                        >
                    </span>
                </label>
                <label class="d-flex align-items-start">
                    <input
                        type="radio"
                        name="mode"
                        value="precise"
                        checked={!smartEnabled}
                        on:click={() => {
                            queryState.setMode(SearchMode.Precise)
                            delayedClose()
                        }}
                    />
                    <span class="d-flex flex-column ml-1">
                        <span>Disable</span>
                        <small class="text-muted">Only show results that previsely match your query.</small>
                    </span>
                </label>
            </div>
        </div>
    </Popover>
    <button class="submit">
        <Icon aria-label="search" svgPath={mdiMagnify} inline />
    </button>
</form>

<style lang="scss">
    form {
        width: 100%;
        display: flex;
        align-items: center;
        background-color: var(--color-bg-1);
        padding-left: 0.5rem;
        border-top-left-radius: 5px;
        border-bottom-left-radius: 5px;
        border: 1px solid var(--border-color);
        margin: 2px;

        &:focus-within {
            outline: 0;
            box-shadow: var(--focus-box-shadow);
        }
    }

    .hidden {
        display: none;
    }

    .context {
        font-family: var(--code-font-family);
        font-size: 0.75rem;
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

    button.submit {
        margin-left: 1rem;
        padding: 0.5rem 1rem;
        border-top-right-radius: 5px;
        border-bottom-right-radius: 5px;
        background-color: var(--primary);
        border: none;
        color: var(--light-text);
        cursor: pointer;

        &:hover {
            background-color: var(--primary-3);
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
