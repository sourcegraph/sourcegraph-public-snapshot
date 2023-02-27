<script lang="ts" context="module">
    import { writable } from 'svelte/store'
    export const lineWrap = writable<boolean>(false)
</script>

<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import { mdiWrap, mdiWrapDisabled } from '@mdi/js'
</script>

<button on:click={() => lineWrap.update(wrap => !wrap)}
    ><Icon svgPath={$lineWrap ? mdiWrap : mdiWrapDisabled} inline /></button
>

<style lang="scss">
    button {
        margin: 0;
        padding: 0;
        border: 0;
        background-color: transparent;
        cursor: pointer;
        margin-right: 1rem;
    }
</style>
