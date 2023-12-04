<script lang="ts">
    import { mdiChevronDown, mdiChevronUp } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'
    import type { SidebarFilter } from '$lib/search/utils'
    import type { SearchSidebarSectionID } from '$lib/shared'
    import { temporarySetting } from '$lib/temporarySettings'
    import Tooltip from '$lib/Tooltip.svelte'
    import Button from '$lib/wildcard/Button.svelte'

    export let id: SearchSidebarSectionID
    export let items: SidebarFilter[]
    export let title: string

    const collapsedSections = temporarySetting('search.collapsedSidebarSections', {})

    $: sections = !$collapsedSections.loading && $collapsedSections.data ? $collapsedSections.data : null
    $: open = sections ? !sections[id] : true

    function toggleSection() {
        collapsedSections.setValue({ ...sections, [id]: open })
    }
</script>

{#if sections}
    <article>
        <Button variant="secondary" outline>
            <button
                slot="custom"
                let:className
                class="{className} header"
                type="button"
                aria-expanded={open}
                aria-label="{open ? 'Collapse' : 'Expand'} {title}"
                on:click={toggleSection}
            >
                <header><h5>{title}</h5></header>
                <Icon svgPath={open ? mdiChevronUp : mdiChevronDown} inline --color="var(--icon-color)" />
            </button>
        </Button>

        {#if open}
            <ul>
                {#each items as item}
                    <li>
                        <Button variant="secondary" outline>
                            <button
                                slot="custom"
                                let:className
                                class="{className} item"
                                on:click
                                data-value={item.value}
                                data-run={item.runImmediately}
                                data-focus-input={!item.runImmediately}
                            >
                                <span class="label">{item.label}</span>
                                {#if item.count !== undefined}
                                    <Tooltip tooltip="At least {item.count} results match this filter.">
                                        <span class="count">{item.count}</span>
                                    </Tooltip>
                                {/if}
                            </button>
                        </Button>
                    </li>
                {/each}
            </ul>
        {/if}
    </article>
{/if}

<style lang="scss">
    ul {
        margin: 0;
        padding: 0.125rem;
        padding-bottom: 1rem;
        list-style: none;
    }

    button {
        display: flex;
        width: 100%;
        align-items: center;
        border: none;
        text-align: left;
    }

    button.header {
        padding: 0.25rem;
        margin: 0.125rem 0;

        header {
            flex: 1;
        }
    }

    button.item {
        font-size: 0.75rem;
        padding: 0.25rem 0.375rem;
        margin: 0;
        font-weight: 400;

        .label {
            flex: 1;
        }

        .count {
            color: var(--link-color);
        }
    }
</style>
