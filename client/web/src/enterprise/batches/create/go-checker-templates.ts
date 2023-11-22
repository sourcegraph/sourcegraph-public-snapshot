import { quoteYAMLString } from '../batch-spec/yaml-util'

export function goCheckerSA6005Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to improve inefficient string comparison with strings.ToLower or strings.ToUpper
on:
    - repositoriesMatchingQuery: |
        if strings.ToLower(:[1]) == strings.ToLower(:[2]) or if strings.ToUpper(:[1]) == strings.ToUpper(:[2]) or if strings.ToLower(:[1]) != strings.ToLower(:[2]) or if strings.ToUpper(:[1]) != strings.ToUpper(:[2]) patternType:structural archived:no

steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [SA6005_01]
            match='if strings.ToLower(:[1]) == strings.ToLower(:[2])'
            rewrite='if strings.EqualFold(:[1], :[2])'

            [SA6005_02]
            match='if strings.ToLower(:[1]) != strings.ToLower(:[2])'
            rewrite='if !strings.EqualFold(:[1], :[2])'

            [SA6005_03]
            match='if strings.ToUpper(:[1]) == strings.ToUpper(:[2])'
            rewrite='if strings.EqualFold(:[1], :[2])'

            [SA6005_04]
            match='if strings.ToUpper(:[1]) != strings.ToUpper(:[2])'
            rewrite='if !strings.EqualFold(:[1], :[2])'

changesetTemplate:
    title: Improve inefficient string comparison with strings.ToLower or strings.ToUpper
    body: This batch change uses [Comby](https://comby.dev) to improve inefficient string comparison with strings.ToLower or strings.ToUpper
    branch: batches/\${{batch_change.name}}
    commit:
        message: Improve inefficient string comparison with strings.ToLower or strings.ToUpper
`
}

export function goCheckerS1002Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to omit comparison with boolean constant
on:
    - repositoriesMatchingQuery: |
        if :[1:e] == false or if :[1:e] == true patternType:structural archived:no

steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [S1002_01]
            match='if :[1:e] == true '
            rewrite='if :[1]'

            [S1002_02]
            match='if :[1:e] == false '
            rewrite='if !:[1]'

changesetTemplate:
    title: Omit comparison with boolean constant
    body: This batch change uses [Comby](https://comby.dev) to omit comparison with boolean constant
    branch: batches/\${{batch_change.name}}
    commit:
        message: Omit comparison with boolean constant
`
}

export function goCheckerS1003Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to replace calls to strings.Index with strings.Contains.
on:
    - repositoriesMatchingQuery: |
        strings.Index(:[1], :[2]) < 0  or strings.Index(:[1], :[2]) == -1 or strings.Index(:[1], :[2]) != -1 or strings.Index(:[1], :[2]) >= 0 or strings.Index(:[1], :[2]) > -1 or strings.IndexAny(:[1], :[2]) < 0 or strings.IndexAny(:[1], :[2]) == -1 or strings.IndexAny(:[1], :[2]) != -1 or strings.IndexAny(:[1], :[2]) >= 0 or strings.IndexAny(:[1], :[2]) > -1 or strings.IndexRune(:[1], :[2]) < 0 or strings.IndexRune(:[1], :[2]) == -1 or strings.IndexRune(:[1], :[2]) != -1 or strings.IndexRune(:[1], :[2]) >= 0 or strings.IndexRune(:[1], :[2]) > -1 patternType:structural archived:no

steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [S1003_01]
            match='strings.IndexRune(:[1], :[2]) > -1'
            rewrite='strings.ContainsRune(:[1], :[2])'

            [S1003_02]
            match='strings.IndexRune(:[1], :[2]) >= 0'
            rewrite='strings.ContainsRune(:[1], :[2])'

            [S1003_03]
            match='strings.IndexRune(:[1], :[2]) != -1'
            rewrite='strings.ContainsRune(:[1], :[2])'

            [S1003_04]
            match='strings.IndexRune(:[1], :[2]) == -1'
            rewrite='!strings.ContainsRune(:[1], :[2])'

            [S1003_05]
            match='strings.IndexRune(:[1], :[2]) < 0'
            rewrite='!strings.ContainsRune(:[1], :[2])'

            [S1003_06]
            match='strings.IndexAny(:[1], :[2]) > -1'
            rewrite='strings.ContainsAny(:[1], :[2])'

            [S1003_07]
            match='strings.IndexAny(:[1], :[2]) >= 0'
            rewrite='strings.ContainsAny(:[1], :[2])'

            [S1003_08]
            match='strings.IndexAny(:[1], :[2]) != -1'
            rewrite='strings.ContainsAny(:[1], :[2])'

            [S1003_09]
            match='strings.IndexAny(:[1], :[2]) == -1'
            rewrite='!strings.ContainsAny(:[1], :[2])'

            [S1003_10]
            match='strings.IndexAny(:[1], :[2]) < 0'
            rewrite='!strings.ContainsAny(:[1], :[2])'

            [S1003_11]
            match='strings.Index(:[1], :[2]) > -1'
            rewrite='strings.Contains(:[1], :[2])'

            [S1003_12]
            match='strings.Index(:[1], :[2]) >= 0'
            rewrite='strings.Contains(:[1], :[2])'

            [S1003_13]
            match='strings.Index(:[1], :[2]) != -1'
            rewrite='strings.Contains(:[1], :[2])'

            [S1003_14]
            match='strings.Index(:[1], :[2]) == -1'
            rewrite='!strings.Contains(:[1], :[2])'

            [S1003_15]
            match='strings.Index(:[1], :[2]) < 0'
            rewrite='!strings.Contains(:[1], :[2])'
changesetTemplate:
    title: Replace calls to strings.Index with strings.Contains.
    body: This batch change uses [Comby](https://comby.dev) to replace calls to strings.Index with strings.Contains.
    branch: batches/\${{batch_change.name}}
    commit:
        message: Replace calls to strings.Index with strings.Contains.
`
}

export function goCheckerS1004Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to replace call to bytes.Compare with bytes.Equal
on:
    - repositoriesMatchingQuery: |
        bytes.Compare(:[1], :[2]) != 0 or bytes.Compare(:[1], :[2]) == 0 patternType:structural archived:no

steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [S1004_01]
            match='bytes.Compare(:[1], :[2]) == 0'
            rewrite='bytes.Equal(:[1], :[2])'

            [S1004_02]
            match='bytes.Compare(:[1], :[2]) != 0'
            rewrite='!bytes.Equal(:[1], :[2])'

changesetTemplate:
    title: Replace call to bytes.Compare with bytes.Equal
    body: This batch change uses [Comby](https://comby.dev) to replace call to bytes.Compare with bytes.Equal
    branch: batches/\${{batch_change.name}}
    commit:
        message: Replace call to bytes.Compare with bytes.Equal
`
}

export function goCheckerS1005Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to drop unnecessary use of the blank identifier
on:
    - repositoriesMatchingQuery: |
        for :[1:e], :[~_] := range or for :[1:e], :[~_] = range or for :[~_] = range or for :[~_], :[~_] = range patternType:structural archived:no

steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [S1005_01]
            match='for :[~_], :[~_] = range'
            rewrite='for range'

            [S1005_02]
            match='for :[~_] = range'
            rewrite='for range'

            [S1005_03]
            match='for :[1:e], :[~_] = range'
            rewrite='for :[1] = range'

            [S1005_04]
            match='for :[1:e], :[~_] := range'
            rewrite='for :[1] := range'
changesetTemplate:
    title: Drop unnecessary use of the blank identifier
    body: This batch change uses [Comby](https://comby.dev) to drop unnecessary use of the blank identifier
    branch: batches/\${{batch_change.name}}
    commit:
        message: Drop unnecessary use of the blank identifier
`
}

export function goCheckerS1006Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to use for { ... } for infinite loops
on:
    - repositoriesMatchingQuery: |
        for true {:[x]} patternType:structural archived:no

steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            match='for true {:[x]}'
            rewrite='for {:[x]}'
changesetTemplate:
    title: Use for { ... } for infinite loops
    body: This batch change uses [Comby](https://comby.dev) to use for { ... } for infinite loops
    branch: batches/\${{batch_change.name}}
    commit:
        message: Use for { ... } for infinite loops
`
}

export function goCheckerS1010Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to omit default slice index
on:
    - repositoriesMatchingQuery: |
        :[s.][:len(:[s])] patternType:structural archived:no

steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            match=':[s.][:len(:[s])]'
            rewrite=':[s.][:]'
