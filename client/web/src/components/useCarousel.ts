import { isEqual } from 'lodash'
import { useCallback, useMemo } from 'react'
import { fromEvent, merge, of, ReplaySubject, Subject } from 'rxjs'
import { map, switchMap, withLatestFrom, tap, distinctUntilChanged } from 'rxjs/operators'
import { useObservable } from '../../../shared/src/util/useObservable'
import { observeResize } from '../util/dom'

interface CarouselOptions {
    amountToScroll?: number
    direction: CarouselDirection
}

type CarouselDirection = 'leftToRight' | 'topToBottom'

interface CarouselState {
    canScrollNegative: boolean
    canScrollPositive: boolean
    onNegativeClicked: () => void
    onPositiveClicked: () => void
    carouselReference: React.RefCallback<HTMLElement>
}

const defaultCarouselState = { canScrollNegative: false, canScrollPositive: false }

const carouselScrollHandlers: Record<
    CarouselDirection,
    (carousel: HTMLElement) => Pick<CarouselState, 'canScrollNegative' | 'canScrollPositive'>
> = {
    leftToRight: carousel => ({
        canScrollNegative: carousel.scrollLeft > 0,
        canScrollPositive: carousel.scrollLeft + carousel.clientWidth < carousel.scrollWidth,
    }),
    topToBottom: carousel => ({
        canScrollNegative: carousel.scrollTop > 0,
        canScrollPositive: carousel.scrollTop + carousel.clientHeight < carousel.scrollHeight,
    }),
}

const carouselClickHandlers: Record<
    CarouselDirection,
    (options: { carousel: HTMLElement; amountToScroll: number; sign: 'positive' | 'negative' }) => void
> = {
    leftToRight: ({ carousel, amountToScroll, sign }) => {
        const width = carousel.clientWidth
        carousel.scrollBy({
            top: 0,
            left: sign === 'positive' ? width * amountToScroll : -(width * amountToScroll),
            behavior: 'smooth',
        })
    },
    topToBottom: ({ carousel, amountToScroll, sign }) => {
        const height = carousel.clientHeight
        carousel.scrollBy({
            top: sign === 'positive' ? height * amountToScroll : -(height * amountToScroll),
            left: 0,
            behavior: 'smooth',
        })
    },
}

export function useCarousel({ amountToScroll = 0.9, direction }: CarouselOptions): CarouselState {
    const carouselReferences = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
    const nextCarouselReference = useCallback((carousel: HTMLElement) => carouselReferences.next(carousel), [
        carouselReferences,
    ])

    const clicks = useMemo(() => new Subject<'positive' | 'negative'>(), [])

    const nextNegativeClick = useCallback(() => clicks.next('negative'), [clicks])
    const nextPositiveClick = useCallback(() => clicks.next('positive'), [clicks])

    // Listen for UIEvents that can affect scrollability (e.g. scroll, resize)
    const { canScrollNegative, canScrollPositive } =
        useObservable(
            useMemo(
                () =>
                    carouselReferences.pipe(
                        switchMap(carousel => {
                            if (!carousel) {
                                return of(defaultCarouselState)
                            }

                            // Initial scroll state
                            const initial = of(undefined)
                            const scrolls = fromEvent<React.UIEvent<HTMLElement>>(carousel, 'scroll')
                            const resizes = observeResize(carousel)

                            return merge(initial, scrolls, resizes).pipe(
                                map(() => carouselScrollHandlers[direction](carousel)),
                                distinctUntilChanged((a, b) => isEqual(a, b))
                            )
                        })
                    ),
                [direction, carouselReferences]
            )
        ) || defaultCarouselState

    // Handle negative and positive click events
    useObservable(
        useMemo(
            () =>
                clicks.pipe(
                    withLatestFrom(carouselReferences),
                    tap(([sign, carousel]) => {
                        if (carousel) {
                            // TODO: check if it can be scrolled before scrolling.
                            // Not urgent, since the component shouldn't allow invalid scrolls,
                            // and it's a noop regardless.
                            carouselClickHandlers[direction]({ sign, amountToScroll, carousel })
                        }
                    })
                ),
            [amountToScroll, direction, clicks, carouselReferences]
        )
    )

    return {
        canScrollNegative,
        canScrollPositive,
        onNegativeClicked: nextNegativeClick,
        onPositiveClicked: nextPositiveClick,
        carouselReference: nextCarouselReference,
    }
}
