<script context="module" lang="ts">
    import type { ComboboxOptionProps } from '@melt-ui/svelte'
    import type { RepositoryGitRefs_Repository_ } from './RepositoryRevPicker.gql'

    export type RepositoryBranches = RepositoryGitRefs_Repository_['gitRefs']
    export type RepositoryBranch = RepositoryBranches['nodes'][number]

    interface Node {
        id: string
    }

    interface Result<T extends Node = Node> {
        nodes: T[]
        totalCount: number
    }
</script>

<script lang="ts" generics="T extends Node">
    import { createCombobox } from '@melt-ui/svelte'

    import { Input, Alert } from '$lib/wildcard'
    import { createPromiseStore } from '$lib/utils'

    export let getData: (query: string) => Promise<Result<T>>
    export let onSelect: (branch: T) => void
    export let toOption: (branch: T) => ComboboxOptionProps<string>

    const {
        elements: { menu, input, option },
        states: { inputValue },
    } = createCombobox<any>({
        portal: null,
        forceVisible: true,
        scrollAlignment: 'nearest',
        closeOnOutsideClick: false,
        onSelectedChange: ({ next }) => {
            const selectedBranch = $data.value?.nodes.find(d => d.id === next?.value)

            if (selectedBranch) {
                onSelect(selectedBranch)
            }

            return next
        },
    })

    let debounceTimer: ReturnType<typeof setTimeout>
    const data = createPromiseStore<Result<T>>()

    // Start query initial suggestion
    data.set(getData($inputValue))

    const debounce = (callback: () => void) => {
        clearTimeout(debounceTimer)
        debounceTimer = setTimeout(callback, 450)
    }

    $: {
        debounce(() => {
            data.set(getData($inputValue))
        })
    }

    $: filteredData = $data.value ? $data.value.nodes : []
</script>

<div class="root">
    <Input
        {...$input}
        actions={[input]}
        loading={$data.pending}
        autofocus={true}
        placeholder="Search branches..."
    />

    <div {...$menu} use:menu class="suggestions">
        <ul class="suggestion-list">
            <!-- Initial loading state (but don't show state if data from prev call is presented) -->
            {#if !$data.value && $data.pending}
                <li class="no-data-state">Loading...</li>
            {/if}

            <!-- Error state (show error immediately) -->
            {#if $data.error}
                <li class="no-data-state">
                    <Alert variant="danger">
                        Unable to load branches information: {$data.error.message}
                    </Alert>
                </li>
            {/if}

            {#if !$data.error}
                {#each filteredData as value (value.id)}
                    <li use:option {...$option(toOption(value))} class="suggestion-list-item">
                        <slot {value}/>
                    </li>
                {/each}
            {/if}

            {#if filteredData.length === 0 && !$data.pending && !$data.error}
                <li class="no-data-state">
                    <slot name="no-data" inputValue={$inputValue}/>
                </li>
            {/if}
        </ul>
    </div>

    <footer class="footer">
        <a href={`/-/branches`}>
            See all commits
            {#if !$data.error && $data.value}({$data.value.totalCount}){/if}
        </a>
    </footer>
</div>

<style lang="scss">
    .root {
        max-height: 25rem;
        display: flex;
        flex-direction: column;
    }

    .suggestions {
        flex-grow: 1;
        min-height: 0;
        overflow: auto;
        margin: 0.5rem -0.5rem 0rem -0.5rem;

        // There is no way to turn off styles that come from
        // melt UI popover element, since we render suggestion
        // no in the melt UI popover we turn it off via CSS here.
        position: static !important;
        width: calc(100% + 1rem) !important;
    }

    .suggestion-list {
        display: grid;
        grid-template-rows: auto;
        grid-template-columns: [title] auto [author] min-content [timestamp] min-content;
        padding: 0 0 0.5rem 0;
        margin: 0;
        list-style: none;
        height: 100%;
        overflow: auto;
    }

    .suggestion-list-item {
        --avatar-size: 1.5rem;

        display: grid;
        grid-column: 1 / 4;
        grid-template-columns: subgrid;
        padding: 0.25rem;
        cursor: pointer;
        gap: 0.5rem;
        border-bottom: 1px solid var(--border-color);

        &:last-child {
            border-bottom: none;
        }

        &:hover,
        &[data-highlighted] {
            background: var(--color-bg-3);
        }
    }

    .timestamp {
        padding-right: 0.75rem;
    }

    .no-data-state {
        grid-column: span 3;
        display: flex;
        align-items: center;
        justify-content: center;
        margin: 2rem;
        color: var(--text-muted);
    }

    .footer {
        margin: 0 -0.5rem -0.5rem -0.5rem;
        border-top: 1px solid var(--border-color);

        a {
            padding: 0.5rem;
            width: 100%;
            height: 100%;
            display: flex;
            justify-content: center;
            align-items: center;

            &:hover {
                background: var(--color-bg-2);
            }
        }
    }
</style>
