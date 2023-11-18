<script lang="ts">
    import { mdiMapSearch } from '@mdi/js'

    import { asError, isErrorLike, type ErrorLike } from '$lib/common'
    import HeroPage from '$lib/HeroPage.svelte'
    import { logViewEvent } from '$lib/logger'

    export let repoName: string
    export let viewerCanAdminister: boolean

    logViewEvent('RepositoryError')
</script>

<HeroPage title="Repository not found" svgIconPath={mdiMapSearch}>
    {#if viewerCanAdminister}
        <p>
            As a site admin, you can add <Code>{repo}</Code> to Sourcegraph to allow users to search and
            view it by <a href="/site-admin/external-services">connecting an external service</a> referencing it.
        </p>
    {:else}
        <p>To access this repository, contact the Sourcegraph admin.</p>
    {/if}
</HeroPage>

<style lang="scss">
    div {
        padding: 2rem;
        width: 100vw;
        max-width: 36rem;
    }
</style>
