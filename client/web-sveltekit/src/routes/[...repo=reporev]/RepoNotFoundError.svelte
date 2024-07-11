<script lang="ts">
    import { onMount } from 'svelte'

    import HeroPage from '$lib/HeroPage.svelte'
    import { TELEMETRY_RECORDER } from '$lib/telemetry'

    export let repoName: string
    export let viewerCanAdminister: boolean

    onMount(() => {
        TELEMETRY_RECORDER.recordEvent('repo.error.notFound', 'view')
    })
</script>

<HeroPage title="Repository not found" icon={ILucideBookX}>
    {#if viewerCanAdminister}
        <p>
            As a site admin, you can add <code>{repoName}</code> to Sourcegraph to allow users to search and view it by
            <a href="/site-admin/external-services">connecting an external service</a> referencing it.
        </p>
    {:else}
        <p>To access this repository, contact the Sourcegraph admin.</p>
    {/if}
</HeroPage>

<style lang="scss">
    p {
        text-align: center;
        max-width: var(--viewport-md);
    }
</style>
