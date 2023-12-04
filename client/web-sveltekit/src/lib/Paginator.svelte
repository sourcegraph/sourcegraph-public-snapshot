<script context="module" lang="ts">
    enum Param {
        before = '$before',
        after = '$after',
        last = '$last',
    }

    export function getPaginationParams(
        searchParams: URLSearchParams,
        pageSize: number
    ):
        | { first: number; last: null; before: null; after: string | null }
        | { first: null; last: number; before: string | null; after: null } {
        if (searchParams.has('$before')) {
            return { first: null, last: pageSize, before: searchParams.get(Param.before), after: null }
        } else if (searchParams.has('$after')) {
            return { first: pageSize, last: null, before: null, after: searchParams.get(Param.after) }
        } else if (searchParams.has('$last')) {
            return { first: null, last: pageSize, before: null, after: null }
        } else {
            return { first: pageSize, last: null, before: null, after: null }
        }
    }
</script>

<script lang="ts">
    import { mdiPageFirst, mdiPageLast, mdiChevronRight, mdiChevronLeft } from '@mdi/js'

    import { page } from '$app/stores'

    import Icon from './Icon.svelte'
    import { Button } from './wildcard'

    export let pageInfo: {
        hasPreviousPage: boolean
        hasNextPage: boolean
        startCursor: string | null
        endCursor: string | null
    }
    export let disabled: boolean

    function urlWithParameter(name: string, value: string | null): string {
        const url = new URL($page.url)
        url.searchParams.delete(Param.before)
        url.searchParams.delete(Param.after)
        url.searchParams.delete(Param.last)
        if (value !== null) {
            url.searchParams.set(name, value)
        }
        return url.toString()
    }

    function preventClickOnDisabledLink(event: MouseEvent) {
        const target = event.target as HTMLElement
        if (target.closest('a[aria-disabled="true"]')) {
            event.preventDefault()
        }
    }

    let firstPageURL = urlWithParameter('', null)
    let lastPageURL = urlWithParameter(Param.last, '')
    $: previousPageURL = urlWithParameter(Param.before, pageInfo.startCursor)
    $: nextPageURL = urlWithParameter(Param.after, pageInfo.endCursor)
    $: firstAndPreviousDisabled = disabled || !pageInfo.hasPreviousPage
    $: nextAndLastDisabled = disabled || !pageInfo.hasNextPage
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- The event handler is used for event delegation -->
<div on:click={preventClickOnDisabledLink}>
    <Button variant="secondary" outline>
        <a slot="custom" let:className href={firstPageURL} class={className} aria-disabled={firstAndPreviousDisabled}>
            <Icon svgPath={mdiPageFirst} inline />
        </a>
    </Button>
    <Button variant="secondary" outline>
        <a
            slot="custom"
            let:className
            class={className}
            href={previousPageURL}
            aria-disabled={firstAndPreviousDisabled}
        >
            <Icon svgPath={mdiChevronLeft} inline />Previous
        </a>
    </Button>
    <Button variant="secondary" outline>
        <a slot="custom" let:className class={className} href={nextPageURL} aria-disabled={nextAndLastDisabled}>
            Next <Icon svgPath={mdiChevronRight} inline />
        </a>
    </Button>
    <Button variant="secondary" outline>
        <a slot="custom" let:className class={className} href={lastPageURL} aria-disabled={nextAndLastDisabled}>
            <Icon svgPath={mdiPageLast} inline />
        </a>
    </Button>
</div>

<style lang="scss">
    a {
        color: var(--body-color);

        &:first-child {
            margin-right: 1rem;
        }
        &:last-child {
            margin-left: 1rem;
        }
    }
</style>
