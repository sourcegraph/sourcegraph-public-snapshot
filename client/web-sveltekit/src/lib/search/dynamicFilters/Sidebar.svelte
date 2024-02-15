<script lang="ts">
    import { mdiBookOpenVariant } from '@mdi/js'

    import type { Filter as QueryFilter } from '@sourcegraph/shared/src/search/query/token'

    import Icon from '$lib/Icon.svelte'
    import LanguageIcon from '$lib/LanguageIcon.svelte'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import SymbolKind from '$lib/search/SymbolKind.svelte'
    import { displayRepoName, scanSearchQuery, type Filter } from '$lib/shared'

    import { type URLQueryFilter, type SectionItem, staticTypeFilters, typeFilterIcons, groupFilters } from './index'
    import LoadingSkeleton from './LoadingSkeleton.svelte'
    import Section from './Section.svelte'

    export let searchQuery: string
    export let loading: boolean
    export let streamFilters: Filter[]
    export let selectedFilters: URLQueryFilter[]

    function queryHasTypeFilter(query: string): boolean {
        const tokens = scanSearchQuery(query)
        if (tokens.type !== 'success') {
            return false
        }
        const filters = tokens.term.filter((token): token is QueryFilter => token.type === 'filter')
        return filters.some(filter => filter.field.value === 'type')
    }

    $: groupedFilters = groupFilters(streamFilters, selectedFilters)
    $: typeFilters = staticTypeFilters.map((staticTypeFilter): SectionItem => {
        const selectedOrStreamFilter = groupedFilters.type.find(
            typeFilter => typeFilter.label === staticTypeFilter.label
        )
        return {
            ...staticTypeFilter,
            count: selectedOrStreamFilter?.count,
            exhaustive: selectedOrStreamFilter?.exhaustive || false,
            selected: selectedOrStreamFilter?.selected || false,
        }
    })
</script>

<aside class="sidebar">
    <h3>Filter results</h3>

    {#if !queryHasTypeFilter(searchQuery)}
        <Section items={typeFilters} title="By type" showAll>
            <svelte:fragment slot="label" let:label>
                <Icon svgPath={typeFilterIcons[label]} inline aria-hidden="true" />&nbsp;
                {label}
            </svelte:fragment>
        </Section>
    {/if}

    <Section items={groupedFilters['symbol type']} title="By symbol type" filterPlaceholder="Filter symbol types">
        <svelte:fragment slot="label" let:label>
            <SymbolKind symbolKind={label.toUpperCase()} />
            {label}
        </svelte:fragment>
    </Section>
    <Section items={groupedFilters.author} title="By author" filterPlaceholder="Filter authors" />
    <Section items={groupedFilters['commit date']} title="By commit date">
        <span class="commit-date-label" slot="label" let:label let:value>
            {label}
            <small><pre>{value}</pre></small>
        </span>
    </Section>
    <Section items={groupedFilters.lang} title="By language" filterPlaceholder="Filter languages">
        <svelte:fragment slot="label" let:label>
            <LanguageIcon class="icon" language={label} inline />&nbsp;
            {label}
        </svelte:fragment>
    </Section>
    <Section items={groupedFilters.repo} title="By repository" filterPlaceholder="Filter repositories">
        <svelte:fragment slot="label" let:label>
            <CodeHostIcon repository={label} />
            {displayRepoName(label)}
        </svelte:fragment>
    </Section>
    <Section items={groupedFilters.file} title="By file" showAll />
    <Section items={groupedFilters.utility} title="Utility" showAll />

    {#if loading}
        <LoadingSkeleton />
    {/if}

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
        height: 100%;
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

    a {
        margin-top: auto;
    }

    pre {
        // Overwrites global default
        margin-bottom: 0;
    }
</style>
