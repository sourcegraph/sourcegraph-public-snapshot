import * as React from 'react'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { ToggleButton } from '../components/ToggleButton'

interface Props {
    onChange: (state: boolean) => void
    enabled: boolean
}

export class DotStarButton extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return <ToggleButton onChange={this.props.onChange} label=".*" enabled={this.props.enabled} />
    }
}
