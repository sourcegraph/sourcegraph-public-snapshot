import * as React from 'react'
import VisibilitySensor from 'react-visibility-sensor'

interface Props<TItem, TExtraItemProps> {
    itemsToShow: number
    items: TItem[]

    /* Additional props passed to the render function. */
    itemProps: TExtraItemProps
    /* Function to render an item once it becomes visible. */
    renderItem: (item: TItem, additionalProps: TExtraItemProps) => JSX.Element
    /* Determines the list key of an item. Needed for stable react array rendering. */
    itemKey: (item: TItem) => string

    /**
     * Called when the user scrolled close to the bottom of the list.
     * The parent component should react to this by increasing `itemsToShow`.
     */
    onShowMoreItems: () => void

    /**
     * Notifies the parent component when an item either has become or is not longer visible.
     */
    onVisibilityChange?: (isVisible: boolean, index: number) => void

    onRef?: (reference: HTMLElement | null) => void

    /**
     * Element to use as a viewport when checking visibility. If undefined,
     * the browser window will be used as a viewport.
     */
    containment?: HTMLElement

    className?: string
}

interface State {}

export class VirtualList<TItem, TExtraItemProps = undefined> extends React.PureComponent<
    Props<TItem, TExtraItemProps>,
    State
> {
    public onChangeVisibility = (isVisible: boolean, index: number): void => {
        if (isVisible && index >= this.props.itemsToShow - 2) {
            this.props.onShowMoreItems()
        }

        if (this.props.onVisibilityChange) {
            this.props.onVisibilityChange(isVisible, index)
        }
    }

    public render(): JSX.Element | null {
        return (
            <div className={this.props.className} ref={this.props.onRef}>
                {this.props.items.slice(0, this.props.itemsToShow).map((item, index) => (
                    <VisibilitySensor
                        // eslint-disable-next-line react/jsx-no-bind
                        onChange={isVisible => this.onChangeVisibility(isVisible, index)}
                        key={this.props.itemKey(item)}
                        containment={this.props.containment}
                        partialVisibility={true}
                    >
                        {this.props.renderItem(item, this.props.itemProps)}
                    </VisibilitySensor>
                ))}
            </div>
        )
    }
}
