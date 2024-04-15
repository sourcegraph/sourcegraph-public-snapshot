<script context="module" lang="ts">
    import type { ComboboxOptionProps } from '@melt-ui/svelte'
    import type { RepositoryGitRefs_Repository_ } from './RepositoryRevPicker.gql'

    export type RepositoryTags = RepositoryGitRefs_Repository_['gitRefs']
    export type RepositoryTag = RepositoryTags['nodes'][number]

    const toOption = (tag: RepositoryTag): ComboboxOptionProps<string> => ({
        value: tag.id,
        label: tag.displayName,
    })
</script>

<script lang="ts">
    import { mdiTagOutline } from '@mdi/js'
    import { createCombobox } from '@melt-ui/svelte'

    import Icon from '$lib/Icon.svelte'
    import Avatar from '$lib/Avatar.svelte'
    import Timestamp from '$lib/Timestamp.svelte'
    import { Input, Alert, Badge } from '$lib/wildcard'
    import { createPromiseStore } from '$lib/utils'

    export let repoURL: string
    export let getRepositoryTags: (query: string) => Promise<RepositoryTags>
    export let onSelect: (tag: RepositoryTag) => void

    const {
        elements: { menu, input, option },
        states: { inputValue },
    } = createCombobox<any>({
        portal: null,
        forceVisible: true,
        scrollAlignment: 'nearest',
        closeOnOutsideClick: false,
        onSelectedChange: ({ next }) => {
            const selectedTag = $repositoryTags.value?.nodes.find(tag => tag.id === next?.value)

            if (selectedTag) {
                onSelect(selectedTag)
            }

            return next
        },
    })

    let debounceTimer: ReturnType<typeof setTimeout>
    const repositoryTags = createPromiseStore<RepositoryTags>()

    // Start query initial suggestion
    repositoryTags.set(getRepositoryTags($inputValue))

    const debounce = (callback: () => void) => {
        clearTimeout(debounceTimer)
        debounceTimer = setTimeout(callback, 450)
    }

    $: {
        debounce(() => {
            repositoryTags.set(getRepositoryTags($inputValue))
        })
    }

    $: filteredTags = $repositoryTags.value ? $repositoryTags.value.nodes : []
</script>

<div class="root">
    <Input
        {...$input}
        actions={[input]}
        loading={$repositoryTags.pending}
        autofocus={true}
        placeholder="Find a tag..."
    />

    <div {...$menu} use:menu class="suggestions">
        <ul class="suggestion-list">
            <!-- Initial loading state (but don't show state if data from prev call is presented) -->
            {#if !$repositoryTags.value && $repositoryTags.pending}
                <li class="no-data-state">Loading...</li>
            {/if}

            <!-- Error state (show error immediately) -->
            {#if $repositoryTags.error}
                <li class="no-data-state">
                    <Alert variant="danger">
                        Unable to load tags information: {$repositoryTags.error.message}
                    </Alert>
                </li>
            {/if}

            {#if !$repositoryTags.error}
                {#each filteredTags as tag (tag.id)}
                    <li use:option {...$option(toOption(tag))} class="suggestion-list-item">
                        <span class="title">
                            <Icon svgPath={mdiTagOutline} inline />
                            <Badge variant="link">{tag.displayName}</Badge>
                        </span>
                        <span class="author">
                            {#if tag.target.commit}
                                <Avatar avatar={tag.target.commit?.author.person} />
                                <span class="author-name">{tag.target.commit.author.person.displayName}</span>
                            {/if}
                        </span>
                        <span class="timestamp">
                            {#if tag.target.commit}
                                <Timestamp date={tag.target.commit.author.date} strict />
                            {/if}
                        </span>
                    </li>
                {/each}
            {/if}

            {#if filteredTags.length === 0 && !$repositoryTags.pending && !$repositoryTags.error}
                <li class="no-data-state">
                    No tags matching&nbsp;<b>{$inputValue}</b>, try different search query
                </li>
            {/if}
        </ul>
    </div>

    <footer class="footer">
        <a href={`${repoURL}/-/tags`}>
            See all tags
            {#if !$repositoryTags.error && $repositoryTags.value && $repositoryTags.value?.totalCount !== 0}({$repositoryTags
                    .value.totalCount}){/if}
            →
        </a>
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
        grid-template-columns: [title] auto [author] 13rem [timestamp] 7rem;
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

        // Branch icon
        :global(svg) {
            flex-shrink: 0;
            color: var(--icon-muted);
        }

        // Branch name badge
        :global(span) {
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
