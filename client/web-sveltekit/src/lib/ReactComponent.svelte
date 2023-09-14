<script lang="ts">
    import React from 'react'

    import { createRoot, type Root } from 'react-dom/client'
    import { onDestroy, onMount } from 'svelte'

    import type { SettingsCascadeOrError } from '$lib/shared'

    import { ReactAdapter } from './react-interop'

    type ComponentProps = $$Generic<{}>

    export let component: React.FunctionComponent<ComponentProps>
    export let props: ComponentProps
    export let route: string
    export let settings: SettingsCascadeOrError

    let container: HTMLDivElement
    let root: Root | null = null

    function renderComponent(
        root: Root | null,
        component: React.FunctionComponent<ComponentProps>,
        props: ComponentProps,
        route: string,
        settings: SettingsCascadeOrError
    ) {
        root?.render(
            React.createElement(
                ReactAdapter,
                {
                    route,
                    settings,
                },
                React.createElement(component, props)
            )
        )
    }

    onMount(() => (root = createRoot(container)))
    onDestroy(() => root?.unmount())
    $: renderComponent(root, component, props, route, settings)
</script>

<div bind:this={container} />
