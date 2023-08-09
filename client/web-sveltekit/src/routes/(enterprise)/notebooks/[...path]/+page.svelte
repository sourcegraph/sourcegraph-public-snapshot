<!--
    This page is rendered by using the corresponding React component.

    CAVEAT: Using back/forward buttons to navigate between headings inside a
    Notebook (has navigation) doesn't scroll the target into view.
    We need to investigate if that's a fundamental problem with
    SvelteKit<->ReactRouter of if Notebooks are doing something that prevents
    the default browser behavior to take place.
-->
<script lang="ts">
    import { from, type Observable } from 'rxjs'

    import { PUBLIC_DOTCOM } from '$env/static/public'
    import { eventLogger } from '$lib/logger'
    import ReactComponent from '$lib/ReactComponent.svelte'
    import { aggregateStreamingSearch, type FetchFileParameters } from '$lib/shared'
    import { GlobalNotebooksArea, type GlobalNotebooksAreaProps } from '$lib/web'

    import type { PageData } from './$types'
    import { fetchFileRangeMatches } from '$lib/search/api/highlighting'

    export let data: PageData

    const isSourcegraphDotCom = !!PUBLIC_DOTCOM

    function fetchHighlightedFileLineRanges(parameters: FetchFileParameters): Observable<string[][]> {
        return from(
            fetchFileRangeMatches({
                result: {
                    repository: parameters.repoName,
                    path: parameters.filePath,
                    commit: parameters.commitID,
                },
                ranges: parameters.ranges,
            })
        )
    }

    $: props = {
        fetchHighlightedFileLineRanges,
        telemetryService: eventLogger,
        isSourcegraphDotCom,
        // FIXME: Terrible hack to avoid having to create a complete context object
        platformContext: data.platformContext as any,
        authenticatedUser: data.user,
        notebooksEnabled: true,
        settingsCascade: data.settings,
        searchContextsEnabled: false,
        streamSearch: aggregateStreamingSearch,
        isSourcegraphApp: false,
        ownEnabled: false,
    } satisfies GlobalNotebooksAreaProps
</script>

<ReactComponent route="/notebooks/*" settings={data.settings} component={GlobalNotebooksArea} {props} />
