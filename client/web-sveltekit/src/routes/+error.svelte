<script lang="ts">
    import { mdiAlertCircle } from '@mdi/js'

    import { page } from '$app/stores'
    import HeroPage from '$lib/HeroPage.svelte'
    import { isCloneInProgressErrorLike, isRepoNotFoundErrorLike, isRevisionNotFoundErrorLike } from '$lib/shared'

    import CloneInProgressError from './[...repo=reporev]/CloneInProgressError.svelte'
    import RepoNotFoundError from './[...repo=reporev]/RepoNotFoundError.svelte'
    import RevisionNotFoundError from './[...repo=reporev]/RevisionNotFoundError.svelte'
</script>

{#if isRepoNotFoundErrorLike($page.error)}
    <RepoNotFoundError repoName={$page.params.repo} viewerCanAdminister={$page.data.user?.siteAdmin ?? false} />
{:else if isCloneInProgressErrorLike($page.error)}
    <CloneInProgressError repoName={$page.params.repo} error={$page.error} />
{:else if isRevisionNotFoundErrorLike($page.error)}
    <RevisionNotFoundError />
{:else}
    <HeroPage title="Error" svgIconPath={mdiAlertCircle}>
        <!-- TODO: format error message with markdown -->
        {$page.error?.message ?? 'Unknown Error'}
    </HeroPage>
{/if}
