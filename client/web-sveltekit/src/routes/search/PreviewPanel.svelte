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
    import { getGraphQLClient, mapOrThrow, toGraphQLResult } from '$lib/graphql'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { createCodeIntelAPI } from '$lib/shared'

    // TODO: should I be importing this from a page? Seems funky.
    import { BlobPageQuery } from '../[...repo=reporev]/(validrev)/(code)/-/blob/[...path]/page.gql'

    export let repoName: string
    export let commitID: string
    export let filePath: string
    export let matchedRanges: Range[]

    const client = getGraphQLClient()
    $: codeIntelAPI = createCodeIntelAPI({
        settings: () => undefined, // TODO: add settings
        requestGraphQL(options) {
            return from(client.query(options.request, options.variables).then(toGraphQLResult))
        },
    })

    const blob = client
        .query(BlobPageQuery, {
            repoName,
            revspec: commitID,
            path: filePath,
        })
        .then(mapOrThrow(result => result.data?.repository?.commit?.blob?.content ?? null))
    // const highlights = client.query(BlobSyntaxHighlightQuery, {
    //     repoName,
    //     revspec: commitID,
    //     path: filePath,
    //     disableTimeout: false,
    // }).then(mapOrThrow(result => result.data?.repository?.commit?.blob?.highlight.lsif ?? '')),
</script>

{#await blob}
    <LoadingSpinner />
{:then content}
    <CodeMirrorBlob
        blobInfo={{
            repoName,
            commitID,
            revision: '',
            filePath,
            content: content ?? '',
            languages: ['TODO'],
        }}
        highlights=""
        {codeIntelAPI}
    />
    <!-- TODO {highlights} -->
    <!-- selectedLines={selectedPosition?.line ? selectedPosition : null} -->
    <!-- on:selectline={event => { -->
    <!--     goto('?' + updateSearchParamsWithLineInformation($page.url.searchParams, event.detail)) -->
    <!-- }} -->
    <!-- {codeIntelAPI} -->
{/await}
