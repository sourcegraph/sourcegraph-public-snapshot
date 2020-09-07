import React from 'react'
import H from 'history'
import { Search2Form } from './form/Search2Form'
import { Search2Results } from './results/Search2Results'
import { Search2Banner } from './banner/Search2Banner'

interface Props {
    location: H.Location
}

export const Search2Area: React.FunctionComponent<Props> = ({}) => (
    <div className="d-flex flex-column w-100 mt-1">
        <Search2Form className="Search2Area__form border-bottom">asdf</Search2Form>
        <div className="Search2Area__main">
            <Search2Banner />
            <Search2Results />
        </div>
    </div>
)
