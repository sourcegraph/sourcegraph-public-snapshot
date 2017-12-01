import * as H from 'history'
import { default as InspireTree } from 'inspire-tree'
import flatten from 'lodash/flatten'
import groupBy from 'lodash/groupBy'
import sortBy from 'lodash/sortBy'
import * as React from 'react'
import { Repo } from '../repo/index'
import { toPrettyBlobURL, toTreeURL } from '../util/url'

const InspireTreeDOM = require('inspire-tree-dom')

export interface Props extends Repo {
    history: H.History
    paths: string[]
    selectedPath: string
}

interface TreeNode {
    text: string
    /**
     * id for node, should be the dir/file path
     */
    id: string
    children: TreeNode[]
    /**
     * node initialization options
     */
    itree?: any
}

export class Tree extends React.PureComponent<Props, {}> {
    public paths: string[]
    private tree: InspireTree
    private nodes: TreeNode[]

    constructor(props: Props) {
        super(props)
        this.nodes = this.parseNodes(props.paths, props.selectedPath)
    }

    public componentDidMount(): void {
        this.tree = new InspireTree({
            data: this.nodes,
        })
        this.tree.on('node.selected', this.handleSelection)
        // tslint:disable-next-line:no-unused-expression
        new InspireTreeDOM(this.tree, {
            target: '.tree',
            tabindex: 1,
        })
    }

    public componentWillReceiveProps(nextProps: Props): void {
        const selectedPathOnDocument = document.querySelector(`[data-uid="${nextProps.selectedPath}"]`)
        if (this.props.paths !== nextProps.paths || !selectedPathOnDocument) {
            this.nodes = this.parseNodes(nextProps.paths, nextProps.selectedPath)
            this.tree.removeAll()
            this.tree.addNodes(this.nodes)
        }
    }

    public render(): JSX.Element | null {
        return <div className="tree" />
    }

    private handleSelection = (node: TreeNode, isLoad: boolean) => {
        if (node.children.length === 0) {
            const rev = this.props.rev && this.props.rev !== this.props.defaultBranch ? this.props.rev : undefined
            this.props.history.push(
                toPrettyBlobURL({
                    repoPath: this.props.repoPath,
                    rev,
                    filePath: node.id,
                })
            )
        }
    }

    private parseNodes = (paths: string[], selectedPath: string) => {
        const id = (prefix: string, restParts: string[]) => {
            if (prefix === '') {
                return restParts.join('/')
            }
            return prefix + '/' + restParts.join('/')
        }

        const parseHelper = (splits: string[][], subpath = ''): TreeNode[] => {
            const splitsByDir = groupBy(splits, split => {
                if (split.length === 1) {
                    return ''
                }
                return split[0]
            })

            const entries = flatten(
                Object.entries(splitsByDir).map(([dir, pathSplits]) => {
                    if (dir === '') {
                        return pathSplits.map(split => {
                            const filePath = id(subpath, split)
                            return {
                                text: split[0],
                                children: [],
                                id: filePath,
                                itree: {
                                    a: {
                                        attributes: {
                                            href: toPrettyBlobURL({
                                                repoPath: this.props.repoPath,
                                                rev: this.props.rev,
                                                filePath,
                                            }),
                                        },
                                    },
                                    state: {
                                        selected: filePath === selectedPath,
                                        focused: filePath === selectedPath,
                                    },
                                },
                            }
                        })
                    }

                    const dirPath = id(subpath, [dir])
                    return [
                        {
                            text: dir,
                            children: parseHelper(
                                pathSplits.map(split => split.slice(1)),
                                subpath ? subpath + '/' + dir : dir
                            ),
                            id: dirPath,
                            itree: {
                                a: {
                                    attributes: {
                                        href: toTreeURL({
                                            repoPath: this.props.repoPath,
                                            rev: this.props.rev,
                                            filePath: dirPath,
                                        }),
                                    },
                                },
                                state: {
                                    selected: dirPath === selectedPath,
                                    focused: dirPath === selectedPath,
                                },
                            },
                        },
                    ]
                })
            )

            // filter entries to those that are siblings to the selectedPath
            let filter: TreeNode[] = entries
            const selectedPathParts = selectedPath.split('/')
            let part = 0
            while (true) {
                if (part >= selectedPathParts.length) {
                    break
                }
                let matchedDir: TreeNode | undefined
                // let matchedFile: TreeNode | undefined
                for (const entry of filter) {
                    if (entry.text === selectedPathParts[part] && entry.children.length > 0) {
                        matchedDir = entry
                        break
                    }
                }
                if (matchedDir) {
                    filter = matchedDir.children
                }
                if (part === selectedPathParts.length - 1) {
                    // on the last part, filter either contains the matched file + siblings, or the matched directories children
                    break
                }
                ++part
            }

            // directories first (nodes w/ children), then sort lexicographically
            return sortBy(filter, [(e: TreeNode) => (e.children.length > 0 ? 0 : 1), 'text'])
        }

        return parseHelper(paths.map(path => path.split('/')))
    }
}
