<script lang="ts" context="module">
    export interface ActiveOccurrence {
        documentInfo: DocumentInfo
        occurrence: Occurrence
    }

    export interface ExplorePanelInputs {
        activeOccurrence?: ActiveOccurrence
        usageKindFilter?: SymbolUsageKind
        treeFilter?: TreeFilter
    }

    export interface ExplorePanelContext {
        openReferences(occurrence: ActiveOccurrence): void
        openDefinitions(occurrence: ActiveOccurrence): void
        openImplementations(occurrence: ActiveOccurrence): void
    }

    const exploreContextKey = Symbol('explore context key')
    export function getExplorePanelContext(): ExplorePanelContext {
        return getContext(exploreContextKey) ?? { openReferences: () => {} }
    }
    export function setExplorePanelContext(ctx: ExplorePanelContext) {
        setContext(exploreContextKey, ctx)
    }

    // A set of usages grouped by unique repository, revision, and path
    interface PathGroup {
        repository: string
        revision: string
        path: string
        usages: ExplorePanel_Usage[]
    }

    // Groups all usages into consecutive groups of matching repo/rev/path.
    // Maintains input order so paging in new results doesn't cause weirdness.
    //
    // NOTE: this expects that usages are already ordered as contiguous
    // blocks for the same repository and the same file, which is a guarantee
    // provided by the usages API.
    function groupUsages(usages: ExplorePanel_Usage[]): PathGroup[] {
        const groups: PathGroup[] = []
        let current: PathGroup | undefined = undefined

        for (const usage of usages) {
            const { repository, revision, path } = usage.usageRange
            if (
                current &&
                current.repository === repository &&
                current.revision === revision &&
                current.path === path
            ) {
                current.usages.push(usage)
            } else {
                if (current) {
                    groups.push(current)
                }
                current = {
                    repository,
                    revision,
                    path,
                    usages: [usage],
                }
            }
        }
        if (current) {
            groups.push(current)
        }

        return groups
    }

    interface RepoTreeEntry {
        type: 'repo'
        name: string
    }

    interface DirTreeEntry {
        type: 'dir'
        repo: string
        path: string // The full path
        name: string // The path element for this dir
    }

    interface FileTreeEntry {
        type: 'file'
        repo: string
        path: string // The full path
        name: string // The file name
    }

    const typeRanks = { repo: 0, dir: 1, file: 2 }

    type TreeEntry = RepoTreeEntry | FileTreeEntry | DirTreeEntry

    function generateTree(pathGroups: PathGroup[]): TreeProvider<TreeEntry> {
        type Tree = Map<string, { entry: TreeEntry; tree: Tree }>
        const tree: Tree = new Map()
        const addToTree = (repo: string, path: string) => {
            if (!tree.get(repo)) {
                tree.set(repo, { entry: { type: 'repo', name: repo }, tree: new Map() })
            }
            const repoEntry = tree.get(repo)!

            const pathElements = path.split('/')
            const dirs = pathElements.slice(0, -1)

            let current = repoEntry
            for (const [index, dir] of dirs.entries()) {
                if (!current.tree.get(dir)) {
                    current.tree.set(dir, {
                        tree: new Map(),
                        entry: {
                            type: 'dir',
                            repo,
                            name: dir,
                            path: pathElements.slice(0, index + 1).join('/') + '/',
                        },
                    })
                }
                current = current.tree.get(dir)!
            }

            const fileName = pathElements.at(-1)! // splitting will always have at least one element
            current.tree.set(fileName, {
                tree: new Map(),
                entry: {
                    type: 'file',
                    repo,
                    path,
                    name: fileName,
                },
            })
        }

        for (const pathGroup of pathGroups) {
            addToTree(pathGroup.repository, pathGroup.path)
        }

        function newTreeProvider(tree: Tree): TreeProvider<TreeEntry> {
            return {
                getNodeID(entry) {
                    if (entry.type === 'repo') {
                        return `repo-${entry.name}`
                    } else {
                        return `path-${entry.repo}-${entry.path}`
                    }
                },
                getEntries(): TreeEntry[] {
                    return Array.from(tree.entries())
                        .map(([_name, entry]) => entry.entry)
                        .toSorted((a, b) => {
                            // Sort directories first, then sort alphabetically
                            if (a.type !== b.type) {
                                return typeRanks[a.type] - typeRanks[b.type]
                            }
                            return a.name.localeCompare(b.name)
                        })
                },
                isExpandable(entry) {
                    return entry.type === 'repo' || entry.type === 'dir'
                },
                isSelectable() {
                    return true
                },
                fetchChildren(entry) {
                    if (entry.type === 'repo' || entry.type === 'dir') {
                        return Promise.resolve(newTreeProvider(tree.get(entry.name)!.tree))
                    } else {
                        throw new Error('path nodes are not expandable')
                    }
                },
            }
        }

        return newTreeProvider(tree)
    }

    export function getUsagesStore(client: GraphQLClient, documentInfo: DocumentInfo, occurrence: Occurrence) {
        const baseVariables: ExplorePanel_UsagesVariables = {
            repoName: documentInfo.repoName,
            revspec: documentInfo.commitID,
            filePath: documentInfo.filePath,
            rangeStart: occurrence.range.start,
            rangeEnd: occurrence.range.end,
            symbolComparator:
                occurrence.symbol && occurrence.symbolProvenance
                    ? {
                          name: { equals: occurrence.symbol },
                          provenance: { equals: occurrence.symbolProvenance },
                      }
                    : null,
            first: 100,
            afterCursor: null,
        }

        return infinityQuery({
            client,
            query: ExplorePanel_Usages,
            variables: baseVariables,
            map: result => ({
                data: result?.data?.usagesForSymbol?.nodes,
                error: result?.error,
                nextVariables: result?.data?.usagesForSymbol?.pageInfo?.hasNextPage
                    ? {
                          ...baseVariables,
                          afterCursor: result?.data?.usagesForSymbol?.pageInfo?.endCursor,
                      }
                    : undefined,
            }),
            merge: (previous, next) => (previous ?? []).concat(next ?? []),
        })
    }

    function matchesUsageKind(usageKindFilter: SymbolUsageKind | undefined): (usage: ExplorePanel_Usage) => boolean {
        return usage => usageKindFilter === undefined || usage.usageKind === usageKindFilter
    }

    interface TreeFilter {
        repository: string
        path?: string
    }

    function matchesTreeFilter(treeFilter: TreeFilter | undefined): (pathGroup: PathGroup) => boolean {
        return pathGroup =>
            treeFilter === undefined ||
            (treeFilter.repository === pathGroup.repository &&
                (treeFilter.path === undefined || pathGroup.path.startsWith(treeFilter.path)))
    }

    export function entryIDForFilter(filter: TreeFilter): string {
        if (filter.path) {
            return `path-${filter.repository}-${filter.path}`
        }
        return `repo-${filter.repository}`
    }
