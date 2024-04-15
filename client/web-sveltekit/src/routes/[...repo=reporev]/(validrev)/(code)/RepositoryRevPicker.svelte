<script context="module" lang="ts">
    import type { RepositoryGitRefs_Repository_, RepositoryGitCommits_Repository_ } from './RepositoryRevPicker.gql'

    export type RepositoryBranches = RepositoryGitRefs_Repository_['gitRefs']
    export type RepositoryBranch = RepositoryBranches['nodes'][number]

    export type RepositoryTags = RepositoryGitRefs_Repository_['gitRefs']
    export type RepositoryTag = RepositoryTags['nodes'][number]

    export type RepositoryCommits = NonNullable<RepositoryGitCommits_Repository_['commit']>['ancestors']
    export type RepositoryGitCommit = RepositoryCommits['nodes'][number]
</script>

<script lang="ts">
    import { mdiSourceBranch, mdiTagOutline, mdiSourceCommit } from '@mdi/js'
    import { Button, Badge } from '$lib/wildcard'
    import Popover from '$lib/Popover.svelte'
    import Icon from '$lib/Icon.svelte'
    import Tabs from '$lib/Tabs.svelte'
    import TabPanel from '$lib/TabPanel.svelte'

    import { goto } from '$app/navigation'
    import { replaceRevisionInURL } from '@sourcegraph/shared/src/util/url'

    import type { ResolvedRevision } from '../../+layout'

    import Picker from './Picker.svelte'
    import RepositoryRevPickerItem from './RepositoryRevPickerItem.svelte'

    export let repoURL: string
    export let revision: string | undefined
    export let resolvedRevision: ResolvedRevision

    // Pickers data sources
    export let getRepositoryTags: (query: string) => Promise<RepositoryTags>
    export let getRepositoryCommits: (query: string) => Promise<RepositoryCommits>
    export let getRepositoryBranches: (query: string) => Promise<RepositoryBranches>

    // Show specific short revision if it's presented in the URL
    // otherwise fallback on the default branch name
    $: revisionLabel = revision
        ? revision === resolvedRevision.commitID
            ? resolvedRevision.commitID.slice(0, 7)
            : revision
        : resolvedRevision.defaultBranch ?? ''

    const handleBranchOrTagSelect = (branchOrTag: RepositoryBranch | RepositoryTag): void => {
        goto(replaceRevisionInURL(location.pathname + location.search + location.hash, branchOrTag.abbrevName))
    }

    const handleCommitSelect = (commit: RepositoryGitCommit): void => {
        goto(replaceRevisionInURL(location.pathname + location.search + location.hash, commit.oid))
    }
</script>

<Popover let:registerTrigger let:toggle placement="right-start">
    <Button variant="secondary" size="sm" data-revision-picker-trigger="true">
        <svelte:fragment slot="custom" let:buttonClass>
            <button use:registerTrigger class={buttonClass} on:click={() => toggle()}>
                @{revisionLabel}
            </button>
        </svelte:fragment>
    </Button>
    <div slot="content" class="content" let:toggle>
        <Tabs>
            <TabPanel title="Branches">
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
                    />
                </Picker>
            </TabPanel>
            <TabPanel title="Tags">
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
            <TabPanel title="Commits">
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
    .content {
        padding: 0.75rem;
        min-width: 35rem;
        max-width: 40rem;

        --tabs-gap: 0.25rem;
        --align-tabs: flex-start;

        :global([data-tab-header]) {
            border-bottom: 1px solid var(--border-color-2);
            margin: -0.75rem -0.75rem 0 -0.75rem;
            padding: 0.75rem 0.75rem 0 0.75rem;
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
            grid-template-columns: [title] auto [author] 10rem [timestamp] 6rem;
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

        // Tags picker layout override
        :global([data-tab-panel='Tags']) :global([data-picker-suggestions-list]) {
            grid-template-columns: [title] auto [author] 13rem [timestamp] 7rem;
        }

        // Local override for commits picker abbreviatedOID badge
        :global([data-tab-panel='Commits']) :global([data-badge]) {
            flex-shrink: 0;
        }
    }
</style>
