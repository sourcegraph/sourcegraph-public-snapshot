<script lang="ts" context="module">
    export interface Tab {
        id: string
        title: string
    }

    export interface TabsContext {
        id: string
        selectedTab: Readable<string>
        register(tab: Tab): void
    }

    export const KEY = {}
</script>

<script lang="ts">
    import * as uuid from 'uuid'
    import { setContext } from 'svelte'
    import { derived, writable, type Readable, type Writable } from 'svelte/store'

    /**
     * The index of the tab that should be selected by default.
     */
    export let initial: number = 0

    const id = uuid.v4()
    const tabs: Writable<Tab[]> = writable([])
    const selectedTab = writable(initial)

    setContext<TabsContext>(KEY, {
        id,
        selectedTab: derived([tabs, selectedTab], ([$tabs, $selectedTab]) => $tabs[$selectedTab]?.id ?? null),
        register(tab: Tab) {
            tabs.update(tabs => {
                if (tabs.some(existingTab => existingTab.id === tab.id)) {
                    return tabs
                }
                return [...tabs, tab]
            })
        },
    })
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
                on:click={() => ($selectedTab = index)}>{tab.title}</button
            >
        {/each}
    </div>
    <slot />
</div>

<style lang="scss">
    .tabs {
        display: flex;
        flex-direction: column;
        align-items: center;
    }

    .tabs-header {
        display: flex;
        gap: 1rem;
    }

    button {
        cursor: pointer;
        border: none;
        background: none;
        align-items: center;
        letter-spacing: normal;
        margin: 0;
        min-height: 2rem;
        padding: 0 0.125rem;
        color: var(--body-color);
        text-transform: none;
        display: inline-flex;
        flex-direction: column;
        justify-content: center;
        border-bottom: 2px solid transparent;

        &[aria-selected='true'] {
            color: var(--body-color);
            font-weight: 700;
            border-bottom: 2px solid var(--brand-secondary);
        }
    }
</style>
