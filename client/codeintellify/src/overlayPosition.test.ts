import * as assert from 'assert'

import { describe, it } from 'vitest'

import { calculateOverlayPosition, type HasGetBoundingClientRect } from './overlayPosition'

const rectangle = (left: number, top: number, width: number, height: number): HasGetBoundingClientRect => ({
    getBoundingClientRect: () => ({
        x: left,
        y: top,
        toJSON: () => 'unimplemented',
        left,
        top,
        width,
        height,
        right: left + width,
        bottom: top + height,
    }),
})

describe('overlay_position', () => {
    describe('calculateOverlayPosition()', () => {
        describe('with the scrolling element being document', () => {
            describe('not scrolled', () => {
                it('should return a position above the given target if the overlay fits above', () => {
                    assert.deepStrictEqual(
                        calculateOverlayPosition({
                            relativeElement: {
                                ...rectangle(28, 20, 800, 600),
                                scrollLeft: 0,
                                scrollTop: 0,
                            },
                            target: rectangle(128, 220, 60, 16),
                            hoverOverlayElement: rectangle(28, 38, 350, 150),
                            windowInnerHeight: 500,
                            windowScrollY: 0,
                        }),
                        { left: 100, bottom: 400 }
                    )
                })

                it('should return a position below the a given target if the overlay does not fit above', () => {
                    assert.deepStrictEqual(
                        calculateOverlayPosition({
                            relativeElement: {
                                ...rectangle(28, 20, 800, 600),
                                scrollLeft: 0,
                                scrollTop: 0,
                            },
                            target: rectangle(128, 120, 60, 16),
                            hoverOverlayElement: rectangle(28, 38, 350, 150),
                            windowInnerHeight: 500,
                            windowScrollY: 0,
                        }),
                        { left: 100, top: 116 }
                    )
                })
            })
            describe('scrolled', () => {
                it('should return a position above the given target if the overlay fits above', () => {
                    assert.deepStrictEqual(
                        calculateOverlayPosition({
                            relativeElement: {
                                ...rectangle(28, -380, 800, 3000),
                                scrollLeft: 0,
                                scrollTop: 0,
                            },
                            target: rectangle(128, 220, 60, 16),
                            hoverOverlayElement: rectangle(28, -362, 350, 150),
                            windowInnerHeight: 500,
                            windowScrollY: 0,
                        }),
                        { left: 100, bottom: 2400 }
                    )
                })

                it('should return a position below the a given target if the overlay does not fit above', () => {
                    assert.deepStrictEqual(
                        calculateOverlayPosition({
                            relativeElement: {
                                ...rectangle(28, -380, 800, 3000),
                                scrollLeft: 0,
                                scrollTop: 0,
                            },
                            target: rectangle(128, 70, 60, 16),
                            hoverOverlayElement: rectangle(28, -362, 350, 150),
                            windowInnerHeight: 500,
                            windowScrollY: 0,
                        }),
                        { left: 100, top: 466 }
                    )
                })
            })
        })

        describe('with the scrolling element being the relativeElement', () => {
            describe('not scrolled', () => {
                it('should return a position above the given target if the overlay fits above2', () => {
                    assert.deepStrictEqual(
                        calculateOverlayPosition({
                            relativeElement: {
                                ...rectangle(28, 20, 800, 600),
                                scrollLeft: 0,
                                scrollTop: 0,
                            },
                            target: rectangle(128, 220, 60, 16),
                            hoverOverlayElement: rectangle(28, 38, 350, 150),
                            windowInnerHeight: 500,
                            windowScrollY: 0,
                        }),
                        { left: 100, bottom: 400 }
                    )
                })

                it('should return a position below the a given target if the overlay does not fit above', () => {
                    assert.deepStrictEqual(
                        calculateOverlayPosition({
                            relativeElement: {
                                ...rectangle(28, 20, 800, 600),
                                scrollLeft: 0,
                                scrollTop: 0,
                            },
                            target: rectangle(128, 160, 60, 16),
                            hoverOverlayElement: rectangle(28, 38, 350, 150),
                            windowInnerHeight: 500,
                            windowScrollY: 0,
                        }),
                        { left: 100, top: 156 }
                    )
                })
            })
            describe('scrolled', () => {
                it('should return a position above the given target if the overlay fits above', () => {
                    assert.deepStrictEqual(
                        calculateOverlayPosition({
                            relativeElement: {
                                ...rectangle(28, 20, 800, 600),
                                scrollLeft: 200,
                                scrollTop: 400,
                            },
                            target: rectangle(128, 220, 60, 16),
                            hoverOverlayElement: rectangle(-172, -362, 350, 150),
                            windowInnerHeight: 500,
                            windowScrollY: 0,
                        }),
                        { left: 300, bottom: 0 }
                    )
                })

                it('should return a position below the a given target if the overlay does not fit above', () => {
                    assert.deepStrictEqual(
                        calculateOverlayPosition({
                            relativeElement: {
                                ...rectangle(28, 20, 800, 600),
                                scrollLeft: 200,
                                scrollTop: 400,
                            },
                            target: rectangle(128, 70, 60, 16),
                            hoverOverlayElement: rectangle(-172, -362, 350, 150),
                            windowInnerHeight: 500,
                            windowScrollY: 0,
                        }),
                        { left: 300, top: 466 }
                    )
                })
            })
        })

        describe('without a relativeElement', () => {
            it('should return a position above the given target if the overlay fits above', () => {
                assert.deepStrictEqual(
                    calculateOverlayPosition({
                        relativeElement: undefined,
                        target: rectangle(128, 220, 60, 16),
                        hoverOverlayElement: rectangle(-172, -362, 350, 150),
                        windowInnerHeight: 500,
                        windowScrollY: 0,
                    }),
                    { left: 128, bottom: 280 }
                )
            })

            it('should return a position below the a given target if the overlay does not fit above', () => {
                assert.deepStrictEqual(
                    calculateOverlayPosition({
                        relativeElement: undefined,
                        target: rectangle(128, 70, 60, 16),
                        hoverOverlayElement: rectangle(-172, -362, 350, 150),
                        windowInnerHeight: 500,
                        windowScrollY: 0,
                    }),
                    { left: 128, top: 86 }
                )
            })
        })
    })
})
