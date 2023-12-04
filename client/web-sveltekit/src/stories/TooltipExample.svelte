<script lang="ts">
    import Tooltip, { type Placement } from '$lib/Tooltip.svelte'

    const placements: Placement[] = ['auto', 'top', 'left', 'bottom', 'right']
    let count = 0
</script>

<h2>Static Tooltip</h2>
<p>
    <Tooltip tooltip="Some static tooltip text" placement="auto">
        <span>Hover over me</span>
    </Tooltip>
</p>

<h2>Dynamic Tooltip</h2>
<p>
    Move the mouse over
    <Tooltip tooltip="This text was moused over {count} times" placement="top">
        <strong on:mousemove={() => count++}>this text</strong>
    </Tooltip>
    to update the counter in the tooltip.
</p>

<h2>Tooltip placement</h2>
<table>
    {#each placements as placement}
        <tr>
            <th>{placement}</th>
            <td><Tooltip tooltip="Tooltip" {placement} alwaysVisible><span>Trigger</span></Tooltip></td>
        </tr>
    {/each}
</table>

<style lang="scss">
    td,
    th {
        padding: 2rem;
    }
</style>
