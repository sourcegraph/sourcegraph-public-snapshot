import * as React from 'react'
import VisibilitySensor from 'react-visibility-sensor'

interface Props {
    initItemsToShow?: number

    /**
     * The unique identifier of the list. Used to e.g. reset the number of
     * shown items back to initItemsToShow when the items list changes.
     *
     * This is needed because simply knowing this.props.items !== nextProps.items
     * is not enough when e.g. the virtual list is a dynamic one where items are
     * constantly added.
     */
    listId: string
    items: JSX.Element[]
}

interface State {
    itemsToShow: number
}

export class VirtualList extends React.Component<Props, State> {
    public state: State = {
        itemsToShow: 5,
    }

    constructor(props: Props) {
        super(props)
        if (props.initItemsToShow) {
            this.state.itemsToShow = props.initItemsToShow
        }
    }

    public componentWillReceiveProps(nextProps: Props): void {
        if (this.props.listId !== nextProps.listId && this.props.items !== nextProps.items) {
            this.setState({ itemsToShow: nextProps.initItemsToShow || 5 })
        }
    }

    public onChangeVisibility = (isVisible: boolean, i: number): void => {
        if (isVisible && i >= this.state.itemsToShow - 2) {
            this.setState({ itemsToShow: this.state.itemsToShow + 3 })
        }
    }

    public render(): JSX.Element | null {
        return (
            <div>
                {this.props.items.slice(0, this.state.itemsToShow).map((item, i) => (
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
