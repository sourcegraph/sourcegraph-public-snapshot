<script lang="ts">
    import { mdiMapSearch } from '@mdi/js'
    import { map, catchError } from 'rxjs/operators'

    import { logViewEvent } from '$lib/logger'
    import { asError, isErrorLike, type ErrorLike } from '$lib/common'
    import { checkMirrorRepositoryConnection } from '$lib/web'
    import HeroPage from '$lib/HeroPage.svelte'

    export let repoName: string
    export let viewerCanAdminister: boolean

    logViewEvent('RepositoryError')

    function check(repoName: string): Promise<ErrorLike | boolean> {
        return checkMirrorRepositoryConnection({ name: repoName })
            .pipe(
                map(result => result.error === null),
                catchError(error => [asError(error)])
            )
            .toPromise()
    }
</script>

<HeroPage title="Repository not found" svgIconPath={mdiMapSearch}>
    {#if viewerCanAdminister}
        <div>
            {#await check(repoName)}
                Checking whether this repository can be added...
            {:then canAddOrError}
                {#if canAddOrError === true}
                    <p>
                        As a site admin, you can add this repository to Sourcegraph to allow users to search and view it
                        by <a href="/site-admin/external-services">connecting an external service</a>
                    </p>
                {:else if canAddOrError === false || isErrorLike(canAddOrError)}
                    <p>
                        If this is a private repository, check that this site is configured with a token that has access
                        to this repository.
                    </p>
                    <p>
                        If this is a public repository, check that this repository is explicitly listed in an
                        <a href="/site-admin/external-services">external service configuration</a>.
                    </p>
                {/if}
            {/await}
        </div>
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
