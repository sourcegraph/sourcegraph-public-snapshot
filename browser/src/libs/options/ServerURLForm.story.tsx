import * as React from 'react'

import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'

import '../../app.scss'

import { interval, Subscription } from 'rxjs'
import { ConnectionErrors, ServerURLForm, ServerURLFormProps } from './ServerURLForm'

class Container extends React.Component<{}, { value: string; status: ServerURLFormProps['status'] }> {
    public state = { value: 'https://sourcegraph.com', status: 'connected' as ServerURLFormProps['status'] }

    public render(): React.ReactNode {
        return (
            <div style={{ maxWidth: 400 }}>
                <ServerURLForm
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
        let status: ServerURLFormProps['status'] = 'connected'
        let error: ServerURLFormProps['connectionError']
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
                <ServerURLForm
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

storiesOf('Options - ServerURLForm', module)
    .add('Interactive', () => <Container />)
    .add('Cycling Status', () => <CyclingStatus />)
    .add('Error Status', () => (
        <div style={{ maxWidth: 400, padding: '1rem' }}>
            <ServerURLForm
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
