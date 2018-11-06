import * as React from 'react'
import { PhabricatorMappings } from './PhabricatorMappings'

export class PhabricatorSettings extends React.Component<{}, {}> {
    public render(): JSX.Element | null {
        return (
            <div>
                <div className="options__section-subheader">
                    Phabricator mappings{' '}
                    <a
                        href="https://about.sourcegraph.com/docs/server/config/phabricator/#docs-content"
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
