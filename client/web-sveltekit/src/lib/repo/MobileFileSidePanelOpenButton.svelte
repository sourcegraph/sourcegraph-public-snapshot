<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import { fileTreeSidePanel } from '$lib/repo/stores'
</script>

<div class="visible-mobile">
    <button aria-label="Open file sidebar" on:click={() => $fileTreeSidePanel?.expand()}>
        <Icon icon={ILucidePanelLeftOpen} aria-hidden="true" inline />
    </button>
</div>

<style lang="scss">
    button {
        all: unset;
        cursor: pointer;
    }
</style>
