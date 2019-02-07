import * as H from 'history'
import * as React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { SearchQueryForm } from './SearchQueryForm'

interface Props extends SettingsCascadeProps {
    authenticatedUser: GQL.IUser | null
    location: H.Location
    history: H.History
    isLightTheme: boolean
    onThemeChange: () => void
}

/**
 * The search page
 */
export class SearchPage extends React.Component<Props, {}> {
    constructor(props: Props) {
        super(props)
    }

    public render(): JSX.Element | null {
        return (
            <div className="search-page">
                <img
                    className="search-page__logo"
                    src={
                        `${window.context.assetsRoot}/img/sourcegraph` +
                        (this.props.isLightTheme ? '-light' : '') +
                        '-head-logo.svg'
                    }
                />
                <SearchQueryForm {...this.props} />
            </div>
        )
    }
}
