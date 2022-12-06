<svelte:options immutable />

<script lang="ts">
    import { Option } from './suggestions'
    export let option: Option
</script>

<span class="box">
    {#if option.matches}
        {#each [...option.value] as char, index}
            {#if option.matches.has(index)}
                <span class="match">{char}</span>
            {:else}
                {char}
            {/if}
        {/each}
    {:else}
        {option.value}
    {/if}
    <span class="separator">:</span>
</span>

<style lang="scss">
    .box {
        // Used to make sure that there is no gap between text and colon
        // (svelte seems to insert a space betwen those no matter what we do)
        display: flex;
        font-family: var(--code-font-family);
        font-size: 12px;
        color: var(--search-filter-keyword-color);
        background-color: var(--oc-blue-0);
        // border: 1px solid var(--oc-blue-1);
        border-radius: 3px;
        padding: 0px;
    }

    .match {
        font-weight: bold;
    }
    .separator {
        color: var(--search-filter-keyword-color);
    }
</style>
