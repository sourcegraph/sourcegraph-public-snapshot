<script lang="ts">
    import { onMount } from 'svelte'
    import type { FormEventHandler } from 'svelte/elements'

    export let value: string
    export let placeholder: string | undefined
    export let autofocus: boolean | undefined
    export let onInput: FormEventHandler<HTMLInputElement> | undefined = undefined
    export let input: HTMLInputElement | undefined = undefined
    export let actions: Array<(node: HTMLInputElement) => unknown>

    $: bindAction = function bindAction(node: HTMLInputElement) {
        if (actions.length === 0) {
            return
        }

        const actionsResults = actions.map(action => action(node))

        return {
            destroy() {
                actionsResults.map(result => result?.destroy?.())
            },
        }
    }

    onMount(() => {
        if (autofocus) {
            requestAnimationFrame(() => {
                input?.focus()
            })
        }
    })
</script>

<input
    bind:this={input}
    type="text"
    use:bindAction
    {value}
    {autofocus}
    {placeholder}
    on:input={onInput}
    {...$$restProps}
/>

<style lang="scss">
    input {
        display: block;
        width: 100%;
        height: var(--input-height);
        padding: var(--input-padding-y) var(--input-padding-x);
        color: var(--input-color);
        font-size: var(--input-font-size);
        font-weight: var(--input-font-weight);
        line-height: var(--input-line-height);
        background-color: var(--input-bg);
        background-clip: padding-box;
        border: var(--input-border-width) solid var(--input-border-color);
        border-radius: var(--border-radius);
    }
</style>
