<script lang="ts">
    // @sg EnableRollout

    import { GitRefType } from '$lib/graphql-types'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import GitReferencesTable from '$lib/repo/GitReferencesTable.svelte'
    import { Alert } from '$lib/wildcard'

    import type { PageData } from './$types'

    export let data: PageData
</script>

<svelte:head>
    <title>Branches - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<div class="scroller">
    <div class="content">
        {#await data.overview}
            <LoadingSpinner />
        {:then result}
            {@const activeBranches = result.branches.nodes.filter(branch => branch.id !== result.defaultBranch?.id)}

            <h2>Default branch</h2>
            {#if result.defaultBranch}
                <GitReferencesTable
                    references={[result.defaultBranch]}
                    referenceType={GitRefType.GIT_BRANCH}
                    defaultBranch={result.defaultBranch.displayName}
                />
            {/if}

            <h2>Active branches</h2>

            {#if activeBranches.length > 0}
                <GitReferencesTable references={activeBranches} referenceType={GitRefType.GIT_BRANCH} />
            {/if}
        {:catch error}
            <Alert variant="danger">
                Unable to fetch branches:
                {error.message}
            </Alert>
        {/await}
    </div>
</div>

<style lang="scss">
    .scroller {
        overflow: auto;
    }

    .content {
        max-width: var(--viewport-xl);
        width: 100%;
        margin: 0 auto;

        padding: 0 1rem;

        @media (--mobile) {
            padding: 0;
        }
    }

    h2 {
        font-weight: 500;
        margin: 1rem 0;
        font-size: var(--font-size-base);

        @media (--mobile) {
            margin-left: 0.5rem;
            margin-right: 0.5rem;
        }
    }
</style>
