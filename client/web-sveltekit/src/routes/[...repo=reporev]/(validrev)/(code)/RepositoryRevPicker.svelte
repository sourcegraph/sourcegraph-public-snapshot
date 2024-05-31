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
    import { mdiClose, mdiSourceBranch, mdiTagOutline, mdiSourceCommit } from '@mdi/js'

    import { goto } from '$app/navigation'
    import Icon from '$lib/Icon.svelte'
    import Popover from '$lib/Popover.svelte'
    import { replaceRevisionInURL } from '$lib/shared'
    import TabPanel from '$lib/TabPanel.svelte'
    import Tabs from '$lib/Tabs.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Button, Badge } from '$lib/wildcard'

    import type { ResolvedRevision } from '../../+layout'

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
</script>

<Popover let:registerTrigger let:registerTarget let:toggle placement="right-start">
    <div
        use:registerTarget
        class="button-group"
        class:is-on-specific-rev={isOnSpecificRev}
        data-repo-rev-picker-trigger
    >
        <Button variant="secondary" size="sm">
            <svelte:fragment slot="custom" let:buttonClass>
                <button use:registerTrigger class={`${buttonClass} revision-trigger`} on:click={() => toggle()}>
                    @{revisionLabel}
                </button>
            </svelte:fragment>
        </Button>

        {#if isOnSpecificRev}
            <span class="reset-button-container">
                <Tooltip tooltip="Go to default branch">
                    <Button
                        size="sm"
                        variant="secondary"
                        on:click={() => handleGoToDefaultBranch(resolvedRevision.defaultBranch)}
                    >
                        <Icon svgPath={mdiClose} --icon-size="16px" />
                    </Button>
                </Tooltip>
            </span>
        {/if}
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
                        iconPath={mdiSourceBranch}
                        label={value.displayName}
                        author={value.target.commit?.author}
                    >
                        <svelte:fragment slot="title">
                            <Icon svgPath={mdiSourceBranch} inline />
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
                        iconPath={mdiTagOutline}
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
                    <RepositoryRevPickerItem label="" iconPath="" author={value.author}>
                        <svelte:fragment slot="title">
                            <Icon svgPath={mdiSourceCommit} inline />
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
    .button-group {
        display: flex;
        min-width: 0;

        .reset-button-container {
            display: contents;

            // Get access to the reset branch button through container class
            :global(button) {
                border-top-left-radius: 0;
                border-bottom-left-radius: 0;
            }

            :global([data-icon]) {
                display: flex;
                align-items: center;
                justify-content: center;
                --icon-size: 16px;
            }
        }

        .revision-trigger {
            white-space: nowrap;
            text-overflow: ellipsis;
            overflow: hidden;
            flex-grow: 1;
            text-align: left;
        }

        &.is-on-specific-rev .revision-trigger {
            border-top-right-radius: 0;
            border-bottom-right-radius: 0;
            border-right: none;
        }
    }

    .content {
        padding: 0.75rem;
        min-width: 35rem;
        max-width: 40rem;
        width: 640px;

        --align-tabs: flex-start;

        :global([data-tab-header]) {
            border-bottom: 1px solid var(--border-color-2);
            margin: -0.75rem -0.75rem 0 -0.75rem;
            padding: 0 0.5rem;
        }

        :global([data-tab]) {
            border-bottom-left-radius: 0;
            border-bottom-right-radius: 0;
        }

        :global([data-tab-panel]) {
            margin: 0 -0.75rem -0.75rem -0.75rem;
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
            grid-template-columns: [title] auto [author] minmax(0, 10rem) [timestamp] minmax(0, 6rem);
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
