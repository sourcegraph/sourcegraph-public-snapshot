import * as H from 'history'
import * as React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import { LinkOrSpan } from '../../../../../shared/src/components/LinkOrSpan'
import { FileDiffNode } from '../../../components/diff/FileDiffNode'
import { ThemeProps } from '../../../../../shared/src/theme'

export interface FileDiffTabNodeProps extends ThemeProps {
    node: GQL.IChangesetPlan | GQL.IExternalChangeset
    persistLines?: boolean
    history: H.History
    location: H.Location
}

export const FileDiffTabNode: React.FunctionComponent<FileDiffTabNodeProps> = ({
    node,
    persistLines,
    history,
    location,
    isLightTheme,
}) =>
    node.diff && (
        <div>
            <h3>
                <SourcePullIcon className="icon-inline mr-2" />{' '}
                <LinkOrSpan to={node.repository.url}>{node.repository.name}</LinkOrSpan>
            </h3>
            {((): (GQL.IPreviewFileDiff | GQL.IFileDiff)[] => node.diff.fileDiffs.nodes)().map(fileDiffNode => (
                <FileDiffNode
                    isLightTheme={isLightTheme}
                    node={fileDiffNode}
                    lineNumbers={true}
                    location={location}
                    history={history}
                    persistLines={persistLines}
                    key={fileDiffNode.internalID}
                ></FileDiffNode>
            ))}
        </div>
    )
