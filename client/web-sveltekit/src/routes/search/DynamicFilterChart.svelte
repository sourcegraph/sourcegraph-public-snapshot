<script lang="ts">
    import { ticks } from 'd3-array'
    import { scaleLinear } from 'd3-scale'

    import type { SectionItem } from '$lib/search/dynamicFilters'
    import Tooltip from '$lib/Tooltip.svelte'

    export let label: string
    export let items: SectionItem[]

    $: items = items.slice(0, 20)

    const padding = { top: 20, right: 15, bottom: 20, left: 25 }

    let width: number
    let height: number

    $: xTicks = items.map(item => item.label)
    $: yMax = Math.max(...items.map(item => item.count ?? 0))
    $: yTicks = ticks(0, yMax, 5)

    $: xScale = scaleLinear()
        .domain([0, xTicks.length])
        .range([padding.left, width - padding.right])

    $: yScale = scaleLinear()
        .domain([0, Math.max.apply(null, yTicks)])
        .range([height - padding.bottom, padding.top])

    $: innerWidth = width - (padding.left + padding.right)
    $: barWidth = innerWidth / xTicks.length
</script>

<div class="chart" bind:clientWidth={width} bind:clientHeight={height}>
    <svg>
        <!-- y axis -->
        <g class="axis y-axis">
            {#each yTicks as tick}
                <g class="tick tick-{tick}" transform="translate(0, {yScale(tick)})">
                    <line x2="100%" />
                    <text y="-4">{tick}</text>
                </g>
            {/each}
        </g>

        <g class="bars">
            {#each items as item, i}
                <Tooltip tooltip={item.label}>
                    <rect
                        x={xScale(i) + 2}
                        y={yScale(item.count ?? 0)}
                        width={barWidth - 4}
                        height={yScale(0) - yScale(item.count ?? 0)}
                    />
                </Tooltip>
            {/each}
        </g>
    </svg>
</div>

<style>
    h2 {
        text-align: center;
    }

    .chart {
        width: 100%;
        max-width: 800px;
        margin: 0 auto;
    }

    svg {
        position: relative;
        width: 100%;
        height: 500px;
    }

    .tick {
        font-family: Helvetica, Arial;
        font-size: 0.725em;
        font-weight: 200;
    }

    .tick line {
        stroke: #e2e2e2;
        stroke-dasharray: 2;
    }

    .tick text {
        text-anchor: start;
    }

    .tick.tick-0 line {
        stroke-dasharray: 0;
    }

    .bars rect {
        fill: var(--brand-secondary);
        stroke: none;
        opacity: 0.65;
    }
</style>
