<script lang="ts" context="module">
    interface Symbol {
        kind: SymbolKind
        name: string
        nameHighlights: [number, number][]
        id: string
        children: Symbol[]
    }

    export class SymbolTree implements TreeProvider<Symbol> {
        constructor(private entries: Symbol[]) {}

        public static fromCodeGraphData(): SymbolTree {}

        public static fromStencil(): SymbolTree {}

        public filtered(term: string): SymbolTree {}

        public getEntries(): Symbol[] {
            return this.entries
        }
        public isExpandable(symbol: Symbol): boolean {
            return symbol.children.entries.length > 0
        }
        public isSelectable(): boolean {
            return true
        }
        public fetchChildren(symbol: Symbol): Promise<TreeProvider<Symbol>> {
            return Promise.resolve(new SymbolTree(symbol.children))
        }
        public getNodeID(symbol: Symbol): string {
            return symbol.id
        }
    }
</script>

<script lang="ts">
    import { writable } from 'svelte/store'

    import { highlightRanges } from '$lib/dom'
    import type { SymbolKind } from '$lib/graphql-types'
    import SymbolKindIcon from '$lib/search/SymbolKindIcon.svelte'
    import { createEmptySingleSelectTreeState, type TreeProvider } from '$lib/TreeView'
    import TreeView, { setTreeContext } from '$lib/TreeView.svelte'

    export let symbolTree: SymbolTree

    const treeState = writable({ ...createEmptySingleSelectTreeState(), disableScope: true })
    setTreeContext(treeState)
</script>

<div>
    <TreeView treeProvider={symbolTree}>
        <svelte:fragment let:entry>
            <SymbolKindIcon symbolKind={entry.kind} />
            <span use:highlightRanges={{ ranges: entry.nameHighlights }}>
                {entry.name}
            </span>
        </svelte:fragment>
    </TreeView>
</div>
