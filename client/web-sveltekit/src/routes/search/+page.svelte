<script lang="ts">
    import { queryStateStore } from '$lib/search/state'
    import { settings } from '$lib/stores'

    import type { PageData } from './$types'
    import SearchHome from './SearchHome.svelte'
    import SearchResults from './SearchResults.svelte'

    export let data: PageData

    const queryState = queryStateStore(data.queryOptions ?? {}, $settings)
    $: queryState.set(data.queryOptions ?? {})
    $: queryState.setSettings($settings)
</script>

{#if data.stream}
    <SearchResults
        stream={data.stream}
        queryFromURL={data.queryOptions.query}
        {queryState}
        queryFilters={data.queryFilters}
    />
{:else}
    <SearchHome {queryState} />
{/if}
