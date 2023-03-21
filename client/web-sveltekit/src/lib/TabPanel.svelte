<script lang="ts">
    import { getContext } from 'svelte'
    import { type TabsContext, KEY } from './Tabs.svelte'
    import * as uuid from 'uuid'

    export let title: string

    const context = getContext<TabsContext>(KEY)
    const id = uuid.v4()
    const tabId = `${context.id}-tab-${id}`
    context.register({
        id: tabId,
        title,
    })
    $: selectedId = context.selectedTab
    $: selected = $selectedId === tabId
</script>

{#if selected}
    <div id="{context.id}-panel-{id}" aria-labelledby={tabId} role="tabpanel" tabindex={selected ? 0 : -1}>
        <slot />
    </div>
{/if}
