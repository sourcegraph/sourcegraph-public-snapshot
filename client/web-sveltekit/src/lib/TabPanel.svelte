<script lang="ts">
    import { getContext, onDestroy } from 'svelte'
    import * as uuid from 'uuid'

    import type { Keys } from '$lib/Hotkey'
    import { type TabsContext, KEY } from './Tabs.svelte'
    import { registerHotkey } from '$lib/Hotkey'

    export let title: string
    export let shortcut: Keys | undefined = undefined

    const id = uuid.v4()
    const context = getContext<TabsContext>(KEY)
    const tabId = `${context.id}-tab-${id}`

    if (shortcut) {
        registerHotkey({
            keys: shortcut!,
            allowDefault: false,
            ignoreInputFields: false,
            handler: event => {
                event.preventDefault()

                const currentTabIndex = context.getTabs().findIndex(tab => tab.id === tabId)
                context.selectTab(currentTabIndex)

                return false
            },
        })
    }

    onDestroy(
        context.register({
            id: tabId,
            title,
            shortcut,
        })
    )
    $: selectedId = context.selectedTabID
    $: selected = $selectedId === tabId
</script>

{#if selected}
    <div
        id="{context.id}-panel-{id}"
        aria-labelledby={tabId}
        role="tabpanel"
        tabindex={selected ? 0 : -1}
        data-tab-panel={title}
    >
        <slot />
    </div>
{/if}

<style lang="scss">
    div {
        flex: 1;
        min-height: 0;
        overflow: hidden;
    }
</style>
