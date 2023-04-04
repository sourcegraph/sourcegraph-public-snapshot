<script lang="ts">
    import { CodeInsightsRouter, type CodeInsightsRouterProps } from '$lib/web'
    import { PUBLIC_DOTCOM } from '$env/static/public'
    import type { PageData } from './$types'
    import ReactComponent from '$lib/ReactComponent.svelte'
    import { eventLogger } from '$lib/logger'

    export let data: PageData

    const isSourcegraphDotCom = !!PUBLIC_DOTCOM

    $: props = {
        telemetryService: eventLogger,
        isSourcegraphDotCom,
        authenticatedUser: data.user,
    } satisfies CodeInsightsRouterProps
</script>

<ReactComponent route="/insights/*" component={CodeInsightsRouter} {props} />
