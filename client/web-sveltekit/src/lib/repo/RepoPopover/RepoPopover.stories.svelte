<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'

    import RepoPopover from './RepoPopover.svelte'

    export const meta = {
        component: RepoPopover,
    }
</script>

<script lang="ts">
    let name = 'pete.roth06'
</script>

<Story name="Default">
    <h2>With name</h2>
    <RepoPopover {name} />
    <br />
    <h2>Without name</h2>
    <RepoPopover />
</Story>
