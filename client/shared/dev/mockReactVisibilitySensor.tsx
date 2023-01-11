import React from 'react'

import _VisibilitySensor from 'react-visibility-sensor'

type VisibilitySensorPropsType = React.ComponentProps<typeof _VisibilitySensor>

export class MockVisibilitySensor extends React.Component<VisibilitySensorPropsType> {
    constructor(props: { onChange?: (isVisible: boolean) => void }) {
        super(props)
        if (props.onChange) {
            props.onChange(true)
        }
    }

    public render(): JSX.Element {
        return <>{this.props.children}</>
    }
}

jest.mock(
    'react-visibility-sensor',
    (): typeof _VisibilitySensor =>
        ({ children, onChange }: VisibilitySensorPropsType) =>
            (
                <>
                    <MockVisibilitySensor onChange={onChange}>{children}</MockVisibilitySensor>
                </>
            )
)
