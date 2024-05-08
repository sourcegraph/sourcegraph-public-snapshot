<script lang="ts">
    import { onMount } from 'svelte'
    import type { FormEventHandler } from 'svelte/elements'
    import type { ActionReturn } from 'svelte/action'
    import LoadingSpinner from '../LoadingSpinner.svelte'

    export let value: string
    export let placeholder: string | undefined = undefined
    export let autofocus: boolean | undefined = undefined
    export let onInput: FormEventHandler<HTMLInputElement> | undefined = undefined
    export let input: HTMLInputElement | undefined = undefined
    export let actions: Array<(node: HTMLInputElement) => ActionReturn> = []
    export let loading: boolean = false

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

<div class="root" data-input-container>
    <input
        bind:this={input}
        type="text"
        use:bindAction
        {value}
        {autofocus}
        {placeholder}
        class:loading
        on:input={onInput}
        on:keydown
        {...$$restProps}
    />

    {#if loading}
        <span class="loader">
            <LoadingSpinner inline />
        </span>
    {/if}
</div>

<style lang="scss">
    .root {
        position: relative;
    }

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

    .loading {
        padding-right: 1.5rem;
    }

    .loader {
        position: absolute;
        right: 0.5rem;
        top: 50%;
        transform: translateY(-50%);
    }
</style>
