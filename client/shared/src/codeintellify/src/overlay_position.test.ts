import * as assert from 'assert'
import { calculateOverlayPosition, CSSOffsets } from './overlay_position'

describe('overlay_position', () => {
    describe('calculateOverlayPosition()', () => {
        /** Positions an element at the given px offsets */
        const applyOffsets = (element: HTMLElement, { left, top }: CSSOffsets): void => {
            element.style.left = left + 'px'
            element.style.top = top + 'px'
        }

        let relativeElement: HTMLElement
        let hoverOverlayElement: HTMLElement

        /** Puts a target `div` into the relativeElement at the given px offsets */
        const createTarget = (position: CSSOffsets): HTMLElement => {
            const target = document.createElement('div')
            target.className = 'target'
            target.textContent = 'target'
            applyOffsets(target, position)
            relativeElement.appendChild(target)
            return target
        }

        beforeEach(() => {
            const style = document.createElement('style')
            style.innerHTML = `
                * {
                    box-sizing: border-box;
                }
                .relative-element {
                    background: lightgray;
                    margin: 20px;
                    width: 800px;
                    position: relative;
                }
                .hover-overlay-element {
                    background: gray;
                    width: 350px;
                    height: 150px;
                    position: absolute;
                }
                .target {
                    background: orange;
                    width: 60px;
                    height: 16px;
                    position: absolute;
                }
            `
            document.head.appendChild(style)

            relativeElement = document.createElement('div')
            relativeElement.className = 'relative-element'
            relativeElement.textContent = 'relativeElement'
            document.body.appendChild(relativeElement)

            hoverOverlayElement = document.createElement('div')
            hoverOverlayElement.className = 'hover-overlay-element'
            hoverOverlayElement.textContent = 'hoverOverlayElement'
            relativeElement.appendChild(hoverOverlayElement)
        })
        afterEach(() => {
            relativeElement.remove()
            document.scrollingElement!.scrollTop = 0
        })

        describe('with the scrolling element being document', () => {
            describe('not scrolled', () => {
                beforeEach(() => {
                    relativeElement.style.height = '600px'
                })

                it('should return a position above the given target if the overlay fits above', () => {
                    const target = createTarget({ left: 100, top: 200 })
                    const position = calculateOverlayPosition({ relativeElement, target, hoverOverlayElement })
                    applyOffsets(hoverOverlayElement, position)
                    assert.deepStrictEqual(position, { left: 100, top: 50 })
                })

                it('should return a position below the a given target if the overlay does not fit above', () => {
                    const target = createTarget({ left: 100, top: 100 })
                    const position = calculateOverlayPosition({ relativeElement, target, hoverOverlayElement })
                    applyOffsets(hoverOverlayElement, position)
                    assert.deepStrictEqual(position, { left: 100, top: 116 })
                })
            })
            describe('scrolled', () => {
                beforeEach(() => {
                    relativeElement.style.height = '3000px'
                    document.scrollingElement!.scrollTop = 400
                })

                it('should return a position above the given target if the overlay fits above', () => {
                    const target = createTarget({ left: 100, top: 600 })
                    const position = calculateOverlayPosition({ relativeElement, target, hoverOverlayElement })
                    applyOffsets(hoverOverlayElement, position)
                    assert.deepStrictEqual(position, { left: 100, top: 450 })
                })

                it('should return a position below the a given target if the overlay does not fit above', () => {
                    const target = createTarget({ left: 100, top: 450 })
                    const position = calculateOverlayPosition({ relativeElement, target, hoverOverlayElement })
                    applyOffsets(hoverOverlayElement, position)
                    assert.deepStrictEqual(position, { left: 100, top: 466 })
                })
            })
        })

        describe('with the scrolling element being the relativeElement', () => {
            beforeEach(() => {
                relativeElement.style.height = '600px'
                relativeElement.style.overflow = 'auto'
            })
            describe('not scrolled', () => {
                beforeEach(() => {
                    const content = document.createElement('div')
                    content.style.height = '500px'
                    content.style.width = '700px'
                    relativeElement.appendChild(content)
                })

                it('should return a position above the given target if the overlay fits above', () => {
                    const target = createTarget({ left: 100, top: 200 })
                    const position = calculateOverlayPosition({ relativeElement, target, hoverOverlayElement })
                    applyOffsets(hoverOverlayElement, position)
                    assert.deepStrictEqual(position, { left: 100, top: 50 })
                })

                it('should return a position below the a given target if the overlay does not fit above', () => {
                    const target = createTarget({ left: 100, top: 140 })
                    const position = calculateOverlayPosition({ relativeElement, target, hoverOverlayElement })
                    applyOffsets(hoverOverlayElement, position)
                    assert.deepStrictEqual(position, { left: 100, top: 156 })
                })
            })
            describe('scrolled', () => {
                beforeEach(() => {
                    const content = document.createElement('div')
                    content.style.height = '3000px'
                    content.style.width = '3000px'
                    relativeElement.appendChild(content)

                    relativeElement.scrollTop = 400
                    relativeElement.scrollLeft = 200
                })

                it('should return a position above the given target if the overlay fits above', () => {
                    const target = createTarget({ left: 300, top: 600 })
                    const position = calculateOverlayPosition({ relativeElement, target, hoverOverlayElement })
                    applyOffsets(hoverOverlayElement, position)
                    assert.deepStrictEqual(position, { left: 300, top: 450 })
                })

                it('should return a position below the a given target if the overlay does not fit above', () => {
                    const target = createTarget({ left: 300, top: 450 })
                    const position = calculateOverlayPosition({ relativeElement, target, hoverOverlayElement })
                    applyOffsets(hoverOverlayElement, position)
                    assert.deepStrictEqual(position, { left: 300, top: 466 })
                })
            })
        })
    })
})
