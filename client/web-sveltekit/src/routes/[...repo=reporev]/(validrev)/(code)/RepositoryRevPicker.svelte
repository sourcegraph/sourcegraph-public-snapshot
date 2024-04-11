<script lang="ts">
    import { Button } from '$lib/wildcard'
    import Popover from '$lib/Popover.svelte'
    import Tabs from '$lib/Tabs.svelte'
    import TabPanel from '$lib/TabPanel.svelte'

    import type { ResolvedRevision } from '../../+layout'
    import RepositoryBranchesPicker, { type RepositoryBranches } from './RepositoryBranchesPicker.svelte'
    import type { RepositoryGitCommits_Repository_, RepositoryGitRefs_Repository_ } from './RepositoryRevPicker.gql'

    export let revision: string | undefined
    export let resolvedRevision: ResolvedRevision

    // Pickers data sources
    export let getRepositoryTags: (query: string) => Promise<RepositoryGitRefs_Repository_['gitRefs']>
    export let getRepositoryCommits: (query: string) => Promise<RepositoryGitCommits_Repository_['commit']>
    export let getRepositoryBranches: (query: string) => Promise<RepositoryBranches>

    // Show specific short revision if it's presented in the URL
    // otherwise fallback on the default branch name
    $: revisionLabel = revision
        ? revision === resolvedRevision.commitID
            ? resolvedRevision.commitID.slice(0, 7)
            : revision
        : resolvedRevision.defaultBranch ?? ''
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
                <div class="tab-content">
                    <RepositoryBranchesPicker
                        {getRepositoryBranches}
                        onSelect={() => {
                            console.log('CLOOOSE')
                            toggle(false)
                        }}
                    />
                </div>
            </TabPanel>
            <TabPanel title="Tags">
                <div class="tab-content">Tags</div>
            </TabPanel>
            <TabPanel title="Commits">
                <div class="tab-content">Commits</div>
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
    }

    .tab-content {
        padding-top: 0.5rem;
    }
</style>
