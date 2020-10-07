import { OptionsPage } from './OptionsPage'
import optionsStyles from '../../options.scss'
import { Observable, of } from 'rxjs'
import { action } from '@storybook/addon-actions'
import { boolean, text } from '@storybook/addon-knobs'

const validateSourcegraphUrl = (url: string): Observable<string | undefined> => of(undefined)
const invalidSourcegraphUrl = (url: string): Observable<string | undefined> => of('Arbitrary error string')

storiesOf('browser/Options/OptionsPage', module)
    .addDecorator(story => (
        <>
            <style>{optionsStyles}</style>
            <div>{story()}</div>
        </>
    ))
    .add('Default', () => (
        <OptionsPage
            version={text('version', '0.0.0')}
            isCurrentRepositoryPrivate={boolean('isCurrentRepositoryPrivate', false)}
            validateSourcegraphUrl={validateSourcegraphUrl}
            onToggleActivated={action('onToggleActivated')}
            isActivated={true}
            sourcegraphUrl={text('sourcegraphUrl', 'https://sourcegraph.com')}
            isFullPage={true}
            showPrivateRepositoryAlert={boolean('showPrivateRepositoryAlert', false)}
        />
    ))
    .add('Interactive', () => {
        const [isActivated, setIsActivated] = useState(false)
        return (
            <OptionsPage
                version={text('version', '0.0.0')}
                isActivated={isActivated}
                onToggleActivated={setIsActivated}
                validateSourcegraphUrl={validateSourcegraphUrl}
                sourcegraphUrl={text('sourcegraphUrl', 'https://sourcegraph.com')}
                isCurrentRepositoryPrivate={boolean('isCurrentRepositoryPrivate', false)}
                isFullPage={true}
            />
        )
    })
    .add('URL validation error', () => {
        const [isActivated, setIsActivated] = useState(false)
        return (
            <OptionsPage
                version={text('version', '0.0.0')}
                isActivated={isActivated}
                onToggleActivated={setIsActivated}
                validateSourcegraphUrl={invalidSourcegraphUrl}
                sourcegraphUrl={text('sourcegraphUrl', 'https://not-sourcegraph.com')}
                isCurrentRepositoryPrivate={boolean('isCurrentRepositoryPrivate', false)}
                isFullPage={true}
            />
        )
    })
    .add('With permission alert', () => (
        <OptionsPage
            version={text('version', '0.0.0')}
            isCurrentRepositoryPrivate={boolean('isCurrentRepositoryPrivate', false)}
            validateSourcegraphUrl={validateSourcegraphUrl}
            onToggleActivated={action('onToggleActivated')}
            isActivated={true}
            sourcegraphUrl={text('sourcegraphUrl', 'https://sourcegraph.com')}
            isFullPage={true}
            permissionAlert={{ name: 'GitLab.com' }}
        />
    ))
