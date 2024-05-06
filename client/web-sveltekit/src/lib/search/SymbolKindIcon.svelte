<script lang="ts" context="module">
    import { mdiFileCodeOutline, mdiNull, mdiShape } from '@mdi/js'

    import { SymbolKind } from '$lib/graphql-types'

    function narrowSymbolKind(kind: SymbolKind | string): SymbolKind {
        if (Object.values(SymbolKind).some(k => k === kind)) {
            return kind as SymbolKind
        }
        return SymbolKind.UNKNOWN
    }

    const AbbreviatedSymbolKinds: Map<string, string> = new Map([
        [SymbolKind.ARRAY, '[]'],
        [SymbolKind.BOOLEAN, 'Bo'],
        [SymbolKind.CLASS, 'C'],
        [SymbolKind.CONSTANT, 'Co'],
        [SymbolKind.CONSTRUCTOR, 'Cs'],
        [SymbolKind.ENUM, 'En'],
        [SymbolKind.ENUMMEMBER, 'EM'],
        [SymbolKind.EVENT, 'Ev'],
        [SymbolKind.FIELD, 'Fd'],
        [SymbolKind.FUNCTION, 'Fn'],
        [SymbolKind.INTERFACE, 'In'],
        [SymbolKind.KEY, 'K'],
        [SymbolKind.METHOD, 'M'],
        [SymbolKind.MODULE, 'Mod'],
        [SymbolKind.NAMESPACE, 'NS'],
        [SymbolKind.NUMBER, '#'],
        [SymbolKind.OBJECT, '{}'],
        [SymbolKind.OPERATOR, 'Op'],
        [SymbolKind.PACKAGE, 'Pkg'],
        [SymbolKind.PROPERTY, 'Pr'],
        [SymbolKind.STRING, 'Str'],
        [SymbolKind.STRUCT, 'Sct'],
        [SymbolKind.TYPEPARAMETER, 'TP'],
        [SymbolKind.VARIABLE, 'Var'],
    ])

    const SymbolKindSymbols: Map<string, string> = new Map([
        [SymbolKind.FILE, mdiFileCodeOutline],
        [SymbolKind.NULL, mdiNull],
    ])

    const moduleFamily: Set<string> = new Set([
        SymbolKind.FILE,
        SymbolKind.MODULE,
        SymbolKind.NAMESPACE,
        SymbolKind.PACKAGE,
    ])

    const classFamily: Set<string> = new Set([
        SymbolKind.CLASS,
        SymbolKind.ENUM,
        SymbolKind.INTERFACE,
        SymbolKind.STRUCT,
    ])

    const functionFamily: Set<string> = new Set([SymbolKind.CONSTRUCTOR, SymbolKind.FUNCTION, SymbolKind.METHOD])

    const variableFamily: Set<string> = new Set([
        SymbolKind.VARIABLE,
        SymbolKind.CONSTANT,
        SymbolKind.PROPERTY,
        SymbolKind.EVENT,
        SymbolKind.FIELD,
        SymbolKind.KEY,
        SymbolKind.ENUMMEMBER,
        SymbolKind.TYPEPARAMETER,
    ])
</script>

<script lang="ts">
    import Tooltip from '$lib/Tooltip.svelte'

    import { humanReadableSymbolKind } from './symbolUtils'

    export let symbolKind: SymbolKind | string
</script>

<Tooltip tooltip={humanReadableSymbolKind(symbolKind)}>
    <svg
        aria-label="Symbol kind {symbolKind.toLowerCase()}"
        class:module={moduleFamily.has(symbolKind)}
        class:class={classFamily.has(symbolKind)}
        class:function={functionFamily.has(symbolKind)}
        class:variable={variableFamily.has(symbolKind)}
        viewBox="0 0 24 24"
    >
        <rect x="2" y="2" width="20" height="20" rx="3" />
        {#if AbbreviatedSymbolKinds.has(symbolKind)}
            <text x="50%" y="50%" font-size="75%" dominant-baseline="central" text-anchor="middle" letter-spacing="-0.5"
                >{AbbreviatedSymbolKinds.get(symbolKind)}</text
            >
        {:else}
            <path transform="scale(.66) translate(6, 6)" d={SymbolKindSymbols.get(symbolKind) ?? mdiShape} />
        {/if}
    </svg>
</Tooltip>

<style lang="scss">
    $size: var(--icon-size, 16px);

    svg {
        --color: var(--text-muted);
        &.module {
            --color: var(--oc-teal-8);
        }
        &.class {
            --color: var(--oc-orange-8);
        }
        &.function {
            --color: var(--oc-violet-8);
        }
        &.variable {
            --color: var(--oc-blue-8);
        }

        width: $size;
        height: $size;
        box-sizing: border-box;

        rect {
            stroke: var(--color);
            stroke-width: 2;
            border-radius: 25%;
            fill: none;
        }

        text {
            font-family: var(--code-font-family);
            fill: var(--color);
            font-weight: bold;
        }

        path {
            fill: var(--color);
        }
    }
</style>
