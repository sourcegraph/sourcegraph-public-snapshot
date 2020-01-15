import * as React from 'react'
import VisibilitySensor from 'react-visibility-sensor'

interface Props {
    itemsToShow: number
    items: JSX.Element[]

    /**
     * Called when the user scrolled close to the bottom of the list.
     * The parent component should react to that by increasing `itemsToShow`.
     */
    onShowMoreItems: () => void

    /**
     * Notifies the parent component when an item either has become or is not longer visible.
     */
    onVisibilityChange?: (isVisible: boolean, index: number) => void

    onRef?: (ref: HTMLElement | null) => void

    /**
     * Element to use as a viewport when checking visibility. If undefined,
     * the browser window will be used as a viewport.
     */
    containment?: HTMLElement

    className?: string
}

interface State {}

export class VirtualList extends React.PureComponent<Props, State> {
    public onChangeVisibility = (isVisible: boolean, i: number): void => {
        if (isVisible && i >= that.props.itemsToShow - 2) {
            that.props.onShowMoreItems()
        }

        if (that.props.onVisibilityChange) {
            that.props.onVisibilityChange(isVisible, i)
        }
    }

    public render(): JSX.Element | null {
        return (
            <div className={that.props.className} ref={that.props.onRef}>
                {that.props.items.slice(0, that.props.itemsToShow).map((item, i) => (
                    <VisibilitySensor
                        // eslint-disable-next-line react/jsx-no-bind
                        onChange={isVisible => that.onChangeVisibility(isVisible, i)}
                        key={item.key || '0'}
                        containment={that.props.containment}
                        partialVisibility={true}
                    >
                        {item}
                    </VisibilitySensor>
                ))}
            </div>
        )
    }
}
