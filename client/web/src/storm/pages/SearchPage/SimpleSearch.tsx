import { FC, useState } from 'react'

import { mdiArrowLeft, mdiHelpCircleOutline } from '@mdi/js'

import { Icon, Tooltip, Button, Text, H3 } from '@sourcegraph/wildcard'

import { CodeSearchSimpleSearch, SimpleSearchProps } from './CodeSearchSimpleSearch'
import { FindChangesSimpleSearch } from './FindChangesSimpleSearch'
import { RepoSearchSimpleSearch } from './RepoSearchSimpleSearch'

import styles from './SimpleSearch.module.scss'

export const SimpleSearch: FC<SimpleSearchProps> = props => {
    const [showState, setShowState] = useState<string>('default')

    function pickRender(): JSX.Element {
        switch (showState) {
            case 'default':
                return <SearchPicker setShowState={setShowState} />
            case 'code':
                return <CodeSearchSimpleSearch {...props} />
            case 'repo':
                return <RepoSearchSimpleSearch {...props} />
            case 'changes':
                return <FindChangesSimpleSearch {...props} />
            default:
                return <SearchPicker setShowState={setShowState} />
        }
    }

    return (
        <div>
            {showState !== 'default' && (
                <div>
                    <Button className="mb-2" onClick={() => setShowState('default')}>
                        <Icon aria-label="hover icon for help tooltip" svgPath={mdiArrowLeft} />
                        Back
                    </Button>
                    <Text>
                        Fill out the fields below to generate a search. Sourcegraph will generate the appropriate search
                        query as you fill out form fields.
                    </Text>
                </div>
            )}
            {pickRender()}
        </div>
    )
}

interface SearchPickerProps {
    setShowState: (state: string) => void
}

const SearchPicker: FC<SearchPickerProps> = ({ setShowState }) => (
    <div className="offset-1">
        <Tooltip content="This is useful if you are looking for something specific, or examples of code. Error messages, class names, variable names, etc.">
            <Button
                onClick={() => setShowState('code')}
                className={styles.searchButton}
                variant="secondary"
                outline={true}
            >
                <div>
                    <H3>Find code</H3>
                    <Text className="mt-2">Look for examples of code, specifically or with a pattern</Text>
                    <Icon aria-label="hover icon for help tooltip" className="ml-2" svgPath={mdiHelpCircleOutline} />
                </div>
            </Button>
        </Tooltip>

        <Tooltip content="This is useful if you are looking for repositories. For example, you are looking for a library you think might exist and search using repository description.">
            <Button
                onClick={() => setShowState('repo')}
                className={styles.searchButton}
                variant="secondary"
                outline={true}
            >
                <H3>Find repositories</H3>
                <Text className="mt-2">Look for repositories by name, file contents, metadata, or owners</Text>
                <Icon aria-label="hover icon for help tooltip" className="ml-2" svgPath={mdiHelpCircleOutline} />
            </Button>
        </Tooltip>

        <Tooltip content="This is useful if you are looking for changes over time, either in commit messages, by author, or code that has changed.">
            <Button
                onClick={() => setShowState('changes')}
                className={styles.searchButton}
                variant="secondary"
                outline={true}
            >
                <H3>Look for changes</H3>
                <Text className="mt-2">Look for changes in commit messages or search over diffs in the code</Text>
                <Icon aria-label="hover icon for help tooltip" className="ml-2" svgPath={mdiHelpCircleOutline} />
            </Button>
        </Tooltip>
    </div>
)
