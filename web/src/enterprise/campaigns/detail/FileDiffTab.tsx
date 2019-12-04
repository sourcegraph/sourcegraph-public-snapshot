import * as H from 'history'
import * as React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import { LinkOrSpan } from '../../../../../shared/src/components/LinkOrSpan'
import { FileDiffNode } from '../../../components/diff/FileDiffNode'
import { ThemeProps } from '../../../../../shared/src/theme'

interface Props extends ThemeProps {
    nodes: (GQL.IExternalChangeset | GQL.IChangesetPlan)[]
    persistLines?: boolean
    history: H.History
    location: H.Location
}

export const FileDiffTab: React.FunctionComponent<Props> = ({
    nodes,
    persistLines = true,
    isLightTheme,
    history,
    location,
}) => (
    <>
        {nodes.map(
            (changesetNode, i) =>
                changesetNode.diff && (
                    <div key={i}>
                        <h3>
                            <SourcePullIcon className="icon-inline mr-2" />{' '}
                            <LinkOrSpan to={changesetNode.repository.url}>{changesetNode.repository.name}</LinkOrSpan>
                        </h3>
                        {((): (GQL.IPreviewFileDiff | GQL.IFileDiff)[] => changesetNode.diff.fileDiffs.nodes)().map(
                            fileDiffNode => (
                                <FileDiffNode
                                    isLightTheme={isLightTheme}
                                    node={fileDiffNode}
                                    lineNumbers={true}
                                    location={location}
                                    history={history}
                                    persistLines={persistLines}
                                    key={fileDiffNode.internalID}
                                ></FileDiffNode>
                            )
                        )}
                    </div>
                )
        )}
    </>
)
