<script lang="ts" context="module">
    const actionNames: Record<Action['type'], string> = {
        completion: 'Add',
        goto: 'Go to',
        command: 'Run',
    }
</script>

<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import { type Option, type Action, RenderAs } from '$lib/branded'
    import SyntaxHighlightedQuery from '../SyntaxHighlightedQuery.svelte'
    import EmphasizedLabel from '../EmphasizedLabel.svelte'

    export let option: Option
    export let groupIndex: number
    export let rowIndex: number
    export let selected: boolean

    function getFieldValue(option: Option): { field: string; value: string } {
        let field = ''
        let value = ''

        if (option.render === RenderAs.FILTER) {
            const separatorIndex = option.label.indexOf(':')
            if (separatorIndex > -1) {
                field = option.label.slice(0, separatorIndex)
                value = option.label.slice(separatorIndex + 1)
            } else {
                field = option.label
            }
        }

        return { field, value }
    }

    $: ({ field, value } = getFieldValue(option))
</script>

<li role="row" id="{groupIndex}x{rowIndex}" aria-selected={selected}>
    {#if option.icon}
        <div class="pr-1 align-self-start">
            <Icon svgPath={option.icon} aria-hidden="true" inline />
        </div>
    {/if}
    <div class="inner-row">
        <div class="d-flex flex-wrap">
            <div role="gridcell" class="label test-option-label">
                {#if field}
                    <span class="filter-option">
                        <span class="search-filter-keyword">
                            <EmphasizedLabel label={field} matches={option.matches} />
                        </span>
                        <span class="separator">:</span>
                        {#if value}
                            <span
                                ><EmphasizedLabel
                                    label={value}
                                    matches={option.matches}
                                    offset={field.length + 1}
                                /></span
                            >
                        {/if}
                    </span>
                {:else if option.render === RenderAs.QUERY}
                    <SyntaxHighlightedQuery query={option.label} matches={option.matches} />
                {:else}
                    <EmphasizedLabel label={option.label} matches={option.matches} />
                {/if}
            </div>
            {#if option.description}
                <div role="gridcell" class="description">
                    {option.description}
                </div>
            {/if}
        </div>
        <div class="note">
            <div role="gridcell" data-action="primary">
                {option.action.name ?? actionNames[option.action.type]}
            </div>
            {#if option.alternativeAction}
                <div role="gridcell" data-action="secondary">
                    {option.alternativeAction.name ?? actionNames[option.alternativeAction.type]}
                </div>
            {/if}
        </div>
    </div>
</li>

<style lang="scss">
    [role='row'] {
        --color: var(icon-color);

        display: flex;
        align-items: center;
        padding: 0.25rem 0.5rem;
        border-radius: var(--border-radius);
        font-family: var(--code-font-family);
        font-size: 0.75rem;
        min-height: 1.5rem;

        &[aria-selected='true'] {
            background-color: var(--subtle-bg);
            border-radius: 4px;
        }

        &:hover {
            background-color: var(--color-bg-2);
            cursor: pointer;
        }

        // Used to make label and actions wrappable
        .inner-row {
            display: flex;
            flex: 1;
            align-items: center;

            @media (--xs-breakpoint-down) {
                flex-direction: column;
                align-items: start;
                gap: 0.25rem;
            }
        }

        .label {
            margin-right: 0.5rem;
        }

        .description {
            color: var(--input-placeholder-color);
        }

        .note {
            font-size: 0.75rem;
            margin-left: auto;
            color: var(--text-muted);
            font-family: var(--font-family-base);
            display: flex;
            white-space: nowrap;

            @media (--xs-breakpoint-down) {
                margin-left: 0;
            }

            > [role='gridcell'] {
                padding: 0 0.5rem;

                &:hover {
                    text-decoration: underline;
                }

                + [role='gridcell'] {
                    border-left: 1px solid var(--border-color-2);
                }

                @media (--xs-breakpoint-down) {
                    &:first-of-type {
                        padding-left: 0;
                    }
                }
            }
        }
    }

    .filter-option {
        font-family: var(--code-font-family);
        font-size: 0.75rem;
        display: flex; // to remove whitespace around the filter parts

        .separator {
            color: var(--search-filter-keyword-color);
        }
    }
</style>
