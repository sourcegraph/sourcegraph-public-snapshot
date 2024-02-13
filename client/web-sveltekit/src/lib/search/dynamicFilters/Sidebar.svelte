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
    <h3>Filter results</h3>
    <div>
        <!-- TODO: unify the type section with Section -->
        <h4>By type</h4>
        <TypeSection {queryFromURL} {queryState} />
    </div>
    {#if hasFilters}
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
        padding: 1rem;
        background-color: var(--sidebar-bg);
        overflow-y: auto;

        display: flex;
        flex-direction: column;
        gap: 1rem;
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
