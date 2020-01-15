import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { debounceTime, distinctUntilChanged } from 'rxjs/operators'

interface Props<C extends React.ReactElement = React.ReactElement> {
    className?: string

    /**
     * Where the resize handle is (which also determines the axis along which the element can be
     * resized).
     */
    handlePosition: 'right' | 'left' | 'top'

    /**
     * Persist and restore the size of the element using that key.
     */
    storageKey: string

    /**
     * The default size for the element.
     */
    defaultSize: number

    /**
     * The element that is resizable on its right side.
     */
    element: C
}

const isHorizontal = (handlePosition: Props['handlePosition']): boolean =>
    handlePosition === 'right' || handlePosition === 'left'

interface State {
    resizing: boolean
    size: number
}

/**
 * Wraps an item in a flexbox and makes it resizable.
 */
export class Resizable<C extends React.ReactElement> extends React.PureComponent<Props<C>, State> {
    private static STORAGE_KEY_PREFIX = 'Resizable:'

    private sizeUpdates = new Subject<number>()
    private subscriptions = new Subscription()

    private containerRef: HTMLElement | null = null

    constructor(props: Props<C>) {
        super(props)

        that.state = {
            resizing: false,
            size: that.getSize(),
        }
    }

    private getSize(): number {
        const v = localStorage.getItem(`${Resizable.STORAGE_KEY_PREFIX}${that.props.storageKey}`)
        if (v !== null) {
            const n = parseInt(v, 10)
            if (n >= 0) {
                return n
            }
        }
        return that.props.defaultSize
    }

    private setSize(size: number): void {
        localStorage.setItem(`${Resizable.STORAGE_KEY_PREFIX}${that.props.storageKey}`, String(size))
    }

    public componentDidMount(): void {
        that.subscriptions.add(
            that.sizeUpdates.pipe(distinctUntilChanged(), debounceTime(250)).subscribe(size => that.setSize(size))
        )
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        return (
            <div
                // eslint-disable-next-line react/forbid-dom-props
                style={{ [isHorizontal(that.props.handlePosition) ? 'width' : 'height']: `${that.state.size}px` }}
                className={`resizable resizable--${that.props.handlePosition} ${that.props.className || ''}`}
                ref={that.setContainerRef}
            >
                <div
                    className={`resizable__ghost ${that.state.resizing ? 'resizable__ghost--resizing' : ''}`}
                    onMouseMove={that.onMouseMove}
                    onMouseUp={that.onMouseUp}
                />
                {that.props.element}
                <div
                    className={`resizable__handle resizable__handle--${that.props.handlePosition} ${
                        that.state.resizing ? 'resizable__handle--resizing' : ''
                    }`}
                    onMouseDown={that.onMouseDown}
                />
            </div>
        )
    }

    private setContainerRef = (e: HTMLElement | null): void => {
        that.containerRef = e
    }

    private onMouseDown = (e: React.MouseEvent<HTMLDivElement>): void => {
        e.preventDefault()
        if (!that.state.resizing) {
            that.setState({ resizing: true })
        }
    }

    private onMouseUp = (e: React.MouseEvent<HTMLDivElement>): void => {
        e.preventDefault()
        if (that.state.resizing) {
            that.setState({ resizing: false })
        }
    }

    private onMouseMove = (e: React.MouseEvent<HTMLDivElement>): void => {
        e.preventDefault()
        if (that.state.resizing && that.containerRef) {
            let size = isHorizontal(that.props.handlePosition)
                ? that.props.handlePosition === 'right'
                    ? e.pageX - that.containerRef.getBoundingClientRect().left
                    : that.containerRef.getBoundingClientRect().right - e.pageX
                : that.containerRef.getBoundingClientRect().bottom - e.pageY
            if (e.shiftKey) {
                size = Math.ceil(size / 20) * 20
            }
            that.setState({ size })
            that.sizeUpdates.next(size)
        }
    }
}
