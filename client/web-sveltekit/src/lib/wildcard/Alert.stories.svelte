<script lang="ts" context="module">
    import Alert from './Alert.svelte'
    import { Story } from '@storybook/addon-svelte-csf'

    export const meta = {
        component: Alert,
    }
</script>

<Story name="Default">
    <Alert variant="info">This is an info alert</Alert>
    <Alert variant="danger">This is a danger alert</Alert>
</Story>
