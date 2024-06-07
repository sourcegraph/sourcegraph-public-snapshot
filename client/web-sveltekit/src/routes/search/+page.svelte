<script lang="ts">
    // @sg EnableRollout
    import { queryStateStore } from '$lib/search/state'
    import { settings } from '$lib/stores'

    import type { PageData, Snapshot } from './$types'
    import SearchHome from './SearchHome.svelte'
    import SearchResults, { type SearchResultsCapture } from './SearchResults.svelte'

    export let data: PageData

    export const snapshot: Snapshot<{ searchResults?: SearchResultsCapture }> = {
        capture() {
            return {
                searchResults: searchResults?.capture(),
            }
        },
        restore(value) {
            if (value) {
                searchResults?.restore(value.searchResults)
            }
        },
    }

    const queryState = queryStateStore(data.queryOptions ?? {}, $settings)
    let searchResults: SearchResults | undefined
    $: queryState.set(data.queryOptions ?? {})
    $: queryState.setSettings($settings)

    const footerLinks = window.context.sourcegraphDotComMode
        ? [
              {
                  name: 'Docs',
                  href: 'https://sourcegraph.com/docs',
              },
              { name: 'About', href: 'https://sourcegraph.com' },
              {
                  name: 'Cody',
                  href: 'https://sourcegraph.com/cody',
              },
              {
                  name: 'Enterprise',
                  href: 'https://sourcegraph.com/get-started?t=enterprise',
              },
              {
                  name: 'Security',
                  href: 'https://sourcegraph.com/security',
              },
              { name: 'Discord', href: 'https://srcgr.ph/discord-server' },
          ]
        : []
</script>

<svelte:head>
    <title>{data.queryFromURL ? `${data.queryFromURL} - ` : ''}Sourcegraph</title>
</svelte:head>

{#if data.searchStream}
    <SearchResults
        bind:this={searchResults}
        stream={data.searchStream}
        queryFromURL={data.queryFromURL}
        {queryState}
        selectedFilters={data.queryFilters}
    />
{:else}
    <SearchHome {queryState} codyHref={data.codyHref} {footerLinks} />
{/if}
