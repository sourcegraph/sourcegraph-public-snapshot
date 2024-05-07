<script lang="ts" context="module">
    export interface Tab {
        id: string
        title: string
        icon?: string
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

    import Icon from './Icon.svelte'

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
        const index = (event.target as HTMLElement).closest('[role="tab"]')?.id.match(/\d+$/)?.[0]
        if (index) {
            $selectedTab = $selectedTab === +index && toggable ? null : +index
            dispatch('select', $selectedTab)
        }
    }
</script>

<div class="tabs" data-tabs>
    <div class="tabs-header" role="tablist" data-tab-header>
        {#each $tabs as tab, index (tab.id)}
            <button
                id="{id}--tab--{index}"
                aria-controls={tab.id}
                aria-selected={$selectedTab === index}
                tabindex={$selectedTab === index ? 0 : -1}
                role="tab"
                on:click={selectTab}
                data-tab
                >{#if tab.icon}<Icon svgPath={tab.icon} aria-hidden inline /> {/if}<span data-tab-title={tab.title}
                    >{tab.title}</span
                ></button
            >
        {/each}
    </div>
    <slot />
</div>

<style lang="scss">
    .tabs {
        display: flex;
        flex-direction: column;
        height: 100%;
    }

    .tabs-header {
        --icon-fill-color: var(--header-icon-color);

        display: flex;
        align-items: stretch;
        justify-content: var(--align-tabs, center);
        gap: var(--tabs-gap, 0);
        border-bottom: 1px solid var(--border-color);
    }

    [role='tab'] {
        all: unset;

        cursor: pointer;
        align-items: center;
        min-height: 2rem;
        padding: 0.25rem 0.75rem;
        color: var(--text-body);
        display: inline-flex;
        flex-flow: row nowrap;
        justify-content: center;
        white-space: nowrap;
        gap: 0.5rem;
        position: relative;

        &::after {
            content: '';
            display: block;
            position: absolute;
            bottom: 0;
            transform: translateY(50%);
            width: 100%;
            border-bottom: 2px solid transparent;
        }

        &:hover {
            color: var(--text-title);
            background-color: var(--color-bg-2);
        }

        &[aria-selected='true'] {
            font-weight: 500;
            color: var(--text-title);

            &::after {
                border-color: var(--brand-secondary);
            }
        }

        span {
            display: inline-block;

            // Hidden rendering of the bold tab title to prevent
            // shifting when the tab is selected.
            &::before {
                content: attr(data-tab-title);
                display: block;
                font-weight: 500;
                height: 0;
                visibility: hidden;
            }
        }
    }
</style>
