import React, { useState, useEffect, useRef, useCallback } from 'react'

import { reject } from 'lodash'
import { useLocation, useNavigate } from 'react-router-dom'

import { Input, Form } from '@sourcegraph/wildcard'
import { mdiDownload } from '@mdi/js'
import { Button, Icon } from '@sourcegraph/wildcard'
import { ChangesetReviewState, ChangesetCheckState, ChangesetState, RepoBatchChange } from '../../../../graphql-operations'
import { ChangesetFilter } from '../../ChangesetFilter'
import { isValidChangesetState, isValidChangesetReviewState, isValidChangesetCheckState } from '../../utils'

export interface ChangesetFilters {
    state: ChangesetState | null
    reviewState: ChangesetReviewState | null
    checkState: ChangesetCheckState | null
    search: string | null
}

export interface ChangesetFilterRowProps {
    onFiltersChange: (newFilters: ChangesetFilters) => void
    batchChange: RepoBatchChange;
    connectionNode: any[]
}

export const ChangesetFilterRow: React.FunctionComponent<React.PropsWithChildren<ChangesetFilterRowProps>> = ({
    onFiltersChange,
    batchChange,
    connectionNode,
}) => {
    
    const location = useLocation()
    const navigate = useNavigate()
    const searchElement = useRef<HTMLInputElement | null>(null)
    const searchParameters = new URLSearchParams(location.search)
    const [state, setState] = useState<ChangesetState | undefined>(() => {
        const value = searchParameters.get('status')
        return value && isValidChangesetState(value) ? value : undefined
    })
    const [reviewState, setReviewState] = useState<ChangesetReviewState | undefined>(() => {
        const value = searchParameters.get('review_state')
        return value && isValidChangesetReviewState(value) ? value : undefined
    })
    const [checkState, setCheckState] = useState<ChangesetCheckState | undefined>(() => {
        const value = searchParameters.get('check_state')
        return value && isValidChangesetCheckState(value) ? value : undefined
    })
    const [search, setSearch] = useState<string | undefined>(() => searchParameters.get('search') ?? undefined)
    useEffect(() => {
        const searchParameters = new URLSearchParams(location.search)
        if (state) {
            searchParameters.set('status', state)
        } else {
            searchParameters.delete('status')
        }
        if (reviewState) {
            searchParameters.set('review_state', reviewState)
        } else {
            searchParameters.delete('review_state')
        }
        if (checkState) {
            searchParameters.set('check_state', checkState)
        } else {
            searchParameters.delete('check_state')
        }
        if (search) {
            searchParameters.set('search', search)
        } else {
            searchParameters.delete('search')
        }
        if (location.search !== searchParameters.toString()) {
            navigate({ ...location, search: searchParameters.toString() }, { replace: true })
        }
        // Update the filters in the parent component.
        onFiltersChange({
            state: state || null,
            reviewState: reviewState || null,
            checkState: checkState || null,
            search: search || null,
        })
        // We cannot depend on the history, since it's modified by this hook and that would cause an infinite render loop.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [state, reviewState, checkState, search, onFiltersChange])

    const onSubmit = useCallback(
        (event?: React.FormEvent<HTMLFormElement>): void => {
            event?.preventDefault()
            setSearch(searchElement.current?.value)
        },
        [setSearch, searchElement]
    )

    const onExportClick = useCallback(() => {
        if (connectionNode) {
          const csvData = [
            ['Title', 'State', 'CheckState', 'ReviewState', 'External URL', 'Repo Name'],
          ];
          
          connectionNode.forEach(node => {
            csvData.push([
              node.title,
              node.state,
              node.checkState,
              node.reviewState,
              node.externalURL?.url || '',
              node.repository.name,
            ]);
          })
      
          const csvString = csvData.map((row) => row.join(',')).join('\n');
          const blob = new Blob([csvString], { type: 'text/csv' });
          const url = URL.createObjectURL(blob);
  
          const body = connectionNode[0].body; // Assuming connectionNode exists
          const startIndex = body.indexOf("batch-changes/") + 14; // Add the length of "batch-changes/"
          const endIndex = body.indexOf(")", startIndex);
          const batchChangeName = body.substring(startIndex, endIndex);
      
          console.log(batchChangeName);
  
          const a = document.createElement('a');
          a.href = url;
          a.download = `${batchChangeName}-changesets.csv`; 
          a.click();
    
          URL.revokeObjectURL(url);
          a.remove();
        }
 
    },[connectionNode, batchChange]);

    return (
        <>
            <div className="row no-gutters">
                <div className="m-0 col">
                    <Form className="form-inline d-flex mb-2" onSubmit={onSubmit}>
                        <Input
                            className="flex-grow-1"
                            inputClassName="flex-grow-1"
                            type="search"
                            ref={searchElement}
                            defaultValue={search}
                            placeholder="Search title and repository name"
                            aria-label="Search title and repository name"
                        />
                    </Form>
                </div>
                <div className="w-100 d-block d-md-none" />
                <div className="m-0 col col-md-auto">
                    <div className="row no-gutters">
                        <div className="col mb-2 ml-0 ml-md-2">
                            <ChangesetFilter<ChangesetState>
                                values={Object.values(ChangesetState)}
                                label="Status"
                                selected={state}
                                onChange={setState}
                                className="w-100"
                            />
                        </div>
                        <div className="col mb-2 ml-2">
                            <ChangesetFilter<ChangesetCheckState>
                                values={Object.values(ChangesetCheckState)}
                                label="Check state"
                                selected={checkState}
                                onChange={setCheckState}
                                className="w-100"
                            />
                        </div>

                        <div className="w-100 d-block d-sm-none" />
                        <div className="col mb-2 ml-0 ml-sm-2">
                            <ChangesetFilter<ChangesetReviewState>
                                values={reject(
                                    Object.values(ChangesetReviewState),
                                    state =>
                                        state === ChangesetReviewState.COMMENTED ||
                                        state === ChangesetReviewState.DISMISSED
                                )}
                                label="Review state"
                                selected={reviewState}
                                onChange={setReviewState}
                                className="w-100"
                            />
                        </div>
                        <div className="col mb-2 ml-2">
                <Button variant="primary" onClick={onExportClick} >
                    <Icon aria-hidden={true} svgPath={mdiDownload} /> Export 
                </Button> 
            </div>
                        <div className="col d-block d-sm-none ml-2" />
                    </div>
                </div>
            </div>
        </>
    )
}
