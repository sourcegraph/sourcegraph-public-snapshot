<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'

    import { SymbolKind } from '$lib/graphql-types'

    import SymbolKindIcon from './SymbolKindIcon.svelte'

    export const meta = {
        component: SymbolKindIcon,
    }
</script>

<script lang="ts">
    let size = '16'
</script>

<Story name="Default">
    <p>Adjust size:</p>
    <input type="range" min="0" max="100" bind:value={size} />

    <div>
        {#each Object.values(SymbolKind) as symbolKind}
            <SymbolKindIcon {symbolKind} --icon-size="{size}px" />
        {/each}
    </div>
</Story>

<style lang="scss">
    div {
        display: flex;
        flex-flow: row wrap;
        gap: 0.5rem;
        max-width: 21rem;
    }
</style>
