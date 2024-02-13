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
    $: filters = groupFilters(streamFilters, selectedFilters)
    $: console.log(filters)
    $: typeFilters = staticTypeFilters.map(staticTypeFilter => {
        const selectedOrStreamFilter = filters.type.find(typeFilter => typeFilter.label === staticTypeFilter.label)
        console.log({ staticTypeFilter, selectedOrStreamFilter })
        return {
            ...staticTypeFilter,
            count: selectedOrStreamFilter?.count,
            exhaustive: selectedOrStreamFilter?.exhaustive || false,
            selected: selectedOrStreamFilter?.selected || false,
        }
    })
    $: console.log(typeFilters)
</script>

<aside class="sidebar" style:width>
    <h3>Filter results</h3>
    <Section items={typeFilters} title="By type" showAll>
        <svelte:fragment slot="label" let:label>
            <Icon svgPath={typeFilterIcons.get(label) ?? ''} inline aria-hidden="true" />
            {label}
        </svelte:fragment>
    </Section>
    <Section items={filters['symbol type']} title="By symbol type" filterPlaceholder="Filter symbol types">
        <svelte:fragment slot="label" let:label>
            <SymbolKind symbolKind={label.toUpperCase()} />
            {label}
        </svelte:fragment>
    </Section>
    <Section items={filters.author} title="By author" filterPlaceholder="Filter authors" />
    <Section items={filters['commit date']} title="By commit date">
        <svelte:fragment slot="label" let:label let:value>
            <span class="commit-date-label">
                {label}
                <small><pre>{value}</pre></small>
            </span>
        </svelte:fragment>
    </Section>
    <Section items={filters.lang} title="By language" filterPlaceholder="Filter languages" />
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
    <Section items={filters.file} title="By file" showAll />
    <Section items={filters.utility} title="Utility" showAll />
    <a class="section help" href="/help/code_search/reference/queries" target="_blank">
        <span class="icon">
            <Icon --color="var(--icon-color)" svgPath={mdiBookOpenVariant} inline />
        </span>
        <div>
            <h4>Need more advanced filters?</h4>
            <span>Explore the query syntax docs</span>
        </div>
    </a>
</aside>

<style lang="scss">
    .sidebar {
        h3 {
            margin: 0;
        }
        padding: 1rem;
        background-color: var(--sidebar-bg);
        overflow-y: auto;

        display: flex;
        flex-direction: column;
        gap: 1.5rem;
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
