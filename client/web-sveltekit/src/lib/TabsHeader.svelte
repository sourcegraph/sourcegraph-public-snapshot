<script lang="ts" context="module">
    import type { Keys } from '$lib/Hotkey'

    export interface Tab {
        id: string
        title: string
        // An icon for the tab. Shown to the left of the title.
        icon?: ComponentProps<Icon>['icon']
        // A shortcut to activate the tab. Shown to the right of the title.
        shortcut?: Keys
        // If provided, will cause the tab to be rendered as a link
        href?: string
    }
</script>

<script lang="ts">
    import { createEventDispatcher, type ComponentProps } from 'svelte'

    import KeyboardShortcut from '$lib/KeyboardShortcut.svelte'

    import Icon from './Icon.svelte'

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
        <svelte:element
            this={tab.href ? 'a' : 'button'}
            id="{id}--tab--{index}"
            aria-controls={tab.id}
            aria-selected={selected === index}
            tabindex={selected === index ? 0 : -1}
            role="tab"
            on:click={selectTab}
            data-tab
            href={tab.href}
        >
            {#if tab.icon}
                <Icon icon={tab.icon} aria-hidden inline />
            {/if}
            <span data-tab-title={tab.title}>
                {tab.title}
            </span>
            {#if tab.shortcut}
                <KeyboardShortcut shortcut={tab.shortcut} />
            {/if}
        </svelte:element>
    {/each}
</div>

<style lang="scss">
    .tabs-header {
        --icon-color: var(--header-icon-color);

        display: flex;
        align-items: stretch;
        justify-content: var(--tabs-header-align, flex-start);
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
            width: 100%;
            border-bottom: 2px solid transparent;
        }

        &:hover {
            --icon-color: currentColor;

            color: var(--text-title);
            background-color: var(--secondary-2);
        }

        &[aria-selected='true'] {
            --icon-color: currentColor;

            color: var(--primary);

            &::after {
                border-color: var(--primary);
            }
        }

        span {
            display: inline-block;

            &[data-tab-title] {
                // Hidden rendering of the bold tab title to prevent
                // shifting when the tab is selected.
                &::before {
                    content: attr(data-tab-title);
                    display: block;
                    height: 0;
                    visibility: hidden;
                }
            }
        }

        &[aria-selected='true'] span,
        span::before {
            // Hidden rendering of the bold tab title to prevent
            // shifting when the tab is selected.
            font-weight: 500;
        }
    }
</style>
