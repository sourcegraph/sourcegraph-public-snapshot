<script lang="ts">
    import { setContext, onMount } from 'svelte'

    import { logoLight, logoDark } from '$lib/images'
    import SearchInput from '$lib/search/input/SearchInput.svelte'
    import type { QueryStateStore } from '$lib/search/state'
    import type { SearchPageContext } from '$lib/search/utils'
    import { TELEMETRY_SEARCH_SOURCE_TYPE } from '$lib/shared'
    import { isLightTheme } from '$lib/stores'
    import { TELEMETRY_RECORDER } from '$lib/telemetry'

    import CodyUpsellBanner from './cody-upsell/CodyUpsellBanner.svelte'
    import DotcomFooterLinks from './DotcomFooterLinks.svelte'
    import SearchHomeNotifications from './SearchHomeNotifications.svelte'

    export let queryState: QueryStateStore
    export let codyHref: string = '/cody'
    export let showDotcomFooterLinks: boolean = false

    setContext<SearchPageContext>('search-context', {
        setQuery(newQuery) {
            queryState.setQuery(newQuery)
        },
    })

    onMount(() => {
        TELEMETRY_RECORDER.recordEvent('home', 'view')
    })

    function handleSubmit() {
        TELEMETRY_RECORDER.recordEvent('search', 'submit', {
            metadata: { source: TELEMETRY_SEARCH_SOURCE_TYPE['home'] },
        })
    }
</script>

<section>
    <div class="content">
        <img class="logo" src={$isLightTheme ? logoLight : logoDark} alt="Sourcegraph Logo" />
        <div class="search">
            <SearchInput {queryState} autoFocus onSubmit={handleSubmit} />
            <SearchHomeNotifications />
        </div>
        <CodyUpsellBanner {codyHref} />
        {#if showDotcomFooterLinks}
            <DotcomFooterLinks />
        {/if}
    </div>
</section>

<style lang="scss">
    section {
        overflow-y: auto;
        padding: 3rem 1rem;
        display: flex;
        flex-direction: column;
        flex: 1;
        align-items: center;
    }

    div.content {
        padding-top: 3rem;
        flex-shrink: 0;
        display: flex;
        gap: 3rem;
        flex-direction: column;
        align-items: center;
        width: 100%;
        max-width: 64rem;

        :global(.search-box) {
            align-self: stretch;
        }
    }

    .search {
        width: 100%;
        display: flex;
        flex-direction: column;
        gap: 2rem;
        z-index: 1;
    }

    img.logo {
        width: 20rem;
        max-width: 90%;
        min-height: 54px;
    }
</style>
