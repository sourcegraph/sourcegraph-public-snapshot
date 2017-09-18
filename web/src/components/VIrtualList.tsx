import * as React from 'react'
import VisibilitySensor from 'react-visibility-sensor'

interface Props {
    initItemsToShow?: number
    items: any[]
}

interface State {
    itemsToShow: number
}

export class VirtualList extends React.Component<Props, State> {
    public state: State = {
        itemsToShow: 5
    }

    constructor(props: Props) {
        super(props)
        if (props.initItemsToShow) {
            this.state.itemsToShow = props.initItemsToShow
        }
    }

    public componentWillReceiveProps(nextProps: Props): void {
        if (this.props.items !== nextProps.items) {
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
            {
                this.props.items.slice(0, this.state.itemsToShow).map((item, i) =>
                    <VisibilitySensor key={i} onChange={isVisible => this.onChangeVisibility(isVisible, i)} partialVisibility={true}>
                        {item}
                    </VisibilitySensor>
                )
            }
            </div>
        )
    }
}
