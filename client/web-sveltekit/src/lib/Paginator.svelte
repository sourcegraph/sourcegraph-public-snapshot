<script lang="ts">
    import { mdiPageFirst, mdiPageLast, mdiChevronRight, mdiChevronLeft } from '@mdi/js'

    import { page } from '$app/stores'

    import Icon from './Icon.svelte'
    import { Button } from './wildcard'
    import { Param } from './Paginator'

    type PageInfo =
        // Bidirection pagination
        | {hasPreviousPage: boolean, hasNextPage: boolean, startCursor: string | null, endCursor: string | null}
        // Unidirection pagination
        | {hasNextPage: boolean, hasPreviousPage: boolean, endCursor: string |null, startCursor?: undefined, previousEndCursor: string | null}

    export let pageInfo: PageInfo
    export let disabled: boolean = false
    export let showLastpageButton: boolean = true

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
    $: previousPageURL = pageInfo.startCursor !== undefined ?
        urlWithParameter(Param.before, pageInfo.startCursor) :
        urlWithParameter(Param.after, pageInfo.previousEndCursor)
    $: nextPageURL = urlWithParameter(Param.after, pageInfo.endCursor)
    $: firstAndPreviousDisabled = disabled || !pageInfo.hasPreviousPage
    $: nextAndLastDisabled = disabled || !pageInfo.hasNextPage
</script>

<!-- The event handler is used for event delegation -->
<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
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
    {#if showLastpageButton}
        <Button variant="secondary" outline>
            <a slot="custom" let:className class={className} href={lastPageURL} aria-disabled={nextAndLastDisabled}>
                <Icon svgPath={mdiPageLast} inline />
            </a>
        </Button>
    {/if}
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
