import H from 'history'
import { flatMap } from 'lodash'
import React from 'react'
import { Observable } from 'rxjs'
import { pluralize } from '../../../shared/src/util/strings'
import { Settings } from '../../../web/src/schema/settings.schema'
import * as GQL from '../graphql/schema'
import { isSettingsValid, SettingsCascadeProps } from '../settings/settings'
import { SymbolIcon } from '../symbols/SymbolIcon'
import { toPositionOrRangeHash } from '../util/url'
import { CodeExcerpt, FetchFileCtx } from './CodeExcerpt'
import { CodeExcerpt2 } from './CodeExcerpt2'
import { FileMatchChildren } from './FileMatchChildren'
import { mergeContext } from './FileMatchContext'
import { Link } from './Link'
import { RepoFileLink } from './RepoFileLink'
import { Props as ResultContainerProps, ResultContainer } from './ResultContainer'

const SUBSET_COUNT_KEY = 'fileMatchSubsetCount'

export type IFileMatch = Partial<Pick<GQL.IFileMatch, 'symbols' | 'limitHit'>> & {
    file: Pick<GQL.IFile, 'path' | 'url'> & { commit: Pick<GQL.IGitCommit, 'oid'> }
    repository: Pick<GQL.IRepository, 'name' | 'url'>
    lineMatches: ILineMatch[]
}

export type ILineMatch = Pick<GQL.ILineMatch, 'preview' | 'lineNumber' | 'offsetAndLengths' | 'limitHit'>

export interface IMatchItem {
    highlightRanges: {
        start: number
        highlightLength: number
    }[]
    preview: string
    line: number
}

interface Props extends SettingsCascadeProps {
    location: H.Location
    /**
     * The file match search result.
     */
    result: IFileMatch

    /**
     * Formatted repository name to be displayed in repository link. If not
     * provided, the default format will be displayed.
     */
    repoDisplayName?: string

    /**
     * The icon to show to the left of the title.
     */
    icon: React.ComponentType<{ className?: string }>

    /**
     * Called when the file's search result is selected.
     */
    onSelect: () => void

    /**
     * Whether this file should be rendered as expanded.
     */
    expanded: boolean

    /**
     * Whether or not to show all matches for this file, or only a subset.
     */
    showAllMatches: boolean

    isLightTheme: boolean

    allExpanded?: boolean

    fetchHighlightedFileLines: (ctx: FetchFileCtx, force?: boolean) => Observable<string[]>
}

export class FileMatch extends React.PureComponent<Props> {
    private subsetMatches = 10

    constructor(props: Props) {
        super(props)

        const subsetMatches = parseInt(localStorage.getItem(SUBSET_COUNT_KEY) || '', 10)
        if (!isNaN(subsetMatches)) {
            this.subsetMatches = subsetMatches
        }
    }

    public render(): React.ReactNode {
        const result = this.props.result
        const items: IMatchItem[] = this.props.result.lineMatches.map(m => ({
            highlightRanges: m.offsetAndLengths.map(offsetAndLength => ({
                start: offsetAndLength[0],
                highlightLength: offsetAndLength[1],
            })),
            preview: m.preview,
            line: m.lineNumber,
        }))

        const title = (
            <RepoFileLink
                repoName={result.repository.name}
                repoURL={result.repository.url}
                filePath={result.file.path}
                fileURL={result.file.url}
                repoDisplayName={this.props.repoDisplayName}
            />
        )

        let containerProps: ResultContainerProps

        const expandedChildren = (
            <FileMatchChildren
                {...this.props}
                items={items}
                result={result}
                allMatches={true}
                subsetMatches={this.subsetMatches}
            />
        )

        if (this.props.showAllMatches) {
            containerProps = {
                collapsible: true,
                defaultExpanded: this.props.expanded,
                icon: this.props.icon,
                title,
                expandedChildren,
                allExpanded: this.props.allExpanded,
            }
        } else {
            const len = items.length - this.subsetMatches
            containerProps = {
                collapsible: items.length > this.subsetMatches,
                defaultExpanded: this.props.expanded,
                icon: this.props.icon,
                title,
                collapsedChildren: (
                    <FileMatchChildren
                        {...this.props}
                        items={items}
                        result={result}
                        allMatches={true}
                        subsetMatches={this.subsetMatches}
                    />
                ),
                expandedChildren,
                collapseLabel: `Hide ${len} ${pluralize('match', len, 'matches')}`,
                expandLabel: `Show ${len} more ${pluralize('match', len, 'matches')}`,
                allExpanded: this.props.allExpanded,
            }
        }

        return <ResultContainer {...containerProps} />
    }
}
