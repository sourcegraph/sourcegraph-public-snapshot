<script lang="ts" context="module">
    import type { ComponentProps } from 'svelte'
    import Icon from '$lib/Icon.svelte'

    import { SymbolKind } from '$lib/graphql-types'
    import Tooltip from '$lib/Tooltip.svelte'

    const icons: Map<string, ComponentProps<Icon>['icon']> = new Map([
        [SymbolKind.ARRAY, ISymbolArray],
        [SymbolKind.BOOLEAN, ISymbolBoolean],
        [SymbolKind.CLASS, ISymbolClass],
        [SymbolKind.CONSTANT, ISymbolConstant],
        [SymbolKind.CONSTRUCTOR, ISymbolConstructor],
        [SymbolKind.ENUM, ISymbolEnum],
        [SymbolKind.ENUMMEMBER, ISymbolEnumMember],
        [SymbolKind.EVENT, ISymbolEvent],
        [SymbolKind.FIELD, ISymbolField],
        [SymbolKind.FILE, ISymbolFile],
        [SymbolKind.FUNCTION, ISymbolFunction],
        [SymbolKind.INTERFACE, ISymbolInterface],
        [SymbolKind.KEY, ISymbolKey],
        [SymbolKind.METHOD, ISymbolMethod],
        [SymbolKind.MODULE, ISymbolModule],
        [SymbolKind.NAMESPACE, ISymbolNamespace],
        [SymbolKind.NULL, ISymbolNull],
        [SymbolKind.NUMBER, ISymbolNumber],
        [SymbolKind.OBJECT, ISymbolObject],
        [SymbolKind.OPERATOR, ISymbolOperator],
        [SymbolKind.PACKAGE, ISymbolPackage],
        [SymbolKind.PROPERTY, ISymbolProperty],
        [SymbolKind.STRING, ISymbolString],
        [SymbolKind.STRUCT, ISymbolStruct],
        [SymbolKind.TYPEPARAMETER, ISymbolTypeParameter],
        [SymbolKind.UNKNOWN, ISymbolUnknown],
        [SymbolKind.VARIABLE, ISymbolVariable],
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
    import { humanReadableSymbolKind } from './symbolUtils'

    export let symbolKind: SymbolKind | string
</script>

<Tooltip tooltip={humanReadableSymbolKind(symbolKind)}>
    <div
        class:module={moduleFamily.has(symbolKind)}
        class:class={classFamily.has(symbolKind)}
        class:function={functionFamily.has(symbolKind)}
        class:variable={variableFamily.has(symbolKind)}
    >
        <Icon icon={icons.get(symbolKind) ?? ISymbolUnknown} aria-label="Symbol kind {symbolKind.toLowerCase()}" />
    </div>
</Tooltip>

<style lang="scss">
    div {
        display: contents;

        // TODO(@taiyab): incorporate these colors into the semantic colors

        --icon-color: currentColor;

        :global(.theme-light) & {
            &.module {
                color: var(--oc-teal-8);
            }
            &.class {
                color: var(--oc-orange-8);
            }
            &.function {
                color: var(--oc-violet-8);
            }
            &.variable {
                color: var(--oc-blue-8);
            }
        }

        :global(.theme-dark) & {
            &.module {
                color: var(--oc-teal-6);
            }
            &.class {
                color: var(--oc-orange-6);
            }
            &.function {
                color: var(--oc-violet-6);
            }
            &.variable {
                color: var(--oc-blue-6);
            }
        }

        :global(svg) {
            height: var(--icon-size, 16px);
            width: var(--icon-size, 16px);
            flex: none;
        }
    }
</style>
