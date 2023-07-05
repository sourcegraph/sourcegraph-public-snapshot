<script lang="ts" context="module">
    import type { Action } from "svelte/action"
    import { derived, writable } from "svelte/store"

    function createDividersStore() {
        const {subscribe, set, update} = writable<Record<string, number>>(JSON.parse(localStorage.getItem('dividers') ?? '{}'))

        return {
            subscribe,
            updateDivider: (name: string, value: number) => update(dividers => {
                const newValue = {...dividers, [name]: value}
                localStorage.setItem('dividers', JSON.stringify(newValue))
                return newValue
            }),
        }
    }

    export function getDividerStore(name: string, defaultValue: number) {
        return derived(dividerStore, dividers => dividers[name] ?? defaultValue)
    }


    const dividerStore = createDividersStore()

    export const divider: Action<HTMLElement, {id: string, defaultValue: number, compute: (ratio: number) => string|{min: string, max: string}}> = (node, {id, defaultValue, compute}) => {

        const subscriber = (value: number) => {
            const result = compute(value)
            if (typeof result === 'string') {
                node.style.minWidth = node.style.maxWidth = result
            } else {
                node.style.minWidth = result.min
                node.style.maxWidth = result.max
            }
        }

        let unsubscribe = getDividerStore(id, defaultValue).subscribe(subscriber)

        return {
            update(args) {
                compute = args.compute
                if (args.id != id) {
                    unsubscribe()
                    id = args.id
                    unsubscribe = getDividerStore(args.id, args.defaultValue).subscribe(subscriber)
                }
            },
            destroy() {
                unsubscribe()
            }
        }
    }
</script>
<script lang="ts">
    export let id: string

    let divider: HTMLElement|null = null
    let offset = 0
    let dragging = false


    function onMouseMove(event: MouseEvent) {
        event.preventDefault()
        if (divider?.parentElement) {
            const width = ((event.x - offset) / divider.parentElement.clientWidth)
            dividerStore.updateDivider(id, width)
        }
    }

    function endResize() {
        dragging = false
        window.removeEventListener('mousemove', onMouseMove)
        window.removeEventListener('mouseup', endResize)
    }

    function startResize(event: MouseEvent) {
        event.preventDefault()
        if (divider?.parentElement ) {
            dragging = true
            offset = divider.parentElement.getBoundingClientRect().x + divider.clientWidth
            window.addEventListener('mousemove', onMouseMove)
            window.addEventListener('mouseup', endResize)
        }
    }
</script>

<div role="separator" bind:this={divider} class:dragging on:mousedown={startResize} />

<style lang="scss">
    div {
        width: 4px;
        background-color: var(--border-color);
        cursor: col-resize;


        &.dragging {
            background-color: var(--oc-blue-3);
        }
    }
</style>
