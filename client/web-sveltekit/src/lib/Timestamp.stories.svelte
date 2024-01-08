<script lang="ts" context="module">
    import Timestamp from '$lib/Timestamp.svelte'
    import { faker } from '@faker-js/faker'
    import { Story } from '@storybook/addon-svelte-csf'
    import type { ComponentProps } from 'svelte'

    export const meta = {
        component: Timestamp,
    }
</script>

<script lang="ts">
    faker.seed(1)
    const date = faker.date.recent()
    const cases: [string, Partial<ComponentProps<Timestamp>>][] = [
        ['default', {}],
        ['with ago', { hideSuffix: true }],
        ['strict', { strict: true }],
        ['utc', { utc: true }],
        ['absolute', { showAbsolute: true }],
        ['with ago, strict', { hideSuffix: true, strict: true }],
        ['absolute, utc', { showAbsolute: true, utc: true }],
    ]
</script>

<Story name="Default">
    <h2>Timestamp props</h2>
    <table>
        {#each cases as [title, props]}
            <tr>
                <th>{title}</th>
                <td><Timestamp {date} {...props} /></td>
            </tr>
        {/each}
    </table>
</Story>

<style lang="scss">
    td,
    th {
        padding: 0.5rem;
    }
</style>
