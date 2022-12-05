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
                        <div role="gridcell">
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
        padding: 0.5rem;

        &:last-of-type {
            border: none;
        }

        // group header
        [role='presentation'] {
            color: var(--text-muted);
            margin-bottom: 0.25rem;
        }

        [role='row'] {
            display: flex;
            align-items: center;
            padding: 0.25rem 0.5rem;
            border-radius: var(--border-radius);

            &[aria-selected='true'] {
                background-color: var(--subtle-bg);
            }

            &:hover {
                background-color: var(--color-bg-2);
            }

            .match {
                font-weight: bold;
            }

            .description {
                margin-left: 0.5rem;
                color: var(--input-placeholder-color);
            }

            .note {
                margin-left: auto;
                color: var(--text-muted);
            }
        }
    }
</style>
