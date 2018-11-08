import { storiesOf } from '@storybook/react'
import * as React from 'react'

import { action } from '@storybook/addon-actions'
import { JSONEditor } from '../src/shared/components/JSONEditor'

class Container extends React.Component<{}, { value: any }> {
    public state = { value: { hello: 'World', nested: { object: true } } }

    public render(): React.ReactNode {
        return (
            <div style={{ maxWidth: 400 }}>
                <JSONEditor value={this.state.value} onChange={this.onChange} />
            </div>
        )
    }

    private onChange = (value: any) => {
        this.setState({ value })

        action('JSON value changed')(value)
    }
}

storiesOf('JSONEditor', module).add('Default', () => <Container />)