changesetTemplate:
    title: Omit default slice index
    body: This batch change uses [Comby](https://comby.dev) to omit default slice index
    branch: batches/\${{batch_change.name}}
    commit:
        message: omit default slice index
`
}

export function goCheckerS1012Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to replace time.Now().Sub(x) with time.Since(x)
on:
    - repositoriesMatchingQuery: |
        time.Now().Sub(:[x]) patternType:structural archived:no

steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            match='time.Now().Sub(:[x])'
            rewrite='time.Since(:[x])'
changesetTemplate:
    title: Replace time.Now().Sub(x) with time.Since(x)
    body: This batch change uses [Comby](https://comby.dev) to replace time.Now().Sub(x) with time.Since(x)
    branch: batches/\${{batch_change.name}}
    commit:
        message: replace time.Now().Sub(x) with time.Since(x)
`
}

export function goCheckerS1019Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to simplify make call by omitting redundant arguments
on:
    - repositoriesMatchingQuery: |
        make(chan int, 0) or make(map[:[[1]]]:[[1]], 0) or make(:[1], :[2], :[2]) patternType:structural archived:no

steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [S1019_01]
            match='make(:[1], :[2], :[2])'
            rewrite='make(:[1], :[2])'

            [S1019_02]
            match='make(map[:[[1]]]:[[1]], 0)'
            rewrite='make(map[:[[1]]]:[[1]])'

            [S1019_03]
            match='make(chan int, 0)'
            rewrite='make(chan int)'
changesetTemplate:
    title: Simplify make call by omitting redundant arguments
    body: This batch change uses [Comby](https://comby.dev) to simplify make call by omitting redundant arguments
    branch: batches/\${{batch_change.name}}
    commit:
        message: simplify make call by omitting redundant arguments
`
}

export function goCheckerS1020Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to omit redundant nil check in type assertion
on:
    - repositoriesMatchingQuery: |
        if :[_.], ok := :[i.].(:[T]); :[i.] != nil && ok {:[body]} or if :[_.], ok := :[i.].(:[T]); ok && :[i.] != nil {:[body]} or if :[i.] != nil { if :[_.], ok := :[i.].(:[T]); ok {:[body]}} patternType:structural archived:no

steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [S1020_01]
            match='if :[_.], ok := :[i.].(:[T]); ok && :[i.] != nil {:[body]}'
            rewrite='if :[_.], ok := :[i.].(:[T]); ok {:[body]}'

            [S1020_02]
            match='if :[_.], ok := :[i.].(:[T]); :[i.] != nil && ok {:[body]}'
            rewrite='if :[_.], ok := :[i.].(:[T]); ok {:[body]}'

            [S1020_03]
            match='''
            if :[i.] != nil {
            if :[_.], ok := :[i.].(:[T]); ok {:[body]}
            }'''
            rewrite='if :[_.], ok := :[i.].(:[T]); ok {:[body]}'
changesetTemplate:
    title: Omit redundant nil check in type assertion
    body: This batch change uses [Comby](https://comby.dev) to omit redundant nil check in type assertion
    branch: batches/\${{batch_change.name}}
    commit:
        message: omit redundant nil check in type assertion
`
}

export function goCheckerS1023Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to omit redundant control flow
on:
    - repositoriesMatchingQuery: |
        func() {:[body] return } or func :[fn.](:[args]) {:[body] return } patternType:structural archived:no

steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [S1023_01]
            match='func :[fn.](:[args]) {:[body] return }'
            rewrite='func :[fn.](:[args]) {:[body]}'

            [S1023_02]
            match='func() {:[body] return }'
            rewrite='func() {:[body]}'

changesetTemplate:
    title: Omit redundant control flow
    body: This batch change uses [Comby](https://comby.dev) to omit redundant control flow
    branch: batches/\${{batch_change.name}}
    commit:
        message: omit redundant control flow
`
}

export function goCheckerS1024Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to replace x.Sub(time.Now()) with time.Until(x)
on:
    - repositoriesMatchingQuery: |
        :[x.].Sub(time.Now()) patternType:structural archived:no
steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [S1023_01]
            match='func :[fn.](:[args]) {:[body] return }'
            rewrite='func :[fn.](:[args]) {:[body]}'

            [S1023_02]
            match='func() {:[body] return }'
            rewrite='func() {:[body]}'

changesetTemplate:
    title: Replace x.Sub(time.Now()) with time.Until(x)
    body: This batch change uses [Comby](https://comby.dev) to replace x.Sub(time.Now()) with time.Until(x)
    branch: batches/\${{batch_change.name}}
    commit:
        message: replace x.Sub(time.Now()) with time.Until(x)
`
}

