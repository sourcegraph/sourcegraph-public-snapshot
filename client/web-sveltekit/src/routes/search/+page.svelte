<script lang="ts">
    // @sg EnableRollout
    import { queryStateStore } from '$lib/search/state'
    import { settings } from '$lib/stores'
    import { TELEMETRY_V2_RECORDER } from '$lib/telemetry2'

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

    type linkNames = 'Docs' | 'About' | 'Cody' | 'Enterprise' | 'Security' | 'Discord'
    const v2LinkNameTypes: { [key in linkNames]: number } = {
        Docs: 1,
        About: 2,
        Cody: 3,
        Enterprise: 4,
        Security: 5,
        Discord: 6,
    }

    function handleLinkClick(name: linkNames): void {
        TELEMETRY_V2_RECORDER.recordEvent('home.footer.CTA', 'click', { metadata: { type: v2LinkNameTypes[name] } })
    }

    const footerLinks = window.context.sourcegraphDotComMode
        ? [
              {
                  name: 'Docs',
                  href: 'https://sourcegraph.com/docs',
                  handleClick: () => handleLinkClick('Docs'),
              },
              { name: 'About', href: 'https://sourcegraph.com', handleClick: () => handleLinkClick('About') },
              {
                  name: 'Cody',
                  href: 'https://sourcegraph.com/cody',
                  handleClick: () => handleLinkClick('Cody'),
              },
              {
                  name: 'Enterprise',
                  href: 'https://sourcegraph.com/get-started?t=enterprise',
                  handleClick: () => handleLinkClick('Enterprise'),
              },
              {
                  name: 'Security',
                  href: 'https://sourcegraph.com/security',
                  handleClick: () => handleLinkClick('Security'),
              },
              {
                  name: 'Discord',
                  href: 'https://srcgr.ph/discord-server',
                  handleClick: () => handleLinkClick('Discord'),
              },
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
