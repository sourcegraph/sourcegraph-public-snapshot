<script context="module" lang="ts">
    import type { Keys } from '$lib/Hotkey'

    import type { RepositoryGitRefs, RevPickerGitCommit } from './RepositoryRevPicker.gql'

    export type RepositoryBranches = RepositoryGitRefs['gitRefs']
    export type RepositoryBranch = RepositoryBranches['nodes'][number]

    export type RepositoryTags = RepositoryGitRefs['gitRefs']
    export type RepositoryTag = RepositoryTags['nodes'][number]

    export type RepositoryCommits = { nodes: RevPickerGitCommit[] }
    export type RepositoryGitCommit = RevPickerGitCommit

    const branchesHotkey: Keys = {
        key: 'shift+b',
    }

    const tagsHotkey: Keys = {
        key: 'shift+t',
    }

    const commitsHotkey: Keys = {
        key: 'shift+c',
    }
</script>

<script lang="ts">
    import { goto } from '$app/navigation'
    import Icon from '$lib/Icon.svelte'
    import Popover from '$lib/Popover.svelte'
    import { replaceRevisionInURL } from '$lib/shared'
    import TabPanel from '$lib/TabPanel.svelte'
    import Tabs from '$lib/Tabs.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Badge } from '$lib/wildcard'
    import { getButtonClassName } from '$lib/wildcard/Button'
    import ButtonGroup from '$lib/wildcard/ButtonGroup.svelte'
    import CopyButton from '$lib/wildcard/CopyButton.svelte'

    import type { ResolvedRevision } from '../+layout'

    import Picker from './Picker.svelte'
    import RepositoryRevPickerItem from './RepositoryRevPickerItem.svelte'

    export let repoURL: string
    export let revision: string | undefined
    export let resolvedRevision: ResolvedRevision

    // Pickers data sources
    export let getRepositoryTags: (query: string) => PromiseLike<RepositoryTags>
    export let getRepositoryCommits: (query: string) => PromiseLike<RepositoryCommits>
    export let getRepositoryBranches: (query: string) => PromiseLike<RepositoryBranches>

    // Show specific short revision if it's presented in the URL
    // otherwise fallback on the default branch name
    $: revisionLabel = revision
        ? revision === resolvedRevision.commitID
            ? resolvedRevision.commitID.slice(0, 7)
            : revision
        : resolvedRevision.defaultBranch ?? ''

    $: isOnSpecificRev = revisionLabel !== resolvedRevision.defaultBranch

    const handleGoToDefaultBranch = (defaultBranch: string): void => {
        goto(replaceRevisionInURL(location.pathname + location.search + location.hash, defaultBranch))
    }

    const handleBranchOrTagSelect = (branchOrTag: RepositoryBranch | RepositoryTag): void => {
        goto(replaceRevisionInURL(location.pathname + location.search + location.hash, branchOrTag.abbrevName))
    }

    const handleCommitSelect = (commit: RepositoryGitCommit): void => {
        goto(replaceRevisionInURL(location.pathname + location.search + location.hash, commit.oid))
    }

    const buttonClass = getButtonClassName({ variant: 'secondary', outline: true, size: 'sm' })
</script>

