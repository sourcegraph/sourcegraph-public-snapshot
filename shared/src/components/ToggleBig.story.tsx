import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import React from 'react'
import { ToggleBig } from './ToggleBig'
import webStyles from '../../../web/src/main.scss'

const onToggle = action('onToggle')

const { add } = storiesOf('shared/ToggleBig', module).addDecorator(story => (
    <>
        <div>{story()}</div>
        <style>{webStyles}</style>
    </>
))

add('Interactive', () => {
    interface State {
        value?: boolean
    }
    class ToggleBigInteractive extends React.Component<{}, State> {
        public state: State = {
            value: false,
        }
        public render(): JSX.Element | null {
            return (
                <div className="d-flex align-items-center">
                    <ToggleBig value={this.state.value} onToggle={this.onToggle} title="Hello" className="mr-2" /> Value
                    is {String(this.state.value)}
                </div>
            )
        }
        private onToggle = (value: boolean): void => this.setState({ value }, (): void => onToggle(value))
    }
    return <ToggleBigInteractive />
})

add('On', () => <ToggleBig value={true} onToggle={onToggle} />)

add('Off', () => <ToggleBig value={false} onToggle={onToggle} />)

add('Disabled & on', () => <ToggleBig value={true} disabled={true} onToggle={onToggle} />)

add('Disabled & off', () => <ToggleBig value={false} disabled={true} onToggle={onToggle} />)
