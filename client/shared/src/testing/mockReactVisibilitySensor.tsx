import React from 'react'

import type _VisibilitySensor from 'react-visibility-sensor'
import { vi } from 'vitest'

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

function mockVisibilitySensor({ children, onChange }: VisibilitySensorPropsType): typeof _VisibilitySensor {
    return (
        <>
            <MockVisibilitySensor onChange={onChange}>{children}</MockVisibilitySensor>
        </>
    )
}

vi.mock('react-visibility-sensor', () => ({ default: mockVisibilitySensor }))
