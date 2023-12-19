<script lang="ts">
    import Timestamp from '$lib/Timestamp.svelte'
    import { faker } from '@faker-js/faker'
    import type { ComponentProps } from 'svelte'

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

<h2>Timestamp props</h2>
<table>
    {#each cases as [title, props]}
        <tr>
            <th>{title}</th>
            <td><Timestamp {date} {...props} /></td>
        </tr>
    {/each}
</table>

<style lang="scss">
    td,
    th {
        padding: 0.5rem;
    }
</style>