</script>

<script lang="ts">
    import { getContext, setContext } from 'svelte'
    import { type Writable } from 'svelte/store'

    import { infinityQuery, type GraphQLClient, type InfinityQueryStore } from '$lib/graphql'
    import { SymbolUsageKind } from '$lib/graphql-types'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import Scroller from '$lib/Scroller.svelte'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import LoadingSkeleton from '$lib/search/dynamicFilters/LoadingSkeleton.svelte'
    import { displayRepoName, Occurrence } from '$lib/shared'
    import { type SingleSelectTreeState, type TreeProvider } from '$lib/TreeView'
    import TreeView, { setTreeContext } from '$lib/TreeView.svelte'
    import type { DocumentInfo } from '$lib/web'
    import { Alert, PanelGroup } from '$lib/wildcard'
    import Panel from '$lib/wildcard/resizable-panel/Panel.svelte'
    import PanelResizeHandle from '$lib/wildcard/resizable-panel/PanelResizeHandle.svelte'

    import type { ExplorePanel_Usage, ExplorePanel_UsagesVariables } from './ExplorePanel.gql'
    import { ExplorePanel_Usages } from './ExplorePanel.gql'
    import ExplorePanelFileUsages from './ExplorePanelFileUsages.svelte'

    export let inputs: Writable<ExplorePanelInputs>
    export let connection: InfinityQueryStore<ExplorePanel_Usage[], ExplorePanel_UsagesVariables> | undefined
    export let treeState: Writable<SingleSelectTreeState>

    $: setTreeContext(treeState)

    // TODO: it would be really nice if the tree API emitted select events with tree elements, not HTML elements
    function handleSelect(target: HTMLElement) {
        const selected = target.querySelector('[data-repo-name]') as HTMLElement
        const repository = selected.dataset.repoName ?? ''
        const path = selected.dataset.path
        inputs.update(old => {
            const deselect = old.treeFilter && old.treeFilter.repository === repository && old.treeFilter.path === path
            return {
                ...old,
                treeFilter: deselect ? undefined : { repository, path },
            }
        })
    }

    $: loading = $connection?.fetching
    $: usages = $connection?.data ?? []
    $: kindFilteredUsages = usages.filter(matchesUsageKind($inputs.usageKindFilter))
    $: pathGroups = groupUsages(kindFilteredUsages)
    $: outlineTree = generateTree(pathGroups)
    $: displayGroups = pathGroups.filter(matchesTreeFilter($inputs.treeFilter))

    let referencesScroller: HTMLElement | undefined
</script>

