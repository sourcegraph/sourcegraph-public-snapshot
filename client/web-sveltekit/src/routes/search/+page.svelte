<script lang="ts">
    import QueryExamples from './QueryExamples.svelte'
    import { getQueryExamples } from '$lib/search/queryExamples'

    import type { PageData } from './$types'
    import SearchHome from './SearchHome.svelte'
    import SearchResults from './SearchResults.svelte'
    import { queryStateStore } from '$lib/search/state'
    import { settings } from '$lib/stores'

    export let data: PageData

    $: queryState = queryStateStore(data.queryOptions ?? {}, $settings)
    $: queryState.setSettings($settings)
</script>

{#if data.stream}
    <SearchResults stream={data.stream} queryFromURL={data.queryOptions.query} {queryState} />
{:else}
    <SearchHome {queryState}>
        <div class="mt-5">
            <!--
                Example for how we might want to make the homepage composable.
                Ideally all logic that determines what to shoe for a specific
                version (e.g. which examples to show) is kept inside pages.
            -->
            <QueryExamples examples={getQueryExamples()} />
        </div>
    </SearchHome>
{/if}
