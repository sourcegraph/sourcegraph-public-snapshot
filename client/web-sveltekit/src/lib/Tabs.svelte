<script lang="ts" context="module">
    export interface Tab {
        id: string
        title: string
    }

    export interface TabsContext {
        id: string
        selectedTabID: Readable<string | null>
        register(tab: Tab): Unsubscriber
    }

    export const KEY = {}
</script>

<script lang="ts">
    import { createEventDispatcher, setContext } from 'svelte'
    import { derived, writable, type Readable, type Writable, type Unsubscriber } from 'svelte/store'
    import * as uuid from 'uuid'

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
    })

    function selectTab(event: MouseEvent) {
        const index = (event.target as HTMLElement).id.match(/\d+$/)?.[0]
        if (index) {
            $selectedTab = $selectedTab === +index && toggable ? null : +index
            dispatch('select', $selectedTab)
        }
    }
</script>

<div class="tabs">
    <div class="tabs-header" role="tablist">
        {#each $tabs as tab, index (tab.id)}
            <button
                id="{id}--tab--{index}"
                aria-controls={tab.id}
                aria-selected={$selectedTab === index}
                tabindex={$selectedTab === index ? 0 : -1}
                role="tab"
                on:click={selectTab}>{tab.title}</button
            >
        {/each}
    </div>
    <slot />
</div>

<style lang="scss">
    .tabs {
        display: flex;
        flex-direction: column;
    }

    .tabs-header {
        display: flex;
        gap: 1rem;
        justify-content: var(--align-tabs, center);
    }

    button {
        cursor: pointer;
        border: none;
        background: none;
        align-items: center;
        letter-spacing: normal;
        margin: 0;
        min-height: 2rem;
        padding: 0 0.25rem;
        color: var(--body-color);
        text-transform: none;
        display: inline-flex;
        flex-direction: column;
        justify-content: center;
        border-bottom: 2px solid transparent;

        &[aria-selected='true'],
        &:hover {
            color: var(--body-color);
            background-color: var(--color-bg-2);
        }

        &[aria-selected='true'] {
            font-weight: 700;
        }
    }
</style>
