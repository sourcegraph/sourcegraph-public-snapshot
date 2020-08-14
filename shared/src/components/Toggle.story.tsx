import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import React from 'react'
import { Toggle } from './Toggle'
import webStyles from '../../../web/src/main.scss'

const onToggle = action('onToggle')

const { add } = storiesOf('shared/Toggle', module).addDecorator(story => (
    <>
        <div>{story()}</div>
        <style>{webStyles}</style>
    </>
))

add('Interactive', () => {
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

add('On', () => <Toggle value={true} onToggle={onToggle} />)

add('Off', () => <Toggle value={false} onToggle={onToggle} />)

add('Disabled & on', () => <Toggle value={true} disabled={true} onToggle={onToggle} />)

add('Disabled & off', () => <Toggle value={false} disabled={true} onToggle={onToggle} />)
