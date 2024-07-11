<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'

    import Button from '$lib/wildcard/Button.svelte'

    import { BUTTON_VARIANTS } from './Button'

    export const meta = {
        component: Button,
    }
</script>

<Story name="Default">
    <h2>Default</h2>
    <table>
        <tbody>
            {#each BUTTON_VARIANTS as variant}
                <tr>
                    <td><Button {variant}>{variant}</Button> </td>
                    <td><Button {variant} disabled>Disabled</Button></td>
                </tr>
            {/each}
        </tbody>
    </table>

    <h2>Sizes</h2>
    <table>
        <tbody>
            {#each BUTTON_VARIANTS as variant}
                <tr>
                    <td><Button {variant}>Button</Button></td>
                    <td><Button {variant} size="sm">Small</Button></td>
                    <td><Button {variant} size="lg">Large</Button></td>
                </tr>
            {/each}
        </tbody>
    </table>

    <h2>Outline</h2>
    <table>
        <tbody>
            {#each BUTTON_VARIANTS as variant}
                <tr>
                    <td><Button {variant} outline>{variant}</Button></td>
                    <td><Button {variant} outline disabled>Disabled</Button></td>
                </tr>
            {/each}
        </tbody>
    </table>

    <style lang="scss">
        td,
        th {
            padding: 0.25rem;
        }
    </style>
</Story>
