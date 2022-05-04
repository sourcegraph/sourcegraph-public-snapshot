import { storiesOf } from '@storybook/react'

import { WebStory } from '../components/WebStory'

import { TosConsentModal } from './TosConsentModal'

const { add } = storiesOf('web/auth/TosConsentModal', module)

add('standard', () => <WebStory>{() => <TosConsentModal afterTosAccepted={() => {}} />}</WebStory>)
