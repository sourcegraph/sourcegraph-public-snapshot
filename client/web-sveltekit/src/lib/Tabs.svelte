<script lang="ts" context="module">
    import type { Tab } from './TabsHeader.svelte'

    export type { Tab }

    export interface TabsContext {
        id: string
        selectedTabID: Readable<string | null>
        register(tab: Tab): Unsubscriber
        getTabs: () => Tab[]
        selectTab: (selectedTabIndex: number) => void
    }

    export const KEY = {}
</script>

<script lang="ts">
    import { createEventDispatcher, setContext } from 'svelte'
    import { derived, writable, type Readable, type Writable, type Unsubscriber } from 'svelte/store'
    import * as uuid from 'uuid'

    import TabsHeader from './TabsHeader.svelte'

    /**
     * The index of the tab that should be selected by default.
     */
    export let selected: number | null = 0
    export let toggable = false

    const dispatch = createEventDispatcher<{ select: number | null }>()
    const id = uuid.v4()
    const tabs: Writable<Tab[]> = writable([])
    const selectedTab = writable(selected)
    $: $selectedTab = selected

    setContext<TabsContext>(KEY, {
        id,
        selectedTabID: derived([tabs, selectedTab], ([$tabs, $selectedTab]) => {
            if ($selectedTab === null) {
                return null
            }
            return $tabs[$selectedTab]?.id ?? null
        }),
        register(tab: Tab) {
            tabs.update(tabs => {
                if (tabs.some(existingTab => existingTab.id === tab.id)) {
                    return tabs
                }
                return [...tabs, tab]
            })
            return () => {
                tabs.update(tabs => tabs.filter(existingTab => existingTab.id !== tab.id))
            }
        },
        getTabs: () => $tabs,
        selectTab: (index: number): void => {
            $selectedTab = $selectedTab === index && toggable ? null : index
            dispatch('select', $selectedTab)
        },
    })

    function selectTab(event: { detail: number }) {
        $selectedTab = $selectedTab === event.detail && toggable ? null : event.detail
        dispatch('select', $selectedTab)
    }
</script>

<div class="tabs" data-tabs>
    <TabsHeader {id} tabs={$tabs} selected={$selectedTab} on:select={selectTab} />
    <slot />
</div>

<style lang="scss">
    .tabs {
        display: flex;
        flex-direction: column;
        height: 100%;
    }
</style>
