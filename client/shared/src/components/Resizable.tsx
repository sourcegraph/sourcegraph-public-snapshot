import classnames from 'classnames'
import React, { useRef, useCallback, useEffect } from 'react'
import { useLocalStorage } from '../../../web/src/util/useLocalStorage'

interface Props {
    /**
     * The element that is resizable.
     */
    children: JSX.Element
    /**
     * The default size for the element.
     */
    width: number
    /**
     * Where the resize handle is (which also determines the axis along which the element can be
     * resized).
     */
    position: 'right' | 'left' | 'top'
}

export const Resizable: React.FunctionComponent<Props> = ({ children, width, position }) => {
    const [isResizable, setIsResizable] = React.useState(false)
    const [size, setSize] = useLocalStorage('sidebar-width', width)
    const reference = useRef<HTMLDivElement>(null)
    const onMouseUp = useCallback(() => setIsResizable(false), [])
    const onMouseDown = useCallback(() => setIsResizable(true), [])

    useEffect(() => {
        document.addEventListener('mousemove', onMouseMove)
        document.addEventListener('mouseup', onMouseUp)

        return () => {
            document.removeEventListener('mousemove', onMouseMove)
            document.removeEventListener('mouseup', onMouseUp)
        }
    }, [isResizable])

    const onMouseMove = (event: MouseEvent): void => {
        if (isResizable && reference.current) {
            if (position === 'left') {
                setSize(event.pageX - reference.current.getBoundingClientRect().left)
            } else if (position === 'right') {
                setSize(reference.current.getBoundingClientRect().right - event.pageX)
            } else if (position === 'top') {
                setSize(reference.current.getBoundingClientRect().bottom - event.pageY)
            }
        }
    }

    return (
        <div
            className={classnames({ 'flex-column-reverse': position === 'top' }, 'd-flex', 'w-100')}
            ref={reference}
            style={{ [position !== 'top' ? 'width' : 'height']: `${size}px` }}
        >
            {children}
            <div
                onMouseDown={onMouseDown}
                className="d-flex border-left"
                aria-hidden={true}
                style={{ borderLeft: '5px', borderColor: 'red', borderStyle: 'solid', cursor: 'col-resize' }}
            />
        </div>
    )
}
