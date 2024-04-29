<script lang="ts">
    import { tick } from 'svelte'
    import { mdiMagnify } from '@mdi/js'
    import { createDialog } from '@melt-ui/svelte'

    import Icon from '$lib/Icon.svelte'
    import SearchInput from '$lib/search/input/SearchInput.svelte'
    import { QueryState, queryStateStore } from '$lib/search/state'
    import { settings } from '$lib/stores'
    import { repositoryInsertText } from '$lib/shared'
    import { SVELTE_LOGGER, SVELTE_TELEMETRY_EVENTS } from '$lib/telemetry'
    import { registerHotkey } from '$lib/Hotkey'

    export let repoName: string

    const {
        elements: { trigger, overlay, content },
        states: { open },
    } = createDialog()

    let searchInput: SearchInput | undefined
    let queryState = queryStateStore({ query: `repo:${repositoryInsertText({ repository: repoName })} ` }, $settings)

    registerHotkey({
        keys: { key: '/' },
        handler: () => open.set(true),
    })

    function handleSearchSubmit(state: QueryState): void {
        SVELTE_LOGGER.log(
            SVELTE_TELEMETRY_EVENTS.SearchSubmit,
            { source: 'repo', query: state.query },
            { source: 'repo', patternType: state.patternType }
        )
    }

    $: if ($open) {
        // @melt-ui automatically focuses the search input but that positions the cursor at the
        // start of the input. We can move the cursor to the end by calling focus(), but we need
        // to wait for the next tick to ensure it happens after @melt-ui has updated the DOM.
        tick().then(() => searchInput?.focus())
    }
</script>

{#if $open}
    <div class="wrapper">
        <div {...$overlay} use:overlay class="overlay" />
        <div {...$content} use:content>
            <SearchInput bind:this={searchInput} {queryState} onSubmit={handleSearchSubmit} />
        </div>
    </div>
{:else}
    <button {...$trigger} use:trigger>
        <Icon svgPath={mdiMagnify} inline aria-hidden="true" />
        Type <kbd>/</kbd> to search
    </button>
{/if}

<style lang="scss">
    .wrapper {
        flex: 1;
        position: absolute;
        left: 1rem;
        right: 1rem;
        // This seems needed to prevent the file headers (which are position: sticky) from overlaying
        // the search input. Alternatively we could portal the search input with melt, but then
        // it would be more difficult to position it over the repo header.
        z-index: 2;

        .overlay {
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background-color: rgba(0, 0, 0, 0.3);
        }
    }

    button {
        margin: 0;
        border: 1px solid var(--input-border-color);
        background-color: var(--input-bg);
        border-radius: 4px;
        padding: 0 0.5rem;
        height: 100%;
        width: 18.75rem;
        text-align: left;
        color: var(--text-muted);
        white-space: nowrap;

        &:focus {
            border-color: var(--input-focus-border-color);
        }
    }
</style>
