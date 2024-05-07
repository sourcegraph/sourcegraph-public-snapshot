<script lang="ts" context="module">
    import { SvelteComponent } from 'svelte'

    import { SymbolKind } from '$lib/graphql-types'
    import Array from '$lib/icons/symbols/Array.svelte'
    import Boolean from '$lib/icons/symbols/Boolean.svelte'
    import Class from '$lib/icons/symbols/Class.svelte'
    import Constant from '$lib/icons/symbols/Constant.svelte'
    import Constructor from '$lib/icons/symbols/Constructor.svelte'
    import Enum from '$lib/icons/symbols/Enum.svelte'
    import EnumMember from '$lib/icons/symbols/EnumMember.svelte'
    import Event from '$lib/icons/symbols/Event.svelte'
    import Field from '$lib/icons/symbols/Field.svelte'
    import File from '$lib/icons/symbols/File.svelte'
    import Function from '$lib/icons/symbols/Function.svelte'
    import Interface from '$lib/icons/symbols/Interface.svelte'
    import Key from '$lib/icons/symbols/Key.svelte'
    import Method from '$lib/icons/symbols/Method.svelte'
    import Module from '$lib/icons/symbols/Module.svelte'
    import Namespace from '$lib/icons/symbols/Namespace.svelte'
    import Null from '$lib/icons/symbols/Null.svelte'
    import Number from '$lib/icons/symbols/Number.svelte'
    import Object from '$lib/icons/symbols/Object.svelte'
    import Operator from '$lib/icons/symbols/Operator.svelte'
    import Package from '$lib/icons/symbols/Package.svelte'
    import Property from '$lib/icons/symbols/Property.svelte'
    import String from '$lib/icons/symbols/String.svelte'
    import Struct from '$lib/icons/symbols/Struct.svelte'
    import TypeParameter from '$lib/icons/symbols/TypeParameter.svelte'
    import Unknown from '$lib/icons/symbols/Unknown.svelte'
    import Variable from '$lib/icons/symbols/Variable.svelte'
    import Tooltip from '$lib/Tooltip.svelte'

    const icons: Map<string, typeof SvelteComponent<{}>> = new Map([
        [SymbolKind.ARRAY, Array],
        [SymbolKind.BOOLEAN, Boolean],
        [SymbolKind.CLASS, Class],
        [SymbolKind.CONSTANT, Constant],
        [SymbolKind.CONSTRUCTOR, Constructor],
        [SymbolKind.ENUM, Enum],
        [SymbolKind.ENUMMEMBER, EnumMember],
        [SymbolKind.EVENT, Event],
        [SymbolKind.FIELD, Field],
        [SymbolKind.FILE, File],
        [SymbolKind.FUNCTION, Function],
        [SymbolKind.INTERFACE, Interface],
        [SymbolKind.KEY, Key],
        [SymbolKind.METHOD, Method],
        [SymbolKind.MODULE, Module],
        [SymbolKind.NAMESPACE, Namespace],
        [SymbolKind.NULL, Null],
        [SymbolKind.NUMBER, Number],
        [SymbolKind.OBJECT, Object],
        [SymbolKind.OPERATOR, Operator],
        [SymbolKind.PACKAGE, Package],
        [SymbolKind.PROPERTY, Property],
        [SymbolKind.STRING, String],
        [SymbolKind.STRUCT, Struct],
        [SymbolKind.TYPEPARAMETER, TypeParameter],
        [SymbolKind.UNKNOWN, Unknown],
        [SymbolKind.VARIABLE, Variable],
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
        aria-label="Symbol kind {symbolKind.toLowerCase()}"
        class:module={moduleFamily.has(symbolKind)}
        class:class={classFamily.has(symbolKind)}
        class:function={functionFamily.has(symbolKind)}
        class:variable={variableFamily.has(symbolKind)}
    >
        <svelte:component this={icons.get(symbolKind) ?? Unknown} />
    </div>
</Tooltip>

<style lang="scss">
    div {
        display: contents;

        // TODO(@taiyab): incorporate these colors into the semantic colors

        :global(.theme-light) & {
            --icon-color: var(--text-muted);
            &.module {
                --icon-color: var(--oc-teal-8);
            }
            &.class {
                --icon-color: var(--oc-orange-8);
            }
            &.function {
                --icon-color: var(--oc-violet-8);
            }
            &.variable {
                --icon-color: var(--oc-blue-8);
            }
        }

        :global(.theme-dark) & {
            --icon-color: var(--text-muted);
            &.module {
                --icon-color: var(--oc-teal-6);
            }
            &.class {
                --icon-color: var(--oc-orange-6);
            }
            &.function {
                --icon-color: var(--oc-violet-6);
            }
            &.variable {
                --icon-color: var(--oc-blue-6);
            }
        }
    }
</style>
