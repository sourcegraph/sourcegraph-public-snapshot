import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import * as React from 'react'
import { HeroPage } from '../components/HeroPage'

/**
 * Regular expression to identify valid organization names.
 */
export const VALID_ORG_NAME_REGEXP = /^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$/

/**
 * Regular expression to identify valid username
 */
export const VALID_USERNAME_REGEXP = VALID_ORG_NAME_REGEXP

/**
 * Display a 404 to the user while providing the sidebar
 */
export class SettingsNotFoundPage extends React.Component {
    public render(): JSX.Element | null {
        return (
            <div className='settings'>
                <HeroPage icon={DirectionalSignIcon} title='404: Not Found' subtitle='Sorry, the requested URL was not found.' />
            </div>
        )
    }
}
