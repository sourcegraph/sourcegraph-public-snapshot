<script lang="ts">
    import { Badge, Button } from '$lib/wildcard'
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
    import Picker from './Picker.svelte'
    import PickerEntry from './PickerEntry.svelte'
    import { mdiSourceBranch, mdiSourceCommit, mdiTagOutline } from '@mdi/js'
    import Icon from '$lib/Icon.svelte'

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
                <Picker getData={getRepositoryBranches} onSelect={branch => {
                    toggle(false)
                    handleBranchOrTagSelect(branch)
                }} toOption={branch => ({value: branch.id, label: branch.displayName})} let:value>
                    <PickerEntry iconPath={mdiSourceBranch} label={value.displayName} commit={value.target.commit} />
                </Picker>
            </TabPanel>
            <TabPanel title="Tags">
                <Picker getData={getRepositoryTags} onSelect={tag => {
                    toggle(false)
                    handleBranchOrTagSelect(tag)
                }} toOption={tag => ({value: tag.id, label: tag.displayName})} let:value>
                    <PickerEntry iconPath={mdiTagOutline} label={value.displayName} commit={value.target.commit} />
                </Picker>
            </TabPanel>
            <TabPanel title="Commits">
                <Picker getData={input => getRepositoryCommits(input).then(result => ({nodes: result?.ancestors.nodes ?? [], totalCount: 0}))} onSelect={commit => {
                    toggle(false)
                    handleCommitSelect(commit)
                }} toOption={commit => ({value: commit.id, label: commit.abbreviatedOID})} let:value>
                    <PickerEntry iconPath={mdiTagOutline} label="" commit={value}>
                        <svelte:fragment slot="title">
                            <Icon svgPath={mdiSourceCommit} inline />
                            <Badge variant="link">{value.abbreviatedOID}</Badge>
                            <span>{value.subject}</span>
                        </svelte:fragment>
                    </PickerEntry>
                </Picker>
            </TabPanel>
        </Tabs>
    </div>
</Popover>

<style lang="scss">
    .content {
        padding: 0.5rem;
        min-width: 35rem;
        max-width: 40rem;

        --tabs-gap: 0.25rem;
        --align-tabs: flex-start;

        :global([data-tab-panel]) {
            padding-top: 0.5rem;
        }


        // Commit oid badge
        :global([data-badge]) {
            font-family: monospace;
        }

    }
</style>
