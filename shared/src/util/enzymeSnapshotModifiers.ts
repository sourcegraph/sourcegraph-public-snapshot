// This file contains modifier functions to customize JSON serialization of snapshot tests using enzyme-to-json.
// These are used in share/dev/enzymeSerializer.js

import { Json } from 'enzyme-to-json'
import { Subject, Observable } from 'rxjs'
import * as H from 'history'
import { isPlainObject } from 'lodash'

function maskProps(props: Record<string, any>, levelsDeep = 3): void {
    if (props.history) {
        props.history = '[History]'
    }
    if (props.location) {
        props.location = `[Location path=${H.createPath(props.location as H.Location)}]`
    }

    for (const property of Object.keys(props)) {
        if (props[property] instanceof Subject) {
            props[property] = '[Subject]'
        } else if (props[property] instanceof Observable) {
            props[property] = '[Observable]'
        } else if (levelsDeep > 0 && isPlainObject(props[property])) {
            maskProps(props[property], levelsDeep - 1)
        }
    }
}

export function replaceVerboseObjects(json: Json): Json {
    if (json.props) {
        maskProps(json.props)
    }

    return json
}