<Popover let:registerTrigger let:registerTarget let:toggle placement="right-start">
    <div use:registerTarget data-repo-rev-picker-trigger>
        <ButtonGroup>
            <button use:registerTrigger class="{buttonClass} rev-name" on:click={() => toggle()}>
                @{revisionLabel}
            </button>

            <CopyButton value={revisionLabel}>
                <button let:handleCopy on:click={() => handleCopy()} class="{buttonClass} copy-button hoverable-button">
                    <Icon icon={ILucideCopy} aria-hidden="true" --icon-size="1em" />
                </button>
            </CopyButton>

            {#if isOnSpecificRev}
                <Tooltip tooltip="Go to default branch">
                    <button
                        class="{buttonClass} hoverable-button"
                        on:click={() => handleGoToDefaultBranch(resolvedRevision.defaultBranch)}
                    >
                        <Icon icon={ILucideX} aria-hidden="true" --icon-size="1em" />
                    </button>
                </Tooltip>
            {/if}
        </ButtonGroup>
    </div>

    <div slot="content" class="content" let:toggle>
        <Tabs>
            <TabPanel title="Branches" shortcut={branchesHotkey}>
                <Picker
                    name="branches"
                    seeAllItemsURL={`${repoURL}/-/branches`}
                    getData={getRepositoryBranches}
                    toOption={branch => ({ value: branch.id, label: branch.displayName })}
                    onSelect={branch => {
                        toggle(false)
                        handleBranchOrTagSelect(branch)
                    }}
                    let:value
                >
                    <RepositoryRevPickerItem
                        icon={ILucideGitBranch}
                        label={value.displayName}
                        author={value.target.commit?.author}
                    >
                        <svelte:fragment slot="title">
                            <Icon icon={ILucideGitBranch} inline aria-hidden="true" />
                            <Badge variant="link">{value.displayName}</Badge>
                            {#if value.displayName === resolvedRevision.defaultBranch}
                                <Badge variant="secondary" small>DEFAULT</Badge>
                            {/if}
                        </svelte:fragment>
                    </RepositoryRevPickerItem>
                </Picker>
            </TabPanel>
            <TabPanel title="Tags" shortcut={tagsHotkey}>
                <Picker
                    name="tags"
                    seeAllItemsURL={`${repoURL}/-/tags`}
                    getData={getRepositoryTags}
                    toOption={tag => ({ value: tag.id, label: tag.displayName })}
                    onSelect={tag => {
                        toggle(false)
                        handleBranchOrTagSelect(tag)
                    }}
                    let:value
                >
                    <RepositoryRevPickerItem
                        icon={ILucideTag}
                        label={value.displayName}
                        author={value.target.commit?.author}
                    />
                </Picker>
            </TabPanel>
            <TabPanel title="Commits" shortcut={commitsHotkey}>
                <Picker
                    name="commits"
                    seeAllItemsURL={`${repoURL}/-/commits`}
                    getData={getRepositoryCommits}
                    toOption={commit => ({ value: commit.id, label: commit.oid })}
                    onSelect={commit => {
                        toggle(false)
                        handleCommitSelect(commit)
                    }}
                    let:value
                >
                    <RepositoryRevPickerItem label="" author={value.author}>
                        <svelte:fragment slot="title">
                            <Icon icon={ILucideGitCommitVertical} inline aria-hidden="true" />
                            <Badge variant="link">{value.abbreviatedOID}</Badge>
                            <span class="commit-subject">{value.subject}</span>
                        </svelte:fragment>
                    </RepositoryRevPickerItem>
                </Picker>
            </TabPanel>
        </Tabs>
    </div>
</Popover>

<style lang="scss">
    [data-repo-rev-picker-trigger] > :global(*) {
        width: 100%;
        height: 100%;
    }

    .rev-name {
        overflow-wrap: anywhere;
        border-right: none;
        text-align: left;
    }

    .copy-button {
        flex: 0;
        border-left: none;
    }

    .hoverable-button {
        color: var(--text-muted);
        &:hover {
            color: var(--body-color);
        }
    }

    .content {
        min-width: 35rem;
        max-width: 40rem;
        width: 640px;

        --align-tabs: flex-start;

        :global([data-tab-header]) {
            padding: 0 0.5rem;
        }

        // Pickers style
        :global([data-picker-root]) {
            // Show the first 8 and half element in the initial suggest block
            // 9th half visible item is needed to indicate that there are more items
            // to pick
            max-height: 25.5rem;
        }

        :global([data-picker-suggestions-list]) {
            display: grid;
            grid-template-rows: auto;
            grid-template-columns: [title] auto [author] minmax(0, 10rem) [timestamp] minmax(0, 8rem);
        }

        :global([data-picker-suggestions-list-item]) {
            display: grid;
            grid-column: 1 / 4;
            grid-template-columns: subgrid;
            gap: 1rem;
        }

        .commit-subject {
            overflow: hidden;
            text-overflow: ellipsis;
        }

        // Local override for commits picker abbreviatedOID badge
        :global([data-tab-panel='Commits']) :global([data-badge]) {
            flex-shrink: 0;
        }
    }
</style>
