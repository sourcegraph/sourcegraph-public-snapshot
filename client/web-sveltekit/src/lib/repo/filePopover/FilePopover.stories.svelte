<script lang="ts" context="module">
    // import { faker } from '@faker-js/faker'
    import { Story } from '@storybook/addon-svelte-csf'

    // import type { LastCommitFragment } from './LastCommit.gql'
    import FilePopover from './FilePopover.svelte'

    export const meta = {
        component: FilePopover,
    }
</script>

<Story name="Default">
    <FilePopover hello="world" />
</Story>
