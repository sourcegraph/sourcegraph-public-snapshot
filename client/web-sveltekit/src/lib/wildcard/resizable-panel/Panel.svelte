<script lang="ts">
    import { getContext, onDestroy } from 'svelte'
    import type { PanelGroupContext } from './types'
    import { getId } from './utils/common'

    export let id: string | null = null
    export let minSize: number | undefined = undefined
    export let maxSize: number | undefined = undefined
    export let defaultSize: number | undefined = undefined
    export let order: number | undefined = undefined
    export let collapsible: boolean | undefined = undefined
    export let collapsedSize: number | undefined = undefined

    const panelId = id ?? getId()
    let panelElement: HTMLElement

    const { groupId, getPanelStyles, registerPanel, expandPanel, collapsePanel, isPanelCollapsed } =
        getContext<PanelGroupContext>('panel-group-context')

    // TODO: Support update registry as any of panelInfo deps change
    const panelInfo = {
        order,
        id: panelId,
        idFromProps: id,
        constraints: {
            minSize,
            maxSize,
            defaultSize,
            collapsible,
            collapsedSize,
        },
        getPanelElement: () => panelElement,
    }

    // External imperative Panel API
    export function collapse(): void {
        collapsePanel(panelInfo)
    }

    export function expand(): void {
        expandPanel(panelInfo)
    }

    export function isCollapsed() {
        return isPanelCollapsed(panelInfo)
    }

    export function isExpanded() {
        return !isPanelCollapsed(panelInfo)
    }

    onDestroy(registerPanel(panelInfo))

    $: styles = getPanelStyles(panelId)
</script>

<div
    id={panelId}
    class="panel"
    style={$styles}
    bind:this={panelElement}
    data-panel
    data-panel-id={panelId}
    data-panel-group-id={groupId}
>
    <slot />
</div>

<style>
    .panel {
        flex-basis: 0;
        flex-shrink: 0;

        /* Without this, Panel sizes may be unintentionally overridden
         by their content */
        overflow: hidden;
    }
</style>
