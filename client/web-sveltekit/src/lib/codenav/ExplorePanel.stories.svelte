<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'
    import { writable, readable } from 'svelte/store'

    import { Range } from '@sourcegraph/shared/src/codeintel/scip'

    import type { InfinityQueryStore } from '$lib/graphql'
    import { SymbolUsageKind } from '$lib/graphql-types'
    import { Occurrence } from '$lib/shared'
    import { createEmptySingleSelectTreeState, type SingleSelectTreeState } from '$lib/TreeView'

    import type { ExplorePanel_UsagesResult, ExplorePanel_UsagesVariables } from './ExplorePanel.gql'
    import ExplorePanel, { type ExplorePanelInputs } from './ExplorePanel.svelte'

    export const meta = {
        component: ExplorePanel,
    }
</script>

<script lang="ts">
    const startingTreeState = writable<SingleSelectTreeState>({
        ...createEmptySingleSelectTreeState(),
        disableScope: true,
    })

    const emptyInputs: ExplorePanelInputs = {}
    const simpleInputs: ExplorePanelInputs = {
        activeOccurrence: {
            documentInfo: {
                repoName: 'test/repo',
                filePath: 'test/file',
                commitID: 'deadbeef',
                revision: 'main',
                languages: [],
            },
            occurrence: new Occurrence(Range.fromNumbers(0, 0, 0, 10)),
        },
        usageKindFilter: SymbolUsageKind.REFERENCE,
    }

    const loadingConnection: InfinityQueryStore<ExplorePanel_UsagesResult, ExplorePanel_UsagesVariables> = readable({
        fetching: true,
    }) as InfinityQueryStore<ExplorePanel_UsagesResult, ExplorePanel_UsagesVariables>
</script>

<Story name="Unset">
    <div class="container">
        <ExplorePanel inputs={writable(emptyInputs)} connection={undefined} treeState={startingTreeState} />
    </div>
</Story>

<Story name="Loading">
    <div class="container">
        <ExplorePanel inputs={writable(simpleInputs)} connection={loadingConnection} treeState={startingTreeState} />
    </div>
</Story>

<style lang="scss">
    .container {
        width: 800px;
        height: 200px;
        border: 1px solid black;
        resize: both;
        overflow: hidden;
    }
</style>
