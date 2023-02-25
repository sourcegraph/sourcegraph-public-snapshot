<script lang="ts">
    import { CodeInsightsRouter, type CodeInsightsRouterProps } from '$lib/web'
    import { PUBLIC_DOTCOM } from '$env/static/public'
    import type { PageData } from './$types'
    import ReactComponent from '$lib/ReactComponent.svelte'

    export let data: PageData

    // TODO: Hook up to telemetry service
    const telemetryService = {
        log: () => undefined,
        logViewEvent: () => undefined,
        logPageView: () => undefined,
    }
    const isSourcegraphDotCom = !!PUBLIC_DOTCOM

    $: props = {
        telemetryService,
        isSourcegraphDotCom,
        authenticatedUser: data.user,
    } satisfies CodeInsightsRouterProps
</script>

<ReactComponent route="insights/*" component={CodeInsightsRouter} {props} />
