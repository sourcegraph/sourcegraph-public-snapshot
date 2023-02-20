<script lang="ts">
    import React from "react"
    import { createRoot, type Root } from "react-dom/client"
    import { onDestroy, onMount } from "svelte"
    import { ReactAdapter } from "./react-interop"

    type T = $$Generic<{}>

    export let component: React.FunctionComponent<T>
    export let props: T
    export let route: string

    let container: HTMLDivElement
    let root: Root|null

    onMount(() => {
        createRootIfNecessary(container)
        renderComponent(component, props, route)
    })
    onDestroy(() => {
        root?.unmount()
    })

    function createRootIfNecessary(container: HTMLElement) {
        if (!root) {
            root = createRoot(container)
        }
    }

    function renderComponent(component: React.FunctionComponent<T>, props: T, route: string) {
        root?.render(
            React.createElement(
                ReactAdapter,
                {
                    route,
                },
                React.createElement(
                    component,
                    props
                )
            )
        )
    }

    $: renderComponent(component, props, route)
</script>

<div bind:this={container} />
