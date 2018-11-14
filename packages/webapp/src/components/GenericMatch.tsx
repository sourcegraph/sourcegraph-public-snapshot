import FileIcon from 'mdi-react/FileIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { ResultContainer } from './ResultContainer'

interface Props {
    result: GQL.IGenericSearchResult
}

export class GenericMatch extends React.Component<Props> {
    constructor(props: Props) {
        super(props)
    }

    private renderTitle = () => <div dangerouslySetInnerHTML={{ __html: this.props.result.label }} />
    public renderBody = () => (
        <>
            {this.props.result.results!.map(item => (
                <Link to={item.url} className="file-match__item">
                    <code dangerouslySetInnerHTML={{ __html: item.body }} />
                </Link>
            ))}
        </>
    )

    public render(): JSX.Element {
        return (
            <ResultContainer
                stringIcon={this.props.result.icon}
                icon={FileIcon}
                title={this.renderTitle()}
                expandedChildren={this.renderBody()}
                collapsedChildren={this.renderBody()}
            />
        )
    }
}
