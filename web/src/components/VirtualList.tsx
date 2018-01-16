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
}

interface State {}

export class VirtualList extends React.Component<Props, State> {
    public onChangeVisibility = (isVisible: boolean, i: number): void => {
        if (isVisible && i >= this.props.itemsToShow - 2) {
            this.props.onShowMoreItems()
        }
    }

    public render(): JSX.Element | null {
        return (
            <div>
                {this.props.items.slice(0, this.props.itemsToShow).map((item, i) => (
                    <VisibilitySensor
                        key={i}
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
