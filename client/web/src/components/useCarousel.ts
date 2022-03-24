import { useCallback, useEffect, useRef, useState } from 'react'

import { isEqual } from 'lodash'
import { Subscription } from 'rxjs'

import { observeResize } from '@sourcegraph/common'

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
    const [carousel, setCarousel] = useState<HTMLElement | null>()
    const nextCarousel = useCallback((carousel: HTMLElement) => {
        setCarousel(carousel)
    }, [])

    const [scrollability, setScrollability] = useState(defaultCarouselState)

    const scrollabilityReference = useRef(scrollability)
    scrollabilityReference.current = scrollability

    // Listen for UIEvents that can affect scrollability (e.g. scroll, resize)
    useEffect(() => {
        function onScroll(): void {
            if (carousel) {
                const newScrollability = carouselScrollHandlers[direction](carousel)
                if (!isEqual(scrollabilityReference.current, newScrollability)) {
                    setScrollability(newScrollability)
                }
            }
        }

        carousel?.addEventListener('scroll', onScroll)

        let subscription: Subscription | undefined

        if (carousel) {
            subscription = observeResize(carousel).subscribe(() => {
                const newScrollability = carouselScrollHandlers[direction](carousel)

                if (!isEqual(scrollabilityReference.current, newScrollability)) {
                    setScrollability(newScrollability)
                }
            })

            // Check initial scroll state
            const newScrollability = carouselScrollHandlers[direction](carousel)
            if (!isEqual(scrollabilityReference.current, newScrollability)) {
                setScrollability(newScrollability)
            }
        }

        return () => {
            carousel?.removeEventListener('scroll', onScroll)
            subscription?.unsubscribe()
        }
    }, [carousel, direction])

    // Handle negative and positive click events
    const onNegativeClicked = useCallback(() => {
        if (carousel) {
            carouselClickHandlers[direction]({ sign: 'negative', amountToScroll, carousel })
        }
    }, [direction, amountToScroll, carousel])

    const onPositiveClicked = useCallback(() => {
        if (carousel) {
            carouselClickHandlers[direction]({ sign: 'positive', amountToScroll, carousel })
        }
    }, [direction, amountToScroll, carousel])

    return {
        canScrollNegative: scrollability.canScrollNegative,
        canScrollPositive: scrollability.canScrollPositive,
        onNegativeClicked,
        onPositiveClicked,
        carouselReference: nextCarousel,
    }
}
