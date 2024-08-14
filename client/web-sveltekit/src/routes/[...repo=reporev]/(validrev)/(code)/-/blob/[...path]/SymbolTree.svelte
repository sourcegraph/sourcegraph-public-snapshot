<script lang="ts" context="module">
    interface Symbol {
        kind: SymbolKind
        name: string
        nameHighlights: [number, number][]
        id: string
        children: Symbol[]
    }

    function escapeRegExp(s: string): string {
        return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    }

    export class SymbolTree implements TreeProvider<Symbol> {
        public constructor(private entries: Symbol[]) {}

        public static fromCodeGraphData(codeGraphData: CodeGraphData): SymbolTree {}

        // public static fromStencil(): SymbolTree {}

        public filtered(term: string): SymbolTree {
            const pattern = new RegExp(escapeRegExp(term))
            function filterEntries(entries: Symbol[]): Symbol[] {
                const filteredEntries = []
                for (const entry of entries) {
                    const entryNameHighlights = pattern.exec(entry.name)
                    const filteredChildren = filterEntries(entry.children)
                    if (filteredChildren.length > 0 || entryNameHighlights !== null) {
                        filteredEntries.push({
                            ...entry,
                            nameHighlights: entryNameHighlights?.indices ?? [],
                            children: filteredChildren,
                        })
                    }
                }
                return filteredEntries
            }
            return this.constructor(filterEntries(this.entries))
        }

        public count(): number {
            function entryCount(entries: Symbol[]): number {
                return entries.reduce((acc, cur) => acc + 1 + entryCount(cur.children), 0)
            }
            return entryCount(this.entries)
        }

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
    import Icon from '$lib/Icon.svelte'
    import SymbolKindIcon from '$lib/search/SymbolKindIcon.svelte'
    import { createEmptySingleSelectTreeState, type TreeProvider } from '$lib/TreeView'
    import TreeView, { setTreeContext } from '$lib/TreeView.svelte'
    import { Badge } from '$lib/wildcard'

    export let symbolTree: SymbolTree

    const treeState = writable({ ...createEmptySingleSelectTreeState(), disableScope: true })
    setTreeContext(treeState)

    let filterString = ''
    $: filteredSymbolTree = filterString === '' ? symbolTree : symbolTree.filtered(filterString)
</script>

<aside>
    <header>
        <h3><Icon icon={ILucideSymbols} inline aria-hidden />Symbols</h3>
        <Badge variant="secondary">{symbolTree.count()}</Badge>
    </header>
    <input bind:value={filterString} />
    <TreeView treeProvider={filteredSymbolTree}>
        <svelte:fragment let:entry>
            <SymbolKindIcon symbolKind={entry.kind} />
            <span use:highlightRanges={{ ranges: entry.nameHighlights }}>
                {entry.name}
            </span>
        </svelte:fragment>
    </TreeView>
</aside>
