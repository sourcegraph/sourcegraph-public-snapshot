import { mdiFileDocumentOutline, mdiFolderOpenOutline, mdiFolderOutline } from '@mdi/js'
import type { Decorator, Meta, StoryFn } from '@storybook/react'
import classNames from 'classnames'

import { BrandedStory } from '../../stories/BrandedStory'
import { Icon } from '../Icon'
import { Link } from '../Link'

import { Tree, type TreeNode } from '.'

import styles from './Tree.story.module.scss'

const decorator: Decorator = story => <BrandedStory>{() => <div className="p-5">{story()}</div>}</BrandedStory>

const config: Meta = {
    title: 'wildcard/Tree',

    decorators: [decorator],

    parameters: {
        component: Tree,
        chromatic: {
            enableDarkMode: false,
            disableSnapshot: false,
        },
    },
}

export default config

const folder = [
    { id: 0, name: '', children: [1, 4, 9, 10, 11], parent: null },
    { id: 1, name: 'src', children: [2, 3], parent: 0 },
    { id: 2, name: 'index.js', children: [], parent: 1 },
    { id: 3, name: 'styles.css', children: [], parent: 1 },
    { id: 4, name: 'node_modules', children: [5, 7], parent: 0 },
    { id: 5, name: 'react-accessible-treeview', children: [6], parent: 4 },
    { id: 6, name: 'bundle.js', children: [], parent: 5 },
    { id: 7, name: 'react', children: [8], parent: 4 },
    { id: 8, name: 'bundle.js', children: [], parent: 7 },
    { id: 9, name: '.npmignore', children: [], parent: 0 },
    { id: 10, name: 'package.json', children: [], parent: 0 },
    { id: 11, name: 'foo.config.js', children: [], parent: 0 },
] satisfies TreeNode[]

export const Basic: StoryFn = () => (
    <Tree
        data={folder}
        defaultExpandedIds={[0, 1, 4, 5, 7]}
        renderNode={({ element, isBranch, isExpanded, handleSelect, props }): React.ReactNode => (
            <Link
                {...props}
                to="#"
                onClick={event => {
                    event.preventDefault()
                    handleSelect(event)
                }}
            >
                <Icon
                    svgPath={isBranch ? (isExpanded ? mdiFolderOpenOutline : mdiFolderOutline) : mdiFileDocumentOutline}
                    className={classNames('mr-1', styles.icon)}
                    aria-hidden={true}
                />
                {element.name}
            </Link>
        )}
    />
)
