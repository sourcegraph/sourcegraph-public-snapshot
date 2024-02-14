<script lang="ts">
    import { mdiBookOpenVariant } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import Section from '$lib/search/dynamicFilters/Section.svelte'
    import TypeSection from '$lib/search/dynamicFilters/TypeSection.svelte'
    import type { QueryStateStore } from '$lib/search/state'
    import SymbolKind from '$lib/search/SymbolKind.svelte'
    import { groupFilters } from '$lib/search/utils'
    import { displayRepoName, type Filter } from '$lib/shared'

    export let queryFromURL: string
    export let streamFilters: Filter[]
    export let queryFilters: string
    export let queryState: QueryStateStore
    export let size: number

    $: width = `max(100px, min(50%, ${size * 100}%))`
    $: filters = groupFilters(streamFilters)
    $: hasFilters = filters.lang.length > 0 || filters.repo.length > 0 || filters.file.length > 0
</script>

<aside class="sidebar" style:width>
    <div class="section">
        <TypeSection {queryFromURL} {queryState} />
    </div>
    {#if hasFilters}
        <div class="section">
            {#if filters['symbol type'].length > 0}
                <Section
                    items={filters['symbol type']}
                    title="By symbol type"
                    filterPlaceholder="Filter symbol types"
                    showFilter
                    {queryFilters}
                >
                    <svelte:fragment slot="label" let:label>
                        <SymbolKind symbolKind={label.toUpperCase()} />
                        {label}
                    </svelte:fragment>
                </Section>
            {/if}
            {#if filters.author.length > 0}
                <Section
                    items={filters.author}
                    title="By author"
                    filterPlaceholder="Filter authors"
                    showFilter
                    {queryFilters}
                />
            {/if}
            {#if filters['commit date'].length > 0}
                <Section items={filters['commit date']} title="By commit date" {queryFilters}>
                    <svelte:fragment slot="label" let:label let:value>
                        <span class="commit-date-label">
                            {label}
                            <small><pre>{value}</pre></small>
                        </span>
                    </svelte:fragment>
                </Section>
            {/if}
            {#if filters.lang.length > 0}
                <Section
                    items={filters.lang}
                    title="By language"
                    showFilter
                    filterPlaceholder="Filter languages"
                    {queryFilters}
                />
            {/if}
            {#if filters.repo.length > 0}
                <Section
                    items={filters.repo}
                    title="By repository"
                    showFilter
                    filterPlaceholder="Filter repositories"
                    preprocessLabel={displayRepoName}
                    {queryFilters}
                >
                    <svelte:fragment slot="label" let:label>
                        <CodeHostIcon repository={label} />
                        {displayRepoName(label)}
                    </svelte:fragment>
                </Section>
            {/if}
            {#if filters.file.length > 0}
                <Section items={filters.file} title="By file" {queryFilters} />
            {/if}
            {#if filters.utility.length > 0}
                <Section items={filters.utility} title="Utility" {queryFilters} />
            {/if}
        </div>
    {/if}
    <a class="section help" href="/help/code_search/reference/queries" target="_blank">
        <span class="icon">
            <Icon svgPath={mdiBookOpenVariant} inline />
        </span>
        <div>
            <h4>Need more advanced filters?</h4>
            <span>Explore the query syntax docs</span>
        </div>
    </a>
</aside>

<style lang="scss">
    .sidebar {
        flex: 0 0 auto;
        background-color: var(--sidebar-bg);
        overflow-y: auto;
        display: flex;
        flex-direction: column;

        h4 {
            font-weight: 600;
            white-space: nowrap;
            margin-bottom: 1rem;
        }
    }

    .section {
        padding: 1rem 0.5rem 1rem 1rem;
        border-top: 1px solid var(--border-color);

        &:first-child {
            border-top: none;
        }

        &:last-child {
            margin-top: auto;
        }
    }

    .help {
        display: flex;
        align-items: center;
        gap: 0.5rem;

        text-decoration: none;
        color: var(--text-muted);
        font-size: 0.75rem;

        h4 {
            margin: 0;
        }

        .icon {
            flex-shrink: 0;
        }
    }

    pre {
        // Overwrites global default
        margin-bottom: 0;
    }
</style>