{#if $inputs.activeOccurrence === undefined}
    <div class="no-selection">
        <Icon icon={ISgSymbols} />
        <p>Select a symbol in the code panel to view references</p>
    </div>
{:else}
    <PanelGroup id="references">
        <Panel id="references-sidebar" defaultSize={25} minSize={20} maxSize={60}>
            <div class="sidebar">
                <fieldset>
                    <legend hidden>Select usage kind</legend>
                    {#each Object.values(SymbolUsageKind) as usageKind}
                        {@const checked = usageKind === $inputs.usageKindFilter}
                        {@const id = `usage-kind-${usageKind}`}
                        <input
                            type="radio"
                            bind:group={$inputs.usageKindFilter}
                            name="usageKind"
                            value={usageKind}
                            {id}
                            {checked}
                        />
                        <label for={id}>{usageKind.toLowerCase()}s</label>
                    {/each}
                </fieldset>
                <div class="outline">
                    {#if pathGroups.length > 0}
                        <h4>Filter by location</h4>
                        <TreeView treeProvider={outlineTree} on:select={event => handleSelect(event.detail)}>
                            <svelte:fragment let:entry>
                                {#if entry.type === 'repo'}
                                    <span class="repo-entry" data-repo-name={entry.name}>
                                        <CodeHostIcon repository={entry.name} />
                                        {displayRepoName(entry.name)}
                                    </span>
                                {:else}
                                    <span class="path-entry" data-repo-name={entry.repo} data-path={entry.path}>
                                        {entry.name}
                                    </span>
                                {/if}
                            </svelte:fragment>
                            <Alert slot="error" let:error variant="danger">
                                {error.message}
                            </Alert>
                        </TreeView>
                    {/if}
                </div>
            </div>
        </Panel>
        <PanelResizeHandle />
        <Panel id="references-content">
            {#if loading}
                <div class="loading">
                    <LoadingSkeleton />
                    <LoadingSkeleton />
                    <LoadingSkeleton />
                </div>
            {:else}
                <Scroller bind:viewport={referencesScroller} margin={600} on:more={() => connection?.fetchMore()}>
                    {#if displayGroups.length > 0}
                        <ul>
                            {#each displayGroups as pathGroup}
                                <li>
                                    <ExplorePanelFileUsages scrollContainer={referencesScroller} {...pathGroup} />
                                </li>
                            {/each}
                        </ul>
                        {#if loading}
                            <LoadingSpinner center />
                        {/if}
                    {:else}
                        <div class="no-results">No results.</div>
                    {/if}
                </Scroller>
            {/if}
        </Panel>
    </PanelGroup>
{/if}

<style lang="scss">
    .sidebar {
        height: 100%;
        overflow-y: auto;
    }

    fieldset {
        display: flex;
        flex-direction: column;
        border-bottom: 1px solid var(--border-color);
        padding: 0.25rem 0;

        input {
            appearance: none;
        }

        label {
            text-transform: capitalize;
            cursor: pointer;
            padding: 0.375rem 0.75rem;
            background-color: transparent;
        }

        input:checked + label {
            background-color: var(--primary);
            color: var(--light-text);
        }

        input:not(:checked) + label:hover {
            background-color: var(--secondary-4);
        }
    }

    .outline {
        h4 {
            padding: 0.5rem 0.75rem;
            margin: 0;
        }
        padding: 0rem;

        :global([data-treeitem]) > :global([data-treeitem-label]) {
            cursor: pointer;

            &:hover {
                background-color: var(--secondary-4);
            }

            word-break: break-all;
        }

        :global([data-treeitem][aria-selected='true']) > :global([data-treeitem-label]) {
            --tree-node-expand-icon-color: var(--body-bg);
            --file-icon-color: var(--body-bg);
            --tree-node-label-color: var(--body-bg);

            background-color: var(--primary);
            &:hover {
                background-color: var(--primary);
            }
        }
    }

    .repo-entry {
        display: flex;
        align-items: center;
        gap: 0.375em;
        font-size: var(--font-size-base);
    }
    .path-entry {
        font-size: var(--font-size-small);
    }

    ul {
        all: unset;
        li {
            all: unset;
        }
    }

    .no-selection {
        :global([data-icon]) {
            flex: none;
        }
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 1rem;
        padding: 1rem;
        text-align: center;

        height: 100%;
        width: 100%;

        --icon-color: currentColor;

        color: var(--text-muted);
        font-weight: 500;
    }

    .no-results {
        display: flex;
        align-items: center;
        justify-content: center;
        height: 100%;
        width: 100%;
        color: var(--text-muted);
        font-weight: 500;
    }

    .loading {
        display: flex;
        flex-direction: column;
        align-items: stretch;
        width: 100%;
        height: 100%;
        padding: 1rem;
    }
</style>
