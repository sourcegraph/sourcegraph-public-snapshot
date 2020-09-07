import MenuDownIcon from 'mdi-react/MenuDownIcon'
import React from 'react'
import { Form } from '../../../components/Form'
import { Search2FormFacet } from './facets/Search2FormFacet'

interface Props {
    className?: string
}

export const Search2Form: React.FunctionComponent<Props> = ({ className = '' }) => (
    <Form className={`Search2Form ${className} container`}>
        <div className="form-row">
            <span className="mr-3 btn btn-secondary btn-sm bg-transparent border-0 p-0">
                Calls to
                <MenuDownIcon />
            </span>
            <span className="mr-3 btn btn-secondary btn-sm bg-transparent border-0 p-0">
                Top 100k public repositories
                <MenuDownIcon />
            </span>
        </div>
        <div className="form-row">
            <input type="text" className="form-control text-monospace w-100" value="errors.WithMessage" />
        </div>
        <div className="form-row mt-2 mb-2">
            <button type="button" className="btn btn-secondary btn-sm bg-transparent p-0 border-0">
                Rules <MenuDownIcon />
            </button>
        </div>
        <div className="form-row mt-2 mb-2">
            <Search2FormFacet title="In repository" value={null} className="Search2Form__facet" />
            <Search2FormFacet title="File path" value={null} className="Search2Form__facet" />
            <Search2FormFacet title="Language" value={null} className="Search2Form__facet" />
            <Search2FormFacet title="Author" value={null} className="Search2Form__facet" />
            <Search2FormFacet title="Date" value={null} className="Search2Form__facet" />
            <Search2FormFacet title="Nearby text" value="yaml" className="Search2Form__facet" />
        </div>
    </Form>
)
