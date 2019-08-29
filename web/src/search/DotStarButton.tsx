import * as React from 'react'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { isErrorLike } from '../../../shared/src/util/errors'
import { ToggleButton } from '../components/ToggleButton'

export class DotStarButton extends React.PureComponent<SettingsCascadeProps> {
    public render(): JSX.Element | null {
        const tpsf = this.props.settingsCascade.final
        const searchVersion = (tpsf && !isErrorLike(tpsf) && tpsf['search.version']) || 'V0'
        const props = {
            label: '.*',
            enabled: searchVersion !== 'V1',
        }
        return <ToggleButton {...props} />
    }
}
