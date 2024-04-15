<script context="module" lang="ts">
    import type { ComboboxOptionProps } from '@melt-ui/svelte'
    import type { RepositoryGitCommits_Repository_, RepositoryGitCommit } from './RepositoryRevPicker.gql'

    export type { RepositoryGitCommit }
    export type RepositoryCommits = RepositoryGitCommits_Repository_['commit']

    const toOption = (commit: RepositoryGitCommit): ComboboxOptionProps<any> => ({
        value: commit.id,
        label: commit.oid,
    })
</script>

<script lang="ts">
    import { mdiSourceCommit } from '@mdi/js'
    import { createCombobox } from '@melt-ui/svelte'

    import Icon from '$lib/Icon.svelte'
    import Avatar from '$lib/Avatar.svelte'
    import Timestamp from '$lib/Timestamp.svelte'
    import { Input, Alert, Badge } from '$lib/wildcard'
    import { createPromiseStore } from '$lib/utils'

    export let repoURL: string
    export let getRepositoryCommits: (query: string) => Promise<RepositoryCommits>
    export let onSelect: (commit: RepositoryGitCommit) => void

    const {
        elements: { menu, input, option },
        states: { inputValue },
    } = createCombobox<any>({
        portal: null,
        forceVisible: true,
        scrollAlignment: 'nearest',
        closeOnOutsideClick: false,
        onSelectedChange: ({ next }) => {
            const selectedCommit = $repositoryCommits.value?.ancestors.nodes.find(commit => commit.id === next?.value)

            if (selectedCommit) {
                onSelect(selectedCommit)
            }

            return next
        },
    })

    let debounceTimer: ReturnType<typeof setTimeout>
    const repositoryCommits = createPromiseStore<RepositoryCommits>()

    // Start query initial suggestion
    repositoryCommits.set(getRepositoryCommits($inputValue))

    const debounce = (callback: () => void) => {
        clearTimeout(debounceTimer)
        debounceTimer = setTimeout(callback, 450)
    }

    $: {
        debounce(() => {
            repositoryCommits.set(getRepositoryCommits($inputValue))
        })
    }

    $: filteredCommits = $repositoryCommits.value ? $repositoryCommits.value.ancestors.nodes : []
</script>

<div class="root">
    <Input
        {...$input}
        actions={[input]}
        loading={$repositoryCommits.pending}
        autofocus={true}
        placeholder="Search commits..."
    />

    <div {...$menu} use:menu class="suggestions">
        <ul class="suggestion-list">
            <!-- Initial loading state (but don't show state if data from prev call is presented) -->
            {#if !$repositoryCommits.value && $repositoryCommits.pending}
                <li class="no-data-state">Loading...</li>
            {/if}

            <!-- Error state (show error immediately) -->
            {#if $repositoryCommits.error}
                <li class="no-data-state">
                    <Alert variant="danger">
                        Unable to load commits information: {$repositoryCommits.error.message}
                    </Alert>
                </li>
            {/if}

            {#if !$repositoryCommits.error}
                {#each filteredCommits as commit (commit.id)}
                    <li use:option {...$option(toOption(commit))} class="suggestion-list-item">
                        <span class="title">
                            <Icon svgPath={mdiSourceCommit} inline />
                            <Badge variant="link">{commit.abbreviatedOID}</Badge>
                            <span>{commit.subject}</span>
                        </span>
                        <span class="author">
                            <Avatar avatar={commit.author.person} />
                            <span class="author-name">{commit.author.person.displayName}</span>
                        </span>
                        <span class="timestamp">
                            <Timestamp date={commit.author.date} strict />
                        </span>
                    </li>
                {/each}
            {/if}

            {#if filteredCommits.length === 0 && !$repositoryCommits.pending && !$repositoryCommits.error}
                <li class="no-data-state">
                    No commits matching&nbsp;<b>{$inputValue}</b>, try different search query
                </li>
            {/if}
        </ul>
    </div>

    <footer class="footer">
        <a href={`${repoURL}/-/commits`}> See all commits →</a>
    </footer>
</div>

<style lang="scss">
    .root {
        display: flex;
        // Show the first 8 and half element in the initial suggest block
        // 9th half visible item is needed to indicate that there are more items
        // to pick
        max-height: 24rem;
        flex-direction: column;
    }

    .suggestions {
        flex-grow: 1;
        min-height: 0;
        overflow: auto;
        margin: 0.75rem -0.75rem 0rem -0.75rem;

        // There is no way to turn off styles that come from
        // melt UI popover element, since we render suggestion
        // no in the melt UI popover we turn it off via CSS here.
        position: static !important;
        width: calc(100% + 1.5rem) !important;
    }

    .suggestion-list {
        display: grid;
        grid-template-rows: auto;
        grid-template-columns: [title] auto [author] 10rem [timestamp] 6rem;
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
        padding: 0.325rem;
        cursor: pointer;
        gap: 1rem;
        border-bottom: 1px solid var(--border-color);

        &:last-child {
            border-bottom: none;
        }

        &:hover,
        &[data-highlighted] {
            background: var(--color-bg-3);

            // Branch icon
            :global(svg) {
                color: var(--icon-color);
            }
        }
    }

    .title,
    .author,
    .timestamp {
        display: flex;
        gap: 0.5rem;
        align-items: center;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;

        // Prevent avatar image from shrinking
        :global([data-avatar]) {
            flex-shrink: 0;
        }

        // Timestamp uses tooltip wrapper element with display:contents
        // override this behavior since we have to overflow text within
        // trigger text
        .author-name,
        :global([data-tooltip-root]) {
            display: block;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
    }

    .title {
        padding-left: 0.75rem;

        // Commit icon
        :global(svg) {
            flex-shrink: 0;
            color: var(--icon-muted);
        }

        // Commit oid badge
        :global([data-badge]) {
            font-family: monospace;
        }

        // Commit subject string
        span {
            overflow: hidden;
            text-overflow: ellipsis;
        }
    }

    .timestamp {
        padding-right: 0.75rem;
        color: var(--text-muted);
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
        margin: 0 -0.75rem -0.75rem -0.75rem;
        border-top: 1px solid var(--border-color);

        a {
            padding: 0.75rem;
            width: 100%;
            height: 100%;
            display: flex;
            justify-content: center;
            align-items: center;
            font-weight: 500;

            &:hover {
                background: var(--color-bg-2);
            }
        }
    }
</style>
