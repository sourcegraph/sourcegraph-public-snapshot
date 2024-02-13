<script lang="ts">
    import { mdiBookOpenVariant } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import Section from '$lib/search/dynamicFilters/Section.svelte'
    import SymbolKind from '$lib/search/SymbolKind.svelte'
    import { displayRepoName, type Filter } from '$lib/shared'

    import { type URLQueryFilter, staticTypeFilters, typeFilterIcons, groupFilters } from './index'

    export let streamFilters: Filter[]
    export let selectedFilters: URLQueryFilter[]
    export let size: number

    $: width = `max(100px, min(50%, ${size * 100}%))`
    // TODO: merge streamed filters and static type filters
    $: filters = groupFilters(streamFilters, selectedFilters)
</script>

<aside class="sidebar" style:width>
    <h3>Filter results</h3>
    <Section items={filters['type']} title="By type" showAll>
        <svelte:fragment slot="label" let:label>
            <Icon svgPath={typeFilterIcons[label]} inline aria-hidden="true" />
            {label}
        </svelte:fragment>
    </Section>
    {#if filters['symbol type'].length > 0}
        <Section items={filters['symbol type']} title="By symbol type" filterPlaceholder="Filter symbol types">
            <svelte:fragment slot="label" let:label>
                <SymbolKind symbolKind={label.toUpperCase()} />
                {label}
            </svelte:fragment>
        </Section>
    {/if}
    {#if filters.author.length > 0}
        <Section items={filters.author} title="By author" filterPlaceholder="Filter authors" />
    {/if}
    {#if filters['commit date'].length > 0}
        <Section items={filters['commit date']} title="By commit date">
            <svelte:fragment slot="label" let:label let:value>
                <span class="commit-date-label">
                    {label}
                    <small><pre>{value}</pre></small>
                </span>
            </svelte:fragment>
        </Section>
    {/if}
    {#if filters.lang.length > 0}
        <Section items={filters.lang} title="By language" filterPlaceholder="Filter languages" />
    {/if}
    {#if filters.repo.length > 0}
        <Section
            items={filters.repo}
            title="By repository"
            filterPlaceholder="Filter repositories"
            preprocessLabel={displayRepoName}
        >
            <svelte:fragment slot="label" let:label>
                <CodeHostIcon repository={label} />
                {displayRepoName(label)}
            </svelte:fragment>
        </Section>
    {/if}
    {#if filters.utility.length > 0}
        <Section items={filters.utility} title="Utility" showAll />
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
