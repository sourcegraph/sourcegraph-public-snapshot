// This file contains modifier functions to customize JSON serialization of snapshot tests using enzyme-to-json.
// These are used in share/dev/enzymeSerializer.js

import { Json } from 'enzyme-to-json'
import * as H from 'history'
import { isPlainObject } from 'lodash'
import { Observable, Subject } from 'rxjs'

function maskProps(props: Json['props'], key = '', levelsDeep = 5): unknown {
    if (!props || levelsDeep <= 0) {
        return props
    }
    if (key === 'history') {
        return '[History]'
    }
    if (key === 'location') {
        return `[Location path=${H.createPath(props as H.Location)}]`
    }
    if (props instanceof Subject) {
        return '[Subject]'
    } else if (props instanceof Observable) {
        return '[Observable]'
    }

    if (Array.isArray(props)) {
        return props.map(item => maskProps(item, '', levelsDeep - 1))
    }
    if (isPlainObject(props)) {
        return Object.fromEntries(
            Object.entries(props).map(([key, value]) => [key, maskProps(value, key, levelsDeep - 1)])
        )
    }
    return props
}

export function replaceVerboseObjects(json: Json): Json {
    if (json.props) {
        return {
            ...json,
            props: maskProps(json.props) as Json['props'],
        }
    }

    return json
}
