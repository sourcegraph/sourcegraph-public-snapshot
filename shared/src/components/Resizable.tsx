import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { debounceTime, distinctUntilChanged } from 'rxjs/operators'

interface Props<C extends React.ReactElement<any> = React.ReactElement<any>> {
    className?: string

    /**
     * Where the resize handle is (which also determines the axis along which the element can be
     * resized).
     */
    handlePosition: 'right' | 'left' | 'top'

    /**
     * Persist and restore the size of the element using this key.
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

    /** Element styles. */
    style?: React.CSSProperties
}

const isHorizontal = (handlePosition: Props['handlePosition']) =>
    handlePosition === 'right' || handlePosition === 'left'

interface State {
    resizing: boolean
    size: number
}

/**
 * Wraps an item in a flexbox and makes it resizable.
 */
export class Resizable<C extends React.ReactElement<any>> extends React.PureComponent<Props<C>, State> {
    private static STORAGE_KEY_PREFIX = 'Resizable:'

    private sizeUpdates = new Subject<number>()
    private subscriptions = new Subscription()

    private containerRef: HTMLElement | null = null

    constructor(props: Props<C>) {
        super(props)

        this.state = {
            resizing: false,
            size: this.getSize(),
        }
    }

    private getSize(): number {
        const v = localStorage.getItem(`${Resizable.STORAGE_KEY_PREFIX}${this.props.storageKey}`)
        if (v !== null) {
            const n = parseInt(v, 10)
            if (n >= 0) {
                return n
            }
        }
        return this.props.defaultSize
    }

    private setSize(size: number): void {
        localStorage.setItem(`${Resizable.STORAGE_KEY_PREFIX}${this.props.storageKey}`, String(size))
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.sizeUpdates
                .pipe(
                    distinctUntilChanged(),
                    debounceTime(250)
                )
                .subscribe(size => this.setSize(size))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        return (
            <div
                className={`resizable resizable--${this.props.handlePosition} ${this.props.className || ''}`}
                ref={this.setContainerRef}
                // tslint:disable-next-line jsx-ban-props
                style={{
                    ...this.props.style,
                    [isHorizontal(this.props.handlePosition) ? 'width' : 'height']: `${this.state.size}px`,
                }}
            >
                <div
                    className={`resizable__ghost ${this.state.resizing ? 'resizable__ghost--resizing' : ''}`}
                    onMouseMove={this.onMouseMove}
                    onMouseUp={this.onMouseUp}
                />
                {this.props.element}
                <div
                    className={`resizable__handle resizable__handle--${this.props.handlePosition} ${
                        this.state.resizing ? 'resizable__handle--resizing' : ''
                    }`}
                    onMouseDown={this.onMouseDown}
                />
            </div>
        )
    }

    private setContainerRef = (e: HTMLElement | null) => (this.containerRef = e)

    private onMouseDown = (e: React.MouseEvent<HTMLDivElement>) => {
        e.preventDefault()
        if (!this.state.resizing) {
            this.setState({ resizing: true })
        }
    }

    private onMouseUp = (e: React.MouseEvent<HTMLDivElement>) => {
        e.preventDefault()
        if (this.state.resizing) {
            this.setState({ resizing: false })
        }
    }

    private onMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
        e.preventDefault()
        if (this.state.resizing && this.containerRef) {
            let size = isHorizontal(this.props.handlePosition)
                ? this.props.handlePosition === 'right'
                    ? e.pageX - this.containerRef.getBoundingClientRect().left
                    : this.containerRef.getBoundingClientRect().right - e.pageX
                : this.containerRef.getBoundingClientRect().bottom - e.pageY
            if (e.shiftKey) {
                size = Math.ceil(size / 20) * 20
            }
            this.setState({ size })
            this.sizeUpdates.next(size)
        }
    }
}