export function goCheckerS1025Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to enforce don’t use fmt.Sprintf("%s", x) unnecessarily
on:
    - repositoriesMatchingQuery: |
        fmt.Println("%s", ":[s]") patternType:structural archived:no
steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            match='fmt.Println("%s", ":[s]")'
            rewrite='fmt.Println(":[s]")'

changesetTemplate:
    title: Don’t use fmt.Sprintf("%s", x) unnecessarily
    body: This batch change uses [Comby](https://comby.dev) to enforce don’t use fmt.Sprintf("%s", x) unnecessarily
    branch: batches/\${{batch_change.name}}
    commit:
        message: don’t use fmt.Sprintf("%s", x) unnecessarily
`
}

export function goCheckerS1028Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to simplify error construction with fmt.Errorf
on:
    - repositoriesMatchingQuery: |
        errors.New(fmt.Sprintf(:[1])) patternType:structural archived:no
steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [S1028_01]
            match='errors.New(fmt.Sprintf(:[1]))'
            rewrite='fmt.Errorf(:[1])'

changesetTemplate:
    title: Simplify error construction with fmt.Errorf
    body: This batch change uses [Comby](https://comby.dev) to simplify error construction with fmt.Errorf
    branch: batches/\${{batch_change.name}}
    commit:
        message: simplify error construction with fmt.Errorf
`
}

export function goCheckerS1029Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to range over the string directly
on:
    - repositoriesMatchingQuery: |
        for :[~_], :[r.] := range []rune(:[s.]) patternType:structural archived:no
steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [S1029_01]
            match='for :[~_], :[r.] := range []rune(:[s.])'
            rewrite='for _, :[r] := range :[s]'

changesetTemplate:
    title: Range over the string directly
    body: This batch change uses [Comby](https://comby.dev) to range over the string directly
    branch: batches/\${{batch_change.name}}
    commit:
        message: range over the string directly
`
}

export function goCheckerS1032Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to use sort.Ints(x), sort.Float64s(x), and sort.Strings(x)
on:
    - repositoriesMatchingQuery: |
        sort.Sort(sort.Float64Slice(:[1])) or sort.Sort(sort.StringSlice(:[1])) or sort.Sort(sort.IntSlice(:[1])) patternType:structural archived:no
steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [S1032_01]
            match='sort.Sort(sort.IntSlice(:[1]))'
            rewrite='sort.Ints(:[1])'

            [S1032_02]
            match='sort.Sort(sort.StringSlice(:[1]))'
            rewrite='sort.Strings(:[1])'

            [S1032_03]
            match='sort.Sort(sort.Float64Slice(:[1]))'
            rewrite='sort.Float64s(:[1])'

changesetTemplate:
    title: Use sort.Ints(x), sort.Float64s(x), and sort.Strings(x)
    body: This batch change uses [Comby](https://comby.dev) to use sort.Ints(x), sort.Float64s(x), and sort.Strings(x)
    branch: batches/\${{batch_change.name}}
    commit:
        message: use sort.Ints(x), sort.Float64s(x), and sort.Strings(x)
`
}

export function goCheckerS1035Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to remove redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header
on:
    - repositoriesMatchingQuery: |
        headers.Set(http.CanonicalHeaderKey(:[1])) or headers.Get(http.CanonicalHeaderKey(:[1])) or headers.Del(http.CanonicalHeaderKey(:[1])) or headers.Add(http.CanonicalHeaderKey(:[1]), :[1]) patternType:structural archived:no
steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [S1035_01]
            match='headers.Add(http.CanonicalHeaderKey(:[1]), :[1])'
            rewrite='headers.Add(:[1], :[1])'

            [S1035_02]
            match='headers.Del(http.CanonicalHeaderKey(:[1]))'
            rewrite='headers.Del(:[1])'

            [S1035_03]
            match='headers.Get(http.CanonicalHeaderKey(:[1]))'
            rewrite='headers.Get(:[1])'

            [S1035_04]
            match='headers.Set(http.CanonicalHeaderKey(:[1]))'
            rewrite='headers.Set(:[1])'

