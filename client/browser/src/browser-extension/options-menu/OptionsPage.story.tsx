import * as React from 'react'
import { storiesOf } from '@storybook/react'
<<<<<<< HEAD:client/browser/src/browser-extension/options-page/ServerUrlForm.story.tsx
import { ConnectionErrors, ServerUrlForm, ServerUrlFormProps } from './ServerUrlForm'
import { BrandedStory } from '../../../../branded/src/components/BrandedStory'
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
    .addDecorator(story => <BrandedStory styles={optionsStyles}>{() => story()}</BrandedStory>)
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
=======
import { OptionsPage } from './OptionsPage'
import optionsStyles from '../../options.scss'

storiesOf('browser/Options/OptionsPage', module)
    .addDecorator(story => (
        <>
            <style>{optionsStyles}</style>
            <div>{story()}</div>
        </>
    ))
    .add('Default', () => <OptionsPage version="0.0.0" />)
>>>>>>> 60538b4f66... massive exploratory commit:client/browser/src/browser-extension/options-menu/OptionsPage.story.tsx
