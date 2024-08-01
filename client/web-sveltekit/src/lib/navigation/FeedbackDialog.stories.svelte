<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'

    import FeedbackDialog from './FeedbackDialog.svelte'

    export const meta = {
        component: FeedbackDialog,
    }
</script>

<Story name="Default">
    <FeedbackDialog handleOptOut={() => {}} />
</Story>
