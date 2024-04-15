<script lang="ts">
    import { Button } from '$lib/wildcard'
    import Popover from '$lib/Popover.svelte'
    import Tabs from '$lib/Tabs.svelte'
    import TabPanel from '$lib/TabPanel.svelte'

    import { goto } from '$app/navigation'
    import { replaceRevisionInURL } from '@sourcegraph/shared/src/util/url'

    import type { ResolvedRevision } from '../../+layout'
    import RepositoryBranchesPicker, {
        type RepositoryBranches,
        type RepositoryBranch,
    } from './RepositoryBranchesPicker.svelte'
    import RepositoryCommitsPicker, {
        type RepositoryCommits,
        type RepositoryGitCommit,
    } from './RepositoryCommitsPicker.svelte'
    import RepositoryTagsPicker, { type RepositoryTags, type RepositoryTag } from './RepositoryTagsPicker.svelte'

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
                <RepositoryBranchesPicker
                    {repoURL}
                    {getRepositoryBranches}
                    onSelect={branch => {
                        toggle(false)
                        handleBranchOrTagSelect(branch)
                    }}
                />
            </TabPanel>
            <TabPanel title="Tags">
                <RepositoryTagsPicker
                    {repoURL}
                    {getRepositoryTags}
                    onSelect={tag => {
                        toggle(false)
                        handleBranchOrTagSelect(tag)
                    }}
                />
            </TabPanel>
            <TabPanel title="Commits">
                <RepositoryCommitsPicker
                    {repoURL}
                    {getRepositoryCommits}
                    onSelect={commit => {
                        toggle(false)
                        handleCommitSelect(commit)
                    }}
                />
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
            padding-top: 0.75rem;
        }
    }
</style>
