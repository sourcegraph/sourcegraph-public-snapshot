<script lang="ts">
    import { PUBLIC_DOTCOM } from '$env/static/public'
    import {GlobalNotebooksArea, type GlobalNotebooksAreaProps} from '@sourcegraph/web/src/notebooks/GlobalNotebooksArea'
    import type { PageData } from './$types'
    import {fetchHighlightedFileLineRanges as _fetchHighlightedFileLineRanges, type FetchFileParameters} from '$lib/shared'
    import type { Observable } from 'rxjs'
    import ReactComponent from '$lib/ReactComponent.svelte'
    import { isLightTheme } from '$lib/stores'

    export let data: PageData

    const telemetryService = {
        log: () => undefined,
        logViewEvent: () => undefined,
        logPageView: () => undefined,
    }
    const isSourcegraphDotCom = !!PUBLIC_DOTCOM

    function fetchHighlightedFileLineRanges(parameters: FetchFileParameters, force?: boolean): Observable<string[][]> {
        return _fetchHighlightedFileLineRanges({...parameters, platformContext: data.platformContext}, force)
    }

    let props: GlobalNotebooksAreaProps
    $: props = {
        fetchHighlightedFileLineRanges,
        telemetryService,
        isSourcegraphDotCom,
        platformContext: data.platformContext,
        authenticatedUser: data.user,
        notebooksEnabled: true,
        globbing: false,
        isLightTheme: $isLightTheme,
    }
</script>

<ReactComponent route="/notebooks/*" component={GlobalNotebooksArea} {props}/>
