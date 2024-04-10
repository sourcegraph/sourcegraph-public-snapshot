<script lang="ts" context="module">
    // import { faker } from '@faker-js/faker'
    import { Story } from '@storybook/addon-svelte-csf'

    import RepoMenu from './RepoMenu.svelte'

    export const meta = {
        component: RepoMenu,
    }
</script>

<Story name="Default">
    <h2>With name</h2>
    <RepoMenu />
    <br />
</Story>
