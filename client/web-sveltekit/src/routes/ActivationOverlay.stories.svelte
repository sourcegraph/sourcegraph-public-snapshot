<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'

    import ActivationOverlay from './ActivationOverlay.svelte'

    export const meta = {
        component: ActivationOverlay,
    }
</script>

<Story name="Default">
    <ActivationOverlay onDismiss={() => {}} />
</Story>
