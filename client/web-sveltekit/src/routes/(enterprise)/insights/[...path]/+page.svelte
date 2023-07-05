<script lang="ts">
    import { PUBLIC_DOTCOM } from '$env/static/public'
    import { eventLogger } from '$lib/logger'
    import ReactComponent from '$lib/ReactComponent.svelte'
    import { CodeInsightsRouter, type CodeInsightsRouterProps } from '$lib/web'

    import type { PageData } from './$types'

    export let data: PageData

    const isSourcegraphDotCom = !!PUBLIC_DOTCOM

    $: props = {
        telemetryService: eventLogger,
        isSourcegraphDotCom,
        authenticatedUser: data.user,
    } satisfies CodeInsightsRouterProps
</script>

<ReactComponent route="/insights/*" settings={data.settings} component={CodeInsightsRouter} {props} />
