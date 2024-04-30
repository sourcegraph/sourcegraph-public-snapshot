<script lang="ts">
    import { ticks } from 'd3-array'
    import { scaleLinear } from 'd3-scale'

    import type { SectionItem } from '$lib/search/dynamicFilters'
    import Tooltip from '$lib/Tooltip.svelte'

    export let title: string
    export let items: SectionItem[]

    const padding = { top: 30, right: 20, bottom: 50, left: 30 }

    let width: number
    let height: number

    $: xTicks = items.map(item => item.label)
    $: yMax = Math.max(...items.map(item => item.count ?? 0))
    $: yTicks = ticks(0, yMax, 5)

    $: xScale = scaleLinear()
        .domain([0, xTicks.length])
        .range([padding.left, width - padding.right])

    $: yScale = scaleLinear()
        .domain([0, yMax])
        .range([height - padding.bottom, padding.top])

    $: innerWidth = width - (padding.left + padding.right)
    $: barWidth = innerWidth / xTicks.length
</script>

<div class="chart" bind:clientWidth={width} bind:clientHeight={height}>
    <h2>{title}</h2>
    <svg>
        <!-- y axis -->
        <g class="axis y-axis">
            {#each yTicks as tick}
                <g class="tick tick-{tick}" transform="translate({padding.left}, {yScale(tick)})">
                    <line x="0" x2="100%" />
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
    .chart {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        width: 100%;
        height: 100%;
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
        text-anchor: end;
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
