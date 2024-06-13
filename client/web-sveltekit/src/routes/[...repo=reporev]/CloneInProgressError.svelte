<script lang="ts">
    import { onMount } from 'svelte'

    import { invalidate } from '$app/navigation'
    import HeroPage from '$lib/HeroPage.svelte'
    import { displayRepoName, type CloneInProgressError } from '$lib/shared'

    // TODO: Find a way to make this type stricter
    export let error: App.Error | null
    export let repoName: string

    $: progress = (error as CloneInProgressError).progress

    onMount(() => {
        // Invalidate the repo root every second to trigger an update
        // and get new progress information.
        const id = setInterval(() => invalidate('repo:root'), 1000)
        return () => clearInterval(id)
    })
</script>

<HeroPage title={displayRepoName(repoName)} icon={ILucideFolderGit2}>
    <code><pre>{progress}</pre></code>
    <!--TODO add DirectImportRepoAlert -->
</HeroPage>

<style lang="scss">
</style>
