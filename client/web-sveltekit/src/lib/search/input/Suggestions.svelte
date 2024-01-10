<script lang="ts">
    import { getSuggestionsState, type Option, type Group, type Action } from '$lib/branded'
    import { EditorView } from '@codemirror/view'
    import SuggestionOption from './SuggestionOption.svelte'
    import { createEventDispatcher, tick } from 'svelte'
    import { isSafari } from '@sourcegraph/common/src/util'
    import Icon from '$lib/Icon.svelte'
    import { mdiInformationOutline } from '@mdi/js'
    import ActionInfo from './ActionInfo.svelte'
    import { restrictToViewport } from '$lib/dom'

    const dispatch = createEventDispatcher<{ select: { option: Option; action: Action } }>()

    let suggestions: Group[] = []
    let open = false
    let activeRowIndex = -1
    let container: HTMLElement | undefined

    // This is crazy but this allows us to make this component a CodeMirror extension.
    // Doesn't work with HMR reloading though.
    export const extension = EditorView.updateListener.of(update => {
        const state = getSuggestionsState(update.state)
        if (state !== getSuggestionsState(update.startState)) {
            if (suggestions !== state.result.groups) {
                suggestions = state.result.groups
            }
            open = state.open
            activeRowIndex = state.selectedOption
        }
    })

    function handleSelection(event: MouseEvent) {
        const target = event.target as HTMLElement
        const match = target.closest('li[role="row"]')?.id.match(/\d+x\d+/)
        if (match) {
            // Extracts the group and row index from the elements ID to pass
            // the right option value to the callback.
            const [groupIndex, optionIndex] = match[0].split('x')
            const option = suggestions[+groupIndex].options[+optionIndex]
            // Determine which action was selected.
            dispatch('select', {
                option,
                action:
                    (target.closest<HTMLElement>('[data-action]')?.dataset?.action === 'secondary' &&
                        option.alternativeAction) ||
                    option.action,
            })
        }
    }

    async function scrollIntoView(container: HTMLElement) {
        // Wait for DOM to update
        await tick()
        // Options are not supported in Safari according to
        // https://developer.mozilla.org/en-US/docs/Web/API/Element/scrollIntoView#browser_compatibility
        container.querySelector('[aria-selected="true"]')?.scrollIntoView(isSafari() ? false : { block: 'nearest' })
    }

    $: flattenedRows = suggestions.flatMap(group => group.options)
    $: selectedOption = flattenedRows[activeRowIndex]
    $: show = open && suggestions.length > 0

    $: if (container && selectedOption) {
        // Scroll into view whenever selectedOption changes
        scrollIntoView(container)
    }
</script>

<div class="root" bind:this={container} use:restrictToViewport={{ offset: -20 }}>
    {#if show}
        <div class="suggestions" role="grid" tabindex="-1" on:mousedown={handleSelection}>
            {#each suggestions as group, groupIndex (group.title)}
                {#if group.options.length > 0}
                    <ul role="rowgroup" aria-labelledby="{groupIndex}-label">
                        <li id="{groupIndex}-lable" role="presentation">{group.title}</li>
                        {#each group.options as option, rowIndex (option)}
                            <SuggestionOption {groupIndex} {rowIndex} {option} selected={option === selectedOption} />
                        {/each}
                    </ul>
                {/if}
            {/each}
        </div>
        {#if selectedOption}
            <div class="footer">
                <span>
                    <ActionInfo action={selectedOption.action} shortcut="Enter" />{' '}
                    {#if selectedOption.alternativeAction}
                        <ActionInfo action={selectedOption.alternativeAction} shortcut="Mod+Enter" />
                    {/if}
                </span>
                <Icon svgPath={mdiInformationOutline} aria-hidden="true" inline />
            </div>
        {/if}
    {/if}
</div>

<style lang="scss">
    .root {
        --color: var(--icon-color);

        overflow-y: hidden;
        display: flex;
        flex-direction: column;

        .footer {
            display: flex;
            align-items: center;
            justify-content: space-between;
            border-top: 1px solid var(--border-color);
            padding: 0.5rem 1.25rem;
            flex: 0 0 auto;
            color: var(--text-muted);
        }

        .suggestions {
            overflow-y: auto;
            margin: 0 var(--suggestions-padding) var(--suggestions-padding);

            // Don't render the suggestions panel if we don't have any suggestions
            // in order to avoid extra paddings appearance
            &:empty {
                display: none;
            }

            ul {
                margin: 0;
                padding: 0;
                list-style: none;
                flex: 1;
            }

            [role='rowgroup'] {
                border-bottom: 1px solid var(--border-color);
                padding-top: 0.75rem;
                padding-bottom: 0.75rem;

                &:first-of-type {
                    padding-top: 0;
                }

                &:last-of-type {
                    border: none;
                    padding-bottom: 0;
                }

                // group header
                [role='presentation'] {
                    color: var(--text-muted);
                    font-size: 0.75rem;
                    font-weight: 500;
                    margin-bottom: 0.25rem;
                    padding: 0 0.5rem;
                }

                /*
                Layout of a row
                The icon is always top aligned next to the label.
                Label and description can wrap around if necessary, in which case the
                action labels are centered.
                On small screens the action labels are shown on a separate row.

                Normal:
                       ┌── inner-row ────────────────────────┐
                       │┌────────────────────────┐           │
               ┌──────┐││┌───────┐┌─────────────┐│┌─────────┐│
               │ Icon ││││ Label ││ Description │││ Actions ││
               └──────┘││└───────┘└─────────────┘│└─────────┘│
                       │└────────────────────────┘           │
                       └─────────────────────────────────────┘

                Wrapped description:
                       ┌─── inner-row ───────────────────────┐
                       │┌────────────────────────┐           │
               ┌──────┐││┌──────────────────────┐│           │
               │ Icon ││││ Label                ││┌─────────┐│
               └──────┘││└──────────────────────┘││ Actions ││
                       ││┌──────────────────────┐│└─────────┘│
                       │││ Description          ││           │
                       ││└──────────────────────┘│           │
                       │└────────────────────────┘           │
                       └─────────────────────────────────────┘

                Mobile:
                       ┌─── inner-row ───────────────────────┐
                       │┌───────────────────────────────────┐│
               ┌──────┐││┌─────────────────────────────────┐││
               │ Icon ││││ Label                           │││
               └──────┘││└─────────────────────────────────┘││
                       ││┌─────────────────────────────────┐││
                       │││ Description                     │││
                       ││└─────────────────────────────────┘││
                       │└───────────────────────────────────┘│
                       │┌───────────────────────────────────┐│
                       ││ Actions                           ││
                       │└───────────────────────────────────┘│
                       └─────────────────────────────────────┘

            */
            }
        }
    }
</style>
