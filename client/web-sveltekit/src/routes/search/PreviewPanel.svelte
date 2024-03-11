<script lang="ts" context="module">
    export interface Range {
        start: Location
        end: Location
    }

    export interface Location {
        // A zero-based line number
        line: number
        // A zero-based column number
        column: number
    }
</script>

<script lang="ts">
    import { from } from 'rxjs'

    import CodeMirrorBlob from '$lib/CodeMirrorBlob.svelte'
    import { isErrorLike } from '$lib/common'
    import { getGraphQLClient, mapOrThrow, toGraphQLResult } from '$lib/graphql'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { createCodeIntelAPI } from '$lib/shared'
    import { settings } from '$lib/stores'

    // TODO: should I be importing this from a page? Seems funky.
    import {
        BlobPageQuery,
        BlobSyntaxHighlightQuery,
    } from '../[...repo=reporev]/(validrev)/(code)/-/blob/[...path]/page.gql'

    export let repoName: string
    export let commitID: string
    export let filePath: string
    export let matchedRanges: Range[]

    const client = getGraphQLClient()
    $: codeIntelAPI = createCodeIntelAPI({
        settings: setting => (isErrorLike($settings?.final) ? undefined : $settings?.final?.[setting]),
        requestGraphQL(options) {
            return from(client.query(options.request, options.variables).then(toGraphQLResult))
        },
    })

    console.log({ filePath })

    const blob = client
        .query(BlobPageQuery, {
            repoName,
            revspec: commitID,
            path: filePath,
        })
        .then(mapOrThrow(result => result.data?.repository?.commit?.blob ?? null))

    const highlights = client
        .query(BlobSyntaxHighlightQuery, {
            repoName,
            revspec: commitID,
            path: filePath,
            disableTimeout: false,
        })
        .then(mapOrThrow(result => result.data?.repository?.commit?.blob?.highlight.lsif ?? ''))
</script>

<div class="content">
    {#await blob}
        <LoadingSpinner />
    {:then blob}
        {#await highlights}
            <CodeMirrorBlob
                blobInfo={{
                    repoName,
                    commitID,
                    revision: '',
                    filePath,
                    content: blob?.content ?? '',
                    languages: blob?.languages ?? [],
                }}
                highlights={''}
                {codeIntelAPI}
            />
        {:then highlights}
            <CodeMirrorBlob
                blobInfo={{
                    repoName,
                    commitID,
                    revision: '',
                    filePath,
                    content: blob?.content ?? '',
                    languages: blob?.languages ?? [],
                }}
                {highlights}
                {codeIntelAPI}
            />
        {/await}
    {:catch error}
        <p>{error}</p>
    {/await}
</div>

<style lang="scss">
    .content {
        display: flex;
        flex-direction: column;
        overflow-x: auto;
        height: 100%;
        flex: 1;
    }
</style>
