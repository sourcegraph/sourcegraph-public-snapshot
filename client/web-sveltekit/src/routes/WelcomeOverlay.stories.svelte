<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'

    import WelcomeOverlay from './WelcomeOverlay.svelte'

    export const meta = {
        component: WelcomeOverlay,
    }
</script>

<Story name="Default">
    <WelcomeOverlay handleDismiss={() => {}} show />
</Story>
