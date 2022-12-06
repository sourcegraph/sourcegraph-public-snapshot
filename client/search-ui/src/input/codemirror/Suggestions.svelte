<svelte:options immutable />

<script lang="ts">
    import { Group, Option } from './suggestions'
    import Icon from './Icon.svelte'

    export let id: string
    export let results: Group[]
    export let activeRowIndex = -1

    $: flattenedRows = results.flatMap(group => group.entries)
    $: focusedItem = flattenedRows[activeRowIndex]
    $: hasResults = results.length > 0

    function handleClick(event: MouseEvent) {
        // TODO: Handle application on click
        console.log(event.target)
        event.preventDefault()
    }

    function getNote(option: Option): string {
        switch (option.type) {
            case 'target':
                return 'Jump to'
            case 'command':
                return ''
            case 'completion':
                return 'Add'
        }
    }
</script>

{#if hasResults}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <div {id} role="grid" on:click={handleClick}>
        {#each results as group, groupIndex (group.title)}
            <ul role="rowgroup" aria-labelledby="{id}-{groupIndex}-label">
                <li id="{id}-{groupIndex}-label" role="presentation">{group.title}</li>
                {#each group.entries as option, rowIndex (option)}
                    <li role="row" id="{id}-item-{groupIndex}x{rowIndex}" aria-selected={focusedItem === option}>
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
    ul {
        margin: 0;
        padding: 0;
        list-style: none;
    }

    [role='rowgroup'] {
        border-bottom: 1px solid var(--border-color);
        padding: 12px;

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
            padding: 6px 8px;
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
                font-weight: 500;
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
