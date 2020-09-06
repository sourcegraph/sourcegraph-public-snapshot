import React from 'react'
import H from 'history'
import { Resizable } from '../../../../shared/src/components/Resizable'
import { Search2Sidebar } from './sidebar/Search2Sidebar'
import { Search2Results } from './results/Search2Results'
import { Search2Banner } from './banner/Search2Banner'

interface Props {
    location: H.Location
}

export const Search2Area: React.FunctionComponent<Props> = ({}) => (
    <div className="d-flex w-100">
        <Resizable
            handlePosition="right"
            className="Search2Area__resizable"
            storageKey="search2-resizable"
            defaultSize={200}
            element={<Search2Sidebar className="w-100">asdf</Search2Sidebar>}
        />
        <div className="Search2Area__main flex-1">
            <Search2Banner />
            <Search2Results />
        </div>
    </div>
)
