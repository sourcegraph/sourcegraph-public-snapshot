import * as React from 'react'
import { OptionsConfiguration } from './OptionsConfiguration'

export class OptionsDashboard extends React.Component<any, {}> {
    public render(): JSX.Element {
        return (
            <div className="site-admin-area area">
                <div className="area__content">
                    <OptionsConfiguration />
                </div>
            </div>
        )
    }
}
