<!--
    This page is rendered by using the corresponding React component.

    CAVEAT: Using back/forward buttons to navigate between headings inside a
    Notebook (has navigation) doesn't scroll the target into view.
    We need to investigate if that's a fundamental problem with
    SvelteKit<->ReactRouter of if Notebooks are doing something that prevents
    the default browser behavior to take place.
-->
<script lang="ts">
    import { PUBLIC_DOTCOM } from '$env/static/public'
    import { GlobalNotebooksArea, type GlobalNotebooksAreaProps } from '$lib/web'
    import type { PageData } from './$types'
    import {
        aggregateStreamingSearch,
        fetchHighlightedFileLineRanges as _fetchHighlightedFileLineRanges,
        type FetchFileParameters,
    } from '$lib/shared'
    import type { Observable } from 'rxjs'
    import ReactComponent from '$lib/ReactComponent.svelte'
    import { eventLogger } from '$lib/logger'

    export let data: PageData

    const isSourcegraphDotCom = !!PUBLIC_DOTCOM

    function fetchHighlightedFileLineRanges(parameters: FetchFileParameters, force?: boolean): Observable<string[][]> {
        return _fetchHighlightedFileLineRanges({ ...parameters, platformContext: data.platformContext }, force)
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
    } satisfies GlobalNotebooksAreaProps
</script>

<ReactComponent route="/notebooks/*" component={GlobalNotebooksArea} {props} />
