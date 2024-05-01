<script lang="ts" context="module">
    import { SymbolKind } from '$lib/graphql-types'

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

    const typeFamily: Set<string> = new Set([
        SymbolKind.STRING,
        SymbolKind.BOOLEAN,
        SymbolKind.NUMBER,
        SymbolKind.ARRAY,
        SymbolKind.OBJECT,
        SymbolKind.NULL,
    ])

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
    import Icon from '$lib/Icon.svelte'
    import { settings } from '$lib/stores'
    import Tooltip from '$lib/Tooltip.svelte'

    import { getSymbolIconPath, humanReadableSymbolKind } from './symbolUtils'

    export let symbolKind: SymbolKind | string

    // Determines whether to show symbol name abbreviations (tags) or icons
    $: useTag = $settings?.experimentalFeatures?.symbolKindTags ?? false
</script>

<Tooltip tooltip={humanReadableSymbolKind(symbolKind)}>
    {#if useTag}
        <span
            aria-label="Symbol kind {symbolKind.toLowerCase()}"
            class="tag"
            class:module={moduleFamily.has(symbolKind)}
            class:class={classFamily.has(symbolKind)}
            class:function={functionFamily.has(symbolKind)}
            class:type={typeFamily.has(symbolKind)}
            class:variable={variableFamily.has(symbolKind)}>{symbolKind[0].toUpperCase()}</span
        >
    {:else}
        <span class="symbol-icon kind-{symbolKind.toLowerCase()}">
            <Icon svgPath={getSymbolIconPath(symbolKind)} inline />
        </span>
    {/if}
</Tooltip>

<style lang="scss">
    // Copied from SymbolIcon.module.scss
    .symbol-icon {
        $oc-level: 5;
        $symbol-kinds: (
            'array': 'red',
            'boolean': 'red',
            'class': 'orange',
            'constant': 'indigo',
            'constructor': 'violet',
            'enum': 'blue',
            'enummember': 'blue',
            'event': 'red',
            'field': 'blue',
            'file': 'gray',
            'function': 'violet',
            'interface': 'green',
            'key': 'yellow',
            'method': 'violet',
            'module': 'grape',
            'namespace': 'grape',
            'null': 'red',
            'number': 'violet',
            'object': 'orange',
            'operator': 'gray',
            'package': 'yellow',
            'property': 'gray',
            'string': 'orange',
            'struct': 'orange',
            'typeparameter': 'blue',
            'variable': 'blue',
        );

        // Default for unknown symbols
        color: var(--oc-gray-#{$oc-level});

        @each $kind, $oc-color in $symbol-kinds {
            &:global(.kind-#{$kind}) {
                color: var(--oc-#{$oc-color}-#{$oc-level});
            }
        }
    }

    .tag {
        display: inline-block;
        font-size: 0.75rem;
        padding: 0.0625rem 0.25rem;
        border-radius: var(--border-radius);
        font-weight: 500;
        color: var(--gray-01);
        font-family: var(--code-font-family);
        // Default background color for "unknown" symbol kinds
        background-color: #343a4d;

        &.module {
            background-color: #237332;
        }

        &.class {
            background-color: #f76707;
        }

        &.function {
            background-color: #0b70db;
        }
        &.type {
            background-color: #a305e1;
        }
        &.variable {
            background-color: #005766;
        }
    }
</style>
