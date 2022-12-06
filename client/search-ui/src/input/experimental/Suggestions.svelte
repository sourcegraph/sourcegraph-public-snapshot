<svelte:options immutable />

<script lang="ts">
    import { Group, Option } from './suggestions'
    import Icon from './Icon.svelte'
    import { createEventDispatcher } from 'svelte'

    export let id: string
    export let results: Group[]
    export let activeRowIndex = -1
    export let open = false

    let dispatch = createEventDispatcher<{ select: Option }>()
    let container: HTMLElement
    let windowHeight: number

    $: maxHeight = container ? windowHeight - container.getBoundingClientRect().top - 20 : 'auto'
    $: flattenedRows = results.flatMap(group => group.options)
    $: focusedItem = flattenedRows[activeRowIndex]
    $: show = open && results.length > 0

    function handleSelection(event: MouseEvent) {
        const match = (event.target as HTMLElement).closest('li[role="row"]')?.id.match(/\d+x\d+/)
        if (match) {
            const [group, option] = match[0].split('x')
            dispatch('select', results[+group].options[+option])
        }
    }

    function getNote(option: Option): string {
        switch (option.type) {
            case 'completion':
                return 'Add'
            case 'target':
                return 'Jump to'
            case 'command':
                return option.note ?? ''
        }
    }
</script>

<svelte:window bind:innerHeight={windowHeight} />
{#if show}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <div {id} role="grid" style="max-height: {maxHeight}px" on:mousedown={handleSelection} bind:this={container}>
        {#each results as group, groupIndex (group.title)}
            <ul role="rowgroup" aria-labelledby="{id}-{groupIndex}-label">
                <li id="{id}-{groupIndex}-label" role="presentation">{group.title}</li>
                {#each group.options as option, rowIndex (option)}
                    <li role="row" id="{id}-{groupIndex}x{rowIndex}" aria-selected={focusedItem === option}>
                        {#if option.icon}
                            <div class="pr-1">
                                <Icon path={option.icon} />
                            </div>
                        {/if}
                        <div role="gridcell" class="content">
                            {#if option.render}
                                <svelte:component this={option.render} {option} />
                            {:else if option.matches}
                                {#each option.value as char, index}
                                    {#if option.matches.has(index)}
                                        <span class="match">{char}</span>
                                    {:else}
                                        {char}
                                    {/if}
                                {/each}
                            {:else}
                                {option.value}
                            {/if}
                        </div>
                        {#if option.description}
                            <div role="gridcell" class="description">
                                {option.description}
                            </div>
                        {/if}
                        <div role="gridcell" class="note">
                            {getNote(option)}
                        </div>
                    </li>
                {/each}
            </ul>
        {/each}
    </div>
{/if}

<style lang="scss">
    [role='grid'] {
        overflow-y: auto;
    }

    ul {
        margin: 0;
        padding: 0;
        list-style: none;
    }

    [role='rowgroup'] {
        border-bottom: 1px solid var(--border-color);
        padding: 12px;

        &:first-of-type {
            padding-top: 0;
        }

        &:last-of-type {
            border: none;
        }

        &:last-of-type {
            border: none;
        }

        // group header
        [role='presentation'] {
            color: var(--text-muted);
            font-size: 12px;
            font-weight: 500;
            margin-bottom: 0.25rem;
            padding: 0 8px;
        }

        [role='row'] {
            display: flex;
            align-items: center;
            padding: 4px 8px;
            border-radius: var(--border-radius);
            font-family: var(--code-font-family);
            font-size: 12px;
            min-height: 24px;

            &[aria-selected='true'] {
                background-color: var(--subtle-bg);
                border-radius: 4px;
            }

            &:hover {
                background-color: var(--color-bg-2);
                cursor: pointer;
            }

            .match {
                font-weight: bold;
            }

            .description {
                margin-left: 8px;
                color: var(--input-placeholder-color);
            }

            .note {
                font-size: 12px;
                margin-left: auto;
                color: var(--text-muted);
                font-family: var(--font-family-base);
            }
        }
    }
</style>
