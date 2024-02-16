<script lang="ts">
    import { mdiClose } from '@mdi/js'

    import { page } from '$app/stores'
    import { pluralize } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Badge, Button } from '$lib/wildcard'

    import { updateFilterInURL, type SectionItem } from './index'

    export let items: SectionItem[]
    export let title: string
    export let filterPlaceholder: string = ''
    export let showAll: boolean = false

    let filterText = ''
    $: processedFilterText = filterText.trim().toLowerCase()
    $: filteredItems = processedFilterText
        ? items.filter(item => item.label.toLowerCase().includes(processedFilterText))
        : items
    let filterInputRef: HTMLInputElement

    let showMore = false
    $: showCount = showAll ? items.length : showMore ? 10 : 5
    $: limitedItems = filteredItems.slice(0, showCount)

    function roundCount(count: number): number {
        const roundNumbers = [10000, 5000, 1000, 500, 100, 50, 10, 5, 1]
        for (const roundNumber of roundNumbers) {
            if (count >= roundNumber) {
                return roundNumber
            }
        }
        return 0
    }
</script>

{#if items.length > 0}
    <article>
        <header><h4>{title}</h4></header>
        {#if items.length > showCount}
            <input bind:this={filterInputRef} bind:value={filterText} placeholder={filterPlaceholder} />
        {/if}
        <ul>
            {#each limitedItems as item}
                <li>
                    <a
                        href={updateFilterInURL($page.url, item, item.selected).toString()}
                        class:selected={item.selected}
                    >
                        <span class="label">
                            <slot name="label" label={item.label} value={item.value}>
                                {item.label}
                            </slot>
                        </span>
                        {#if item.count !== undefined}
                            <span class="count">
                                {#if item.exhaustive}
                                    <Badge variant="secondary">{item.count}</Badge>
                                {:else}
                                    <Tooltip
                                        tooltip="At least {item.count} {pluralize(
                                            'result',
                                            item.count
                                        )} match this filter."
                                    >
                                        <Badge variant="secondary">{roundCount(item.count)}+</Badge>
                                    </Tooltip>
                                {/if}
                            </span>
                        {/if}
                        {#if item.selected}
                            <span class="close">
                                <Icon svgPath={mdiClose} inline />
                            </span>
                        {/if}
                    </a>
                </li>
            {/each}
        </ul>
        {#if showMore}
            {#if filteredItems.length > limitedItems.length}
                <small class="not-shown-message">
                    {filteredItems.length - limitedItems.length} not shown. Use
                    <Button variant="link" display="inline" on:click={() => filterInputRef.focus()}>search</Button>
                    to see more.
                </small>
            {/if}
            <footer class="show-more">
                <Button variant="link" on:click={() => (showMore = false)}>Show less</Button>
            </footer>
        {:else if !showMore && filteredItems.length > limitedItems.length}
            <footer class="show-more">
                <Button variant="link" on:click={() => (showMore = true)}>Show more</Button>
            </footer>
        {/if}
    </article>
{/if}

<style lang="scss">
    article {
        padding: 0 1rem;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }

    input {
        display: block;
        width: 100%;
        height: var(--input-height);
        padding: var(--input-padding-y) var(--input-padding-x);
        font-size: var(--input-font-size);
        font-weight: var(--input-font-weight);
        line-height: var(--input-line-height);
        color: var(--input-color);
        background-color: var(--input-bg);
        background-clip: padding-box;
        border: var(--input-border-width) solid var(--input-border-color);
        border-radius: var(--border-radius);
    }

    ul {
        margin: 0;
        padding: 0;
        list-style: none;
    }

    .not-shown-message {
        text-align: center;
        background-color: var(--secondary-2);
        color: var(--text-muted);
        padding: 0.75rem 1rem;
        border-radius: 0.5rem;
        :global(button) {
            padding: 0;
            text-align: inherit;
            line-height: inherit;
            font-size: inherit;
            font-style: inherit;
            vertical-align: inherit;
        }
    }

    .show-more {
        text-align: center;
    }

    li a {
        display: flex;
        width: 100%;
        align-items: center;
        border: none;
        text-align: left;
        text-decoration: none;
        border-radius: var(--border-radius);
        color: inherit;
        white-space: nowrap;
        gap: 0.25rem;

        padding: 0.25rem 0.25rem 0.25rem 0.5rem;
        margin: 0;
        font-weight: 400;

        .label {
            flex: 1;
            text-overflow: ellipsis;
            overflow: hidden;
        }

        &:hover {
            background-color: var(--secondary-4);
        }

        &.selected {
            background-color: var(--primary);
            color: var(--primary-4);
            --color: var(--primary-4);
        }

        .close {
            flex-shrink: 0;
        }
    }
</style>
