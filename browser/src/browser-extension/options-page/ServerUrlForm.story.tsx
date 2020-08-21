import * as React from 'react'
import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import { ConnectionErrors, ServerUrlForm, ServerUrlFormProps } from './ServerUrlForm'
import optionsStyles from '../../options.scss'

class Container extends React.Component<{}, { value: string; status: ServerUrlFormProps['status'] }> {
    public state = { value: 'https://sourcegraph.com', status: 'connected' as ServerUrlFormProps['status'] }

    public render(): React.ReactNode {
        return (
            <div style={{ maxWidth: 400 }}>
                <ServerUrlForm
                    {...this.state}
                    onChange={this.onChange}
                    onSubmit={this.onSubmit}
                    requestPermissions={() => undefined}
                    urlHasPermissions={true}
                />
            </div>
        )
    }

    private onChange = (value: string): void => {
        this.setState({ value })

        action('URL Changed')(value)
    }

    private onSubmit = (): void => {
        action('Form submitted')(this.state.value)
    }
}

storiesOf('browser/Options/ServerUrlForm', module)
    .addDecorator(story => (
        <>
            <style>{optionsStyles}</style>
            <div>{story()}</div>
        </>
    ))
    .add('Interactive', () => <Container />)
    .add('Error Status', () => (
        <div style={{ maxWidth: 400, padding: '1rem' }}>
            <ServerUrlForm
                value="https://sourcegraph.com"
                status="error"
                connectionError={ConnectionErrors.AuthError}
                onChange={action('Change')}
                onSubmit={action('Submit')}
                requestPermissions={() => undefined}
                urlHasPermissions={true}
            />
        </div>
    ))
