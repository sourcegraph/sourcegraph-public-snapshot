<script lang="ts">
    import { page } from '$app/stores'
    import Icon from '$lib/Icon.svelte'
    import type { SidebarFilter } from '$lib/search/utils'
    import { updateFilter } from '$lib/shared'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Badge } from '$lib/wildcard'
    import { mdiClose } from '@mdi/js'

    export let items: SidebarFilter[]
    export let title: string
    export let queryFilters: string

    function generateURL(filter: SidebarFilter, remove: boolean) {
        const url = new URL($page.url)
        let filters = queryFilters
        if (remove) {
            filters = filters.replace(filter.value, '').trim()
        } else {
            try {
                const separator = filter.value.indexOf(':')
                const key = filter.value.slice(0, separator)
                const value = filter.value.slice(separator + 1)
                filters = updateFilter(queryFilters, key, value)
            } catch {
                filters = filter.value
            }
        }
        if (filters) {
            url.searchParams.set('filters', filters)
        } else {
            url.searchParams.delete('filters')
        }
        return url.toString()
    }
</script>

<article>
    <header><h4>{title}</h4></header>
    <ul>
        {#each items as item}
            {@const selected = queryFilters.includes(item.value)}
            <li>
                <a href={generateURL(item, selected)} class:selected>
                    <span class="label">
                        <slot name="label" label={item.label}>
                            {item.label}
                        </slot>
                    </span>
                    {#if item.count !== undefined}
                        <span class="count">
                            <Tooltip tooltip="At least {item.count} results match this filter.">
                                <Badge variant="secondary">{item.count}</Badge>
                            </Tooltip>
                        </span>
                    {/if}
                    {#if selected}
                        <span class="close">
                            <Icon svgPath={mdiClose} inline />
                        </span>
                    {/if}
                </a>
            </li>
        {/each}
    </ul>
</article>

<style lang="scss">
    h4 {
        white-space: nowrap;
    }

    ul {
        margin: 0;
        padding: 0.125rem;
        padding-bottom: 1rem;
        list-style: none;
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
