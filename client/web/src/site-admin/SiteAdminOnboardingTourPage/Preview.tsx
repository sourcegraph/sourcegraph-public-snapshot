import { type FC, useMemo, useState } from 'react'

import { noop } from 'lodash'

import { Alert, Checkbox, Text } from '@sourcegraph/wildcard'

import { ErrorBoundary } from '../../components/ErrorBoundary'
import { TourContext } from '../../tour/components/Tour/context'
import { TourContent } from '../../tour/components/Tour/TourContent'
import type { TourConfig } from '../../tour/data'

// We use the variable names themselves as values to make them
// more prominent in the generated query links (which are not
// supposed to be clicked)
const userrepo = '$$userrepo'
const userlang = '$$userlang'
const useremail = '$$useremail'

// To avoid running actual search queries in the preview, every
// query successful
function isQuerySuccessful(): Promise<boolean> {
    return Promise.resolve(true)
}

export const TourPreview: FC<{ config: TourConfig }> = ({ config }) => {
    const [isHorizontal, setIsHorizontal] = useState(true)

    const tasks = useMemo(
        () =>
            config.tasks.map(task => ({
                ...task,
                steps: task.steps.map(step => {
                    if (step.action.type === 'search-query') {
                        // Hardcode $$snippet as possible snippet value to ensure that the
                        // placholder itself is shown in the URL
                        step = { ...step, action: { ...step.action, snippets: ['$$snippet'] } }
                    }
                    return step
                }),
                // Adding this property is necessary for the tour to display correctly
                completed: 0,
            })),
        [config]
    )

    return (
        <>
            <Text className="d-flex">
                View:&nbsp;
                <Checkbox
                    id="TourPreviewVariant"
                    checked={isHorizontal}
                    onChange={event => setIsHorizontal(event.target.checked)}
                    label="Horizontal"
                />
            </Text>
            <ErrorBoundary
                location={null}
                render={() => (
                    <Alert variant="danger">
                        An error occured while rendering the tour. Make sure the config is valid.
                    </Alert>
                )}
            >
                <TourContext.Provider
                    value={{
                        onStepClick: noop,
                        onRestart: noop,
                        userInfo: { language: userlang, repo: userrepo, email: useremail },
                        isQuerySuccessful,
                    }}
                >
                    <TourContent variant={isHorizontal ? 'horizontal' : undefined} tasks={tasks} />
                </TourContext.Provider>
            </ErrorBoundary>
        </>
    )
}
