import classNames from 'classnames'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { debounceTime, distinctUntilChanged } from 'rxjs/operators'

import styles from './Resizable.module.scss'

interface Props<C extends React.ReactElement = React.ReactElement> {
    className?: string
    handleClassName?: string

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
}

const containerClassNameMap: Record<Props['handlePosition'], string> = {
    top: styles.resizableTop,
    left: styles.resizableLeft,
    right: '',
}

const handleClassNameMap: Record<Props['handlePosition'], string> = {
    top: styles.handleTop,
    left: styles.handleLeft,
    right: styles.handleRight,
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

        this.state = {
            resizing: false,
            size: this.getSize(),
        }
    }

    private getSize(): number {
        const storedSize = localStorage.getItem(`${Resizable.STORAGE_KEY_PREFIX}${this.props.storageKey}`)
        if (storedSize !== null) {
            const sizeNumber = parseInt(storedSize, 10)
            if (sizeNumber >= 0) {
                return sizeNumber
            }
        }
        return this.props.defaultSize
    }

    private setSize(size: number): void {
        localStorage.setItem(`${Resizable.STORAGE_KEY_PREFIX}${this.props.storageKey}`, String(size))
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.sizeUpdates.pipe(distinctUntilChanged(), debounceTime(250)).subscribe(size => this.setSize(size))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        const { size, resizing } = this.state
        const { element, handlePosition, handleClassName, className } = this.props

        return (
            <div
                // eslint-disable-next-line react/forbid-dom-props
                style={{ [isHorizontal(handlePosition) ? 'width' : 'height']: `${size}px` }}
                className={classNames(styles.resizable, containerClassNameMap[handlePosition], className)}
                ref={this.setContainerRef}
            >
                {element}
                <div
                    role="presentation"
                    className={classNames(
                        styles.handle,
                        handleClassNameMap[handlePosition],
                        resizing && styles.handleResizing,
                        handleClassName
                    )}
                    onMouseDown={this.onMouseDown}
                />
            </div>
        )
    }

    private setContainerRef = (event: HTMLElement | null): void => {
        this.containerRef = event
    }

    private onMouseDown = (event: React.MouseEvent<HTMLDivElement>): void => {
        event.preventDefault()
        if (!this.state.resizing) {
            this.setState({ resizing: true })

            const onMouseMove = (event: MouseEvent): void => {
                event.preventDefault()
                if (this.state.resizing && this.containerRef) {
                    let size = isHorizontal(this.props.handlePosition)
                        ? this.props.handlePosition === 'right'
                            ? event.pageX - this.containerRef.getBoundingClientRect().left
                            : this.containerRef.getBoundingClientRect().right - event.pageX
                        : this.containerRef.getBoundingClientRect().bottom - event.pageY
                    if (event.shiftKey) {
                        size = Math.ceil(size / 20) * 20
                    }
                    this.setState({ size })
                    this.sizeUpdates.next(size)
                }
            }

            const onMouseUp = (event: Event): void => {
                event.preventDefault()
                if (this.state.resizing) {
                    this.setState({ resizing: false })
                    if (event.currentTarget) {
                        event.currentTarget.removeEventListener('mouseup', onMouseUp)
                        event.currentTarget.removeEventListener('mousemove', onMouseMove as EventListener)
                    }
                }
            }

            event.currentTarget.ownerDocument.addEventListener('mousemove', onMouseMove)
            event.currentTarget.ownerDocument.addEventListener('mouseup', onMouseUp)
        }
    }
}