changesetTemplate:
    title: Remove redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header
    body: This batch change uses [Comby](https://comby.dev) to remove redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header
    branch: batches/\${{batch_change.name}}
    commit:
        message: remove redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header
`
}

export function goCheckerS1037Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to remove redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header
on:
    - repositoriesMatchingQuery: |
        headers.Set(http.CanonicalHeaderKey(:[1])) or headers.Get(http.CanonicalHeaderKey(:[1])) or headers.Del(http.CanonicalHeaderKey(:[1])) or headers.Add(http.CanonicalHeaderKey(:[1]), :[1]) patternType:structural archived:no
steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [S1037_01]
            match='''
            select {
                case <-time.After(:[t]):
            }'''
            rewrite='time.Sleep(:[t])'

changesetTemplate:
    title: Redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header
    body: This batch change uses [Comby](https://comby.dev) to remove redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header
    branch: batches/\${{batch_change.name}}
    commit:
        message: remove redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header
`
}

export function goCheckerS1038Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to remove redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header
on:
    - repositoriesMatchingQuery: |
        headers.Set(http.CanonicalHeaderKey(:[1])) or headers.Get(http.CanonicalHeaderKey(:[1])) or headers.Del(http.CanonicalHeaderKey(:[1])) or headers.Add(http.CanonicalHeaderKey(:[1]), :[1]) patternType:structural archived:no
steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [S1037_01]
            match='''
            select {
                case <-time.After(:[t]):
            }'''
            rewrite='time.Sleep(:[t])'

changesetTemplate:
    title: Redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header
    body: This batch change uses [Comby](https://comby.dev) to remove redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header
    branch: batches/\${{batch_change.name}}
    commit:
        message: remove redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header
`
}

export function goCheckerS1039Template(name: string): string {
    return `name: ${quoteYAMLString(name)}
description: |
    This batch change uses [Comby](https://comby.dev) to improve unnecessary use of fmt.Sprint
on:
    - repositoriesMatchingQuery: |
        fmt.Sprintf("%s", ":[s]") patternType:structural archived:no
steps:
    - run: comby -config /tmp/rule.toml -f .go -i -exclude-dir vendor,.
      container: comby/comby
      files:
        /tmp/rule.toml: |
            [S1039]
            match='fmt.Sprintf("%s", ":[s]")'
            rewrite='":[s]"'
changesetTemplate:
    title: Improve unnecessary use of fmt.Sprint
    body: This batch change uses [Comby](https://comby.dev) to improve unnecessary use of fmt.Sprint
    branch: batches/\${{batch_change.name}}
    commit:
        message: unnecessary use of fmt.Sprint
`
}

export function getTemplateRenderer(kind: string | null): ((name: string) => string) | undefined {
    if (!kind) {
        return undefined
    }

    switch (kind) {
        case 'goCheckerSA6005': {
            return goCheckerSA6005Template
        }
        case 'goCheckerS1002': {
            return goCheckerS1002Template
        }
        case 'goCheckerS1003': {
            return goCheckerS1003Template
        }
        case 'goCheckerS1004': {
            return goCheckerS1004Template
        }
        case 'goCheckerS1005': {
            return goCheckerS1005Template
        }
        case 'goCheckerS1006': {
            return goCheckerS1006Template
        }
        case 'goCheckerS1010': {
            return goCheckerS1010Template
        }
        case 'goCheckerS1012': {
            return goCheckerS1012Template
        }
        case 'goCheckerS1019': {
            return goCheckerS1019Template
        }
        case 'goCheckerS1020': {
            return goCheckerS1020Template
        }
        case 'goCheckerS1023': {
            return goCheckerS1023Template
        }
        case 'goCheckerS1024': {
            return goCheckerS1024Template
        }
        case 'goCheckerS1025': {
            return goCheckerS1025Template
        }
        case 'goCheckerS1028': {
            return goCheckerS1028Template
        }
        case 'goCheckerS1029': {
            return goCheckerS1029Template
        }
        case 'goCheckerS1032': {
            return goCheckerS1032Template
        }
        case 'goCheckerS1035': {
            return goCheckerS1035Template
        }
        case 'goCheckerS1037': {
            return goCheckerS1037Template
        }
        case 'goCheckerS1038': {
            return goCheckerS1038Template
        }
        case 'goCheckerS1039': {
            return goCheckerS1039Template
        }
        default: {
            return undefined
        }
    }
}
