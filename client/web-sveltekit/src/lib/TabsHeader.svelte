<script lang="ts" context="module">
    import type { Keys } from '$lib/Hotkey'

    export interface Tab {
        id: string
        title: string
        shortcut?: Keys
    }
</script>

<script lang="ts">
    import { createEventDispatcher } from 'svelte'
    import KeyboardShortcut from '$lib/KeyboardShortcut.svelte'

    export let id: string
    export let tabs: Tab[]
    export let selected: number | null = 0

    const dispatch = createEventDispatcher<{ select: number }>()

    function selectTab(event: MouseEvent) {
        const index = (event.target as HTMLElement).closest('[role="tab"]')?.id.match(/\d+$/)?.[0]
        if (index) {
            dispatch('select', +index)
        }
    }
</script>

<div class="tabs-header" role="tablist" data-tab-header>
    {#each tabs as tab, index (tab.id)}
        <button
            id="{id}--tab--{index}"
            aria-controls={tab.id}
            aria-selected={selected === index}
            tabindex={selected === index ? 0 : -1}
            role="tab"
            on:click={selectTab}
            data-tab
        >
            <span data-tab-title={tab.title}>
                {tab.title}
            </span>
            {#if tab.shortcut}
                <KeyboardShortcut shorcut={tab.shortcut} />
            {/if}
        </button>
    {/each}
</div>

<style lang="scss">
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
        gap: 0.25rem;
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
            background-color: var(--secondary-2);
        }

        &[aria-selected='true'] {
            font-weight: 500;
            color: var(--text-title);
            background-color: var(--secondary-2);

            :global(kbd) {
                color: white;
                box-shadow: none;
                border-color: var(--primary);
                background-color: var(--primary);
            }

            &::after {
                border-color: var(--primary);
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
