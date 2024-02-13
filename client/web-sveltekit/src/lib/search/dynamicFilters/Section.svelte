<script lang="ts">
    import { mdiClose } from '@mdi/js'

    import { page } from '$app/stores'
    import { pluralize } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Badge, Button } from '$lib/wildcard'

    import { updateFilterInURL, type SidebarFilter } from './index'

    export let items: SidebarFilter[]
    export let title: string
    export let filterPlaceholder: string = ''
    export let showAll: boolean = false
    export let preprocessLabel: (label: string) => string = label => label

    function generateURL(filter: SidebarFilter) {
        return updateFilterInURL($page.url, filter, filter.selected)
    }

    let filterText = ''
    $: processedFilterText = filterText.trim().toLowerCase()
    $: filteredItems = processedFilterText
        ? items.filter(item => preprocessLabel(item.label).toLowerCase().includes(processedFilterText))
        : items
    $: limitedItems = showAll ? filteredItems : filteredItems.slice(0, 5)
</script>

{#if items.length > 0}
    <article>
        <header><h4>{title}</h4></header>
        {#if !showAll && items.length > 5}
            <input bind:value={filterText} placeholder={filterPlaceholder} />
        {/if}
        <ul>
            {#each limitedItems as item}
                <li>
                    <a href={generateURL(item).toString()} class:selected={item.selected}>
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
                                        <Badge variant="secondary">{item.count}+</Badge>
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
        {#if limitedItems.length < filteredItems.length}
            <footer>
                <Button variant="link" on:click={() => (showAll = true)}>
                    Show all ({filteredItems.length})
                </Button>
            </footer>
        {:else if !showAll && limitedItems.length > 5}
            <footer>
                <Button variant="link" on:click={() => (showAll = false)}>Show less</Button>
            </footer>
        {/if}
    </article>
{/if}

<style lang="scss">
    h4 {
        white-space: nowrap;
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
        margin-bottom: 0.5rem;
    }

    ul {
        margin: 0;
        padding: 0;
        list-style: none;
    }

    footer {
        text-align: center;
    }

    a {
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

        padding: 0.25rem 0.5rem;
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
            margin-left: 0.25rem;
            flex-shrink: 0;
        }
    }
</style>
