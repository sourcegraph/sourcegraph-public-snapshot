import LinkIcon from 'mdi-react/LinkIcon'
import React from 'react'
import { Link } from '../../../shared/src/components/Link'
import { QuickLink } from '../schema/settings.schema'

interface Props {
    quickLinks: QuickLink[]
}

export class QuickLinks extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return this.props.quickLinks.length > 0 ? (
            <>
                {this.props.quickLinks.map((quickLink, i) => (
                    <small className="quicklink text-nowrap mr-2" key={i}>
                        <Link to={quickLink.url} data-tooltip={quickLink.description}>
                            <LinkIcon className="icon-inline pr-1" />
                            {quickLink.name}
                        </Link>
                    </small>
                ))}
            </>
        ) : null
    }
}
