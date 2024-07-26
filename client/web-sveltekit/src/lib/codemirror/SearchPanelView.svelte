<script lang="ts">
    import { onMount } from 'svelte'

    import { pluralize } from '$lib/common'
    import { formatShortcut } from '$lib/Hotkey'
    import Icon from '$lib/Icon.svelte'
    import KeyboardShortcut from '$lib/KeyboardShortcut.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import type { SearchPanelState } from '$lib/web'
    import Button from '$lib/wildcard/Button.svelte'
    import ButtonGroup from '$lib/wildcard/ButtonGroup.svelte'
    import Switch from '$lib/wildcard/Switch.svelte'

    import { keyboardShortcut } from './inline-search'

    export let searchPanelState: SearchPanelState
    export let onSearch: (query: string) => void
    export let findNext: () => void
    export let findPrevious: () => void
    export let setCaseSensitive: (enabled: boolean) => void
    export let setRegexp: (enabled: boolean) => void
    export let onClose: () => void
    export let setOverrideBrowserSearch: (override: boolean) => void

    /**
     * Exported for the caller to control the input element.
     */
    export function getInput(): HTMLInputElement {
        return input
    }

    let input: HTMLInputElement

    $: ({ matches, inputValue, searchQuery, currentMatchIndex } = searchPanelState)
    $: totalMatches = matches.size
    $: regexp = searchQuery.regexp
    $: caseSensitive = searchQuery.caseSensitive

    function handleOverrideBrowserSearchChange(event: Event) {
        setOverrideBrowserSearch((event.target as HTMLInputElement).checked)
    }

    onMount(() => {
        // CodeMirror doesn't focus the input element when the search panel is opened
        // the first time.
        input.focus()
        input.select()
    })
</script>

<span class="input cm-sg-search-input" class:no-match={inputValue && totalMatches === 0}>
    <input
        bind:this={input}
        placeholder="Find..."
        value={inputValue}
        autocomplete="off"
        on:input={event => onSearch(event.currentTarget.value)}
        main-field={true}
    />
    <Tooltip tooltip="{caseSensitive ? 'Disable' : 'Enable'} case sensitivity">
        <button class:enabled={caseSensitive} on:click={() => setCaseSensitive(!caseSensitive)}>
            <Icon icon={IMdiFormatLetterCase} inline aria-hidden />
        </button>
    </Tooltip>
    <Tooltip tooltip="{regexp ? 'Disable' : 'Enable'} regular expression">
        <button class:enabled={regexp} on:click={() => setRegexp(!regexp)}>
            <Icon icon={IMdiRegex} inline aria-hidden />
        </button>
    </Tooltip>
</span>
{#if matches.size > 1}
    <ButtonGroup --icon-color="var(--icon-color)">
        <Button size="sm" outline variant="secondary" on:click={findPrevious} aria-label="previous result">
            <Icon inline icon={ILucideChevronLeft} aria-hidden />
        </Button>
        <Button size="sm" outline variant="secondary" on:click={findNext} aria-label="next result">
            <Icon inline icon={ILucideChevronRight} aria-hidden />
        </Button>
    </ButtonGroup>
{/if}
{#if searchQuery.search}
    <span class="results">
        {#if currentMatchIndex !== null}
            {currentMatchIndex} of
        {/if}
        {totalMatches}
        {pluralize('result', totalMatches)}
    </span>
{/if}

<div class="actions">
    <!-- svelte-ignore a11y-label-has-associated-control
        <Switch /> renders an <input /> element
    -->
    <label>
        <Switch checked={searchPanelState.overrideBrowserSearch} on:change={handleOverrideBrowserSearchChange} />
        <KeyboardShortcut shortcut={keyboardShortcut} />
    </label>
    <Tooltip
        tooltip="When enabled, {formatShortcut(
            keyboardShortcut
        )} searches the file only. Disable to search the page, and press {formatShortcut(
            keyboardShortcut
        )} for changes to apply."
    >
        <Icon icon={ILucideInfo} inline aria-hidden />
    </Tooltip>
    <Button variant="icon" aria-label="Close" on:click={onClose}>
        <Icon icon={ILucideX} inline aria-hidden />
    </Button>
</div>

<style lang="scss">
    .input {
        display: flex;
        padding: 0 0.5rem;
        gap: 0.25rem;

        &:focus-within {
            border-color: var(--primary);
        }

        input {
            all: unset;

            color: var(--text-body);
        }

        button {
            --icon-color: currentColor;

            all: unset;

            width: 1.5rem;
            height: 1.5rem;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            color: var(--text-muted);
            cursor: pointer;
            padding: 0.125rem;
            border-radius: var(--border-radius);

            &:hover {
                color: var(--text-body);
                background-color: var(--color-bg-2);
            }

            &.enabled {
                background-color: var(--primary);
                color: var(--body-bg);
            }
        }

        &.no-match {
            input {
                color: var(--danger);
            }
        }
    }

    .results {
        color: var(--text-muted);
    }

    .actions {
        margin-left: auto;
        display: flex;
        align-items: center;
        gap: 0.5rem;

        label {
            display: inline-flex;
            align-items: center;
            gap: 0.25rem;
        }
    }
</style>
