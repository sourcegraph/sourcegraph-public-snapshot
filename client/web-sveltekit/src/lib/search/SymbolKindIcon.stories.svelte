<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'

    import { SymbolKind } from '$lib/graphql-types'

    import SymbolKindIcon from './SymbolKindIcon.svelte'

    export const meta = {
        component: SymbolKindIcon,
    }
</script>

<Story name="Default">
    <div>
        {#each Object.values(SymbolKind) as symbolKind}
            <SymbolKindIcon {symbolKind} />
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
