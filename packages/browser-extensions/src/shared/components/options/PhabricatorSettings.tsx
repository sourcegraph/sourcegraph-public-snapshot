import * as React from 'react'
import { PhabricatorMappings } from './PhabricatorMappings'

export class PhabricatorSettings extends React.Component<{}, {}> {
    public render(): JSX.Element | null {
        return (
            <div>
                <div className="options__section-subheader">
                    Phabricator mappings{' '}
                    <a
                        href="https://docs.sourcegraph.com/integration/phabricator"
                        target="_blank"
                        // tslint:disable-next-line
                        onClick={e => e.stopPropagation()}
                        className="options__alert-link"
                    >
                        (Learn more)
                    </a>
                </div>
                <PhabricatorMappings />
            </div>
        )
    }
}
