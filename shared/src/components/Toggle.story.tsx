import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import React from 'react'
import { Toggle } from './Toggle'
import './Toggle.scss'

const onToggle = action('onToggle')

const { add } = storiesOf('Toggle', module)

add('interactive', () => {
    interface State {
        value?: boolean
    }
    class ToggleInteractive extends React.Component<{}, State> {
        public state: State = {}
        public render(): JSX.Element | null {
            return (
                <div className="d-flex align-items-center">
                    <Toggle value={this.state.value} onToggle={this.onToggle} title="Hello" className="mr-2" /> Value is{' '}
                    {String(this.state.value)}
                </div>
            )
        }
        private onToggle = (value: boolean): void => this.setState({ value }, (): void => onToggle(value))
    }
    return <ToggleInteractive />
})

add('on', () => <Toggle value={true} onToggle={onToggle} />)

add('off', () => <Toggle value={false} onToggle={onToggle} />)

add('disabled/on', () => <Toggle value={true} disabled={true} onToggle={onToggle} />)

add('disabled/off', () => <Toggle value={false} disabled={true} onToggle={onToggle} />)
