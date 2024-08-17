<script lang="ts">
    import { goto } from '$app/navigation'
    import { page } from '$app/stores'
    import HeroPage from '$lib/HeroPage.svelte'
    import { isCloneInProgressErrorLike, isRepoNotFoundErrorLike, isRevisionNotFoundErrorLike } from '$lib/shared'

    import { SgRedirect } from '../hooks.client'

    import CloneInProgressError from './[...repo=reporev]/CloneInProgressError.svelte'
    import RepoNotFoundError from './[...repo=reporev]/RepoNotFoundError.svelte'
    import RevisionNotFoundError from './[...repo=reporev]/RevisionNotFoundError.svelte'

    let showError = true

    // This is a nasty workaround to handle a production bug where redirects are not recognized by SvelteKit.
    // See the comment in hooks.client.ts ans SRCH-926 for more information
    $: if ($page.error instanceof SgRedirect) {
        showError = false
        goto($page.error.redirect.location)
    }
</script>

{#if showError}
    {#if isRepoNotFoundErrorLike($page.error)}
        <RepoNotFoundError repoName={$page.params.repo} viewerCanAdminister={$page.data.user?.siteAdmin ?? false} />
    {:else if isCloneInProgressErrorLike($page.error)}
        <CloneInProgressError repoName={$page.params.repo} error={$page.error} />
    {:else if isRevisionNotFoundErrorLike($page.error)}
        <RevisionNotFoundError />
    {:else}
        <HeroPage title="Unexpected Error" icon={ILucideCircleX}>
            <!-- TODO: format error message with markdown -->
            {$page.error?.message ?? '(no error message)'}
        </HeroPage>
    {/if}
{/if}
