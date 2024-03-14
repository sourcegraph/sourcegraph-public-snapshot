<script lang="ts">
    import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'

    import { pluralize } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import { Button } from '$lib/wildcard'

    import type { HighlightedRange } from './static-highlights'

    export let ranges: HighlightedRange[]
    export let handlePrevious: () => void
    export let handleNext: () => void

    const totalMatches = ranges.length
    const selectedIdx = ranges.findIndex(range => range.selected)
</script>

<div class="root">
    {#if totalMatches > 1}
        <div>
            <Button
                size="sm"
                outline
                variant="secondary"
                on:click={handlePrevious}
                data-testid="blob-view-static-previous"
                aria-label="previous result"
            >
                <Icon svgPath={mdiChevronLeft} aria-hidden={true} />
            </Button>

            <Button
                size="sm"
                outline
                variant="secondary"
                on:click={handleNext}
                data-testid="blob-view-static-next"
                aria-label="next result"
            >
                <Icon svgPath={mdiChevronRight} aria-hidden={true} />
            </Button>
        </div>
    {/if}
    <text>
        {selectedIdx === -1 ? '' : `${selectedIdx + 1} of `}
        {totalMatches}
        {pluralize('result', totalMatches)}
    </text>
</div>
