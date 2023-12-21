<script lang="ts">
    import Badge, { BADGE_VARIANTS } from '$lib/wildcard/Badge.svelte'
</script>

<h2>Default</h2>
<table>
    <thead>
        <tr><th>Variant</th><th /><th>pill=true</th><th>small=true</th><th>pill=true small=true</th></tr>
    </thead>
    <tbody>
        {#each BADGE_VARIANTS as variant}
            <tr>
                <th>{variant}</th>
                <td><Badge {variant}>Badge</Badge> </td>
                <td><Badge {variant} pill>Badge</Badge></td>
                <td><Badge {variant} small>Badge</Badge></td>
                <td><Badge {variant} pill small>Badge</Badge></td>
            </tr>
        {/each}
    </tbody>
</table>

<h2 id="custom">Custom elements</h2>
<Badge variant="primary">
    <a slot="custom" let:class={cls} class={cls} href="#custom">I'm a link</a>
</Badge>

<style lang="scss">
    td,
    th {
        padding: 0.5rem;
    }
</style>
