<script context="module" lang="ts">
    import type { ComboboxOptionProps } from '@melt-ui/svelte'
    import type { RepositoryGitRefs_Repository_ } from './RepositoryRevPicker.gql'

    export type RepositoryBranches = RepositoryGitRefs_Repository_['gitRefs']

    const toOption = (branch: any): ComboboxOptionProps<any> => ({
        value: branch.id,
        label: branch.displayName,
    })
</script>

<script lang="ts">
    import { mdiSourceBranch } from '@mdi/js'
    import { goto } from '$app/navigation'
    import { createCombobox } from '@melt-ui/svelte'
    import { replaceRevisionInURL } from '@sourcegraph/shared/src/util/url'

    import Icon from '$lib/Icon.svelte'
    import Avatar from '$lib/Avatar.svelte'
    import Timestamp from '$lib/Timestamp.svelte'
    import { Input, Alert } from '$lib/wildcard'
    import { createPromiseStore } from '$lib/utils'

    export let getRepositoryBranches: (query: string) => Promise<RepositoryBranches>
    export let onSelect: () => void

    const {
        elements: { menu, input, option },
        states: { inputValue },
    } = createCombobox<any>({
        portal: null,
        forceVisible: true,
        scrollAlignment: 'nearest',
        onSelectedChange: ({ next }) => {
            const selectedBranch = $repositoryBranches.value?.nodes.find(branch => branch.id === next?.value)

            if (selectedBranch) {
                goto(
                    replaceRevisionInURL(location.pathname + location.search + location.hash, selectedBranch.abbrevName)
                )
            }

            onSelect()
            return next
        },
    })

    let debounceTimer: ReturnType<typeof setTimeout>
    const repositoryBranches = createPromiseStore<RepositoryBranches>()

    // Start query initial suggestion
    repositoryBranches.set(getRepositoryBranches($inputValue))

    const debounce = (callback: () => void) => {
        clearTimeout(debounceTimer)
        debounceTimer = setTimeout(callback, 450)
    }

    $: {
        debounce(() => {
            repositoryBranches.set(getRepositoryBranches($inputValue))
        })
    }

    $: filteredBranches = $repositoryBranches.value ? $repositoryBranches.value.nodes : []
</script>

<div class="root">
    <Input {...$input} actions={[input]} autofocus={true} placeholder="Search branches..." />

    <div {...$menu} use:menu class="suggestions">
        <ul class="suggestion-list">
            <!-- Initial loading state (but don't show state if data from prev call is presented) -->
            {#if !$repositoryBranches.value && $repositoryBranches.pending}
                <li class="no-data-state">Loading...</li>
            {/if}

            <!-- Error state (show error immediately) -->
            {#if $repositoryBranches.error}
                <li class="no-data-state">
                    <Alert variant="danger">
                        Unable to load branches information: {$repositoryBranches.error.message}
                    </Alert>
                </li>
            {/if}

            {#if !$repositoryBranches.error}
                {#each filteredBranches as branch (branch.id)}
                    <li use:option {...$option(toOption(branch))} class="suggestion-list-item">
                        <span class="title">
                            <Icon svgPath={mdiSourceBranch} inline />
                            <span>{branch.displayName}</span>
                        </span>
                        <span class="author">
                            {#if branch.target.commit}
                                <Avatar avatar={branch.target.commit?.author.person} />
                                {branch.target.commit.author.person.displayName}
                            {/if}
                        </span>
                        <span class="timestamp">
                            {#if branch.target.commit}
                                <Timestamp date={branch.target.commit.author.date} strict />
                            {/if}
                        </span>
                    </li>
                {/each}
            {/if}

            {#if filteredBranches.length === 0 && !$repositoryBranches.pending && !$repositoryBranches.error}
                <li class="no-data-state">
                    No branches matching&nbsp;<b>{$inputValue}</b>, try different search query
                </li>
            {/if}
        </ul>
    </div>

    <footer class="footer">
        <a href="">
            See all commits
            {#if !$repositoryBranches.error && $repositoryBranches.value}({$repositoryBranches.value.totalCount}){/if}
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
        gap: 0.25rem;
        border-bottom: 1px solid var(--border-color);

        &:last-child {
            border-bottom: none;
        }

        &:hover,
        &[data-highlighted] {
            background: var(--color-bg-3);
        }
    }

    .title,
    .author,
    .timestamp {
        display: flex;
        gap: 0.25rem;
        align-items: center;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .title {
        padding-left: 0.75rem;
        padding-right: 0.5rem;

        // Branch icon
        :global(svg) {
            flex-shrink: 0;
        }

        // Branch name badge
        :global(span) {
            overflow: hidden;
            text-overflow: ellipsis;
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
