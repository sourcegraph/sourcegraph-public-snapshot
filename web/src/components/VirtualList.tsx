import * as React from 'react'
import VisibilitySensor from 'react-visibility-sensor'

interface Props {
    itemsToShow: number
    items: JSX.Element[]

    /**
     * Called when the user scrolled close to the bottom of the list.
     * The parent component should react to this by increasing `itemsToShow`.
     */
    onShowMoreItems: () => void

    /**
     * Notifies the parent component when an item either has become or is not longer visible.
     */
    onVisibilityChange?: (isVisible: boolean, index: number) => void

    onRef?: (ref: HTMLElement | null) => void

    className?: string
}

interface State {}

export class VirtualList extends React.PureComponent<Props, State> {
    public onChangeVisibility = (isVisible: boolean, i: number): void => {
        if (isVisible && i >= this.props.itemsToShow - 2) {
            this.props.onShowMoreItems()
        }

        if (this.props.onVisibilityChange) {
            this.props.onVisibilityChange(isVisible, i)
        }
    }

    public render(): JSX.Element | null {
        return (
            <div className={this.props.className} ref={this.props.onRef}>
                {this.props.items.slice(0, this.props.itemsToShow).map((item, i) => (
                    <VisibilitySensor
                        key={item.key}
                        // tslint:disable-next-line:jsx-no-lambda
                        onChange={(isVisible: boolean) => this.onChangeVisibility(isVisible, i)}
                        partialVisibility={true}
                    >
                        {item}
                    </VisibilitySensor>
                ))}
            </div>
        )
    }
}
