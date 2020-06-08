import * as React from 'react'
import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import { interval, Subscription } from 'rxjs'
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

class CyclingStatus extends React.Component<{}, { step: number }> {
    public state = { step: 0 }
    private subscription = new Subscription()

    private onChange = action('Input onChange fired')
    private onSubmit = action('Form onSubmit fired')

    public componentDidMount(): void {
        this.subscription.add(
            interval(1000).subscribe(() => {
                this.setState(({ step }) => ({ step: (step + 1) % 4 }))
            })
        )
    }

    public componentWillUnmount(): void {
        this.subscription.unsubscribe()
    }

    public render(): React.ReactNode {
        let status: ServerUrlFormProps['status'] = 'connected'
        let error: ServerUrlFormProps['connectionError']
        let isUpdating: boolean | undefined

        if (this.state.step === 1) {
            status = 'connecting'
        } else if (this.state.step === 2) {
            status = 'error'
            error = ConnectionErrors.AuthError
        } else if (this.state.step === 3) {
            isUpdating = true
        }

        return (
            <div style={{ maxWidth: 400 }}>
                <ServerUrlForm
                    value="https://sourcegraph.com"
                    status={status}
                    connectionError={error}
                    onChange={this.onChange}
                    onSubmit={this.onSubmit}
                    overrideUpdatingState={isUpdating}
                    requestPermissions={() => undefined}
                    urlHasPermissions={true}
                />
            </div>
        )
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
    .add('Cycling Status', () => <CyclingStatus />)
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
