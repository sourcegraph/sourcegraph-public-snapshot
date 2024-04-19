<script lang="ts" context="module">
    // import { faker } from '@faker-js/faker'
    import { Story } from '@storybook/addon-svelte-csf'

    // import type { LastCommitFragment } from './LastCommit.gql'
    import NodeLine from './NodeLine.svelte'

    export const meta = {
        component: NodeLine,
    }
</script>

<Story name="Default">
    <div>
        <NodeLine />
    </div>
</Story>

<style lang="scss">
    div {
        width: fit-content;
        height: 100px;
    }
</style>
