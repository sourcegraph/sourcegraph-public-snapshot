import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'

import { type CompletionRequest, getCodyCompletionOneShot } from './api'

export async function translateToQuery(input: string, user: AuthenticatedUser | null): Promise<string | null> {
    const messages = getCompletionRequestMessages(input, user)
    const result = await getCodyCompletionOneShot(messages)
    if (!result.includes('contents>') && !result.includes('filters>')) {
        return null
    }

    const contents =
        result
            ?.match(/<contents>.*<\/contents>/)
            ?.reduce(
                (content, match) => `${content}${match.replace('<contents>', ' ').replace('</contents>', ' ')}`,
                ''
            )
            ?.trim() || '.*'
    const filters =
        result
            ?.match(/<filters>.*<\/filters>/)
            ?.reduce((content, match) => `${content}${match.replace('<filters>', ' ').replace('</filters>', ' ')}`, '')
            ?.trim() || ''
    const query = `${contents} ${filters}`
        .replaceAll('\n', ' ')
        .replaceAll(/\s{2,}/g, ' ')
        .trim()

    return query
}

const getCompletionRequestMessages = (input: string, user: AuthenticatedUser | null): CompletionRequest['messages'] => {
    const username = user?.username || 'anonymous'

    return [
        // teaching formating
        {
            speaker: 'human',
            text:
                `I am a Sourcegraph user and my SOURCEGRAPH_USERNAME is ${username}.\n` +
                'I want you to act like an expert at writing Sourcegraph search queries based on what a I request you.\n' +
                'I will request you in the following format: <request>HUMAN_REQUEST</request>.\n' +
                'I want you to convert this request into Sourcegraph search query in the following format: <contents>QUERY_CONTENT</contents><filters>QUERY_FILTER</filters>.\n' +
                'I want you to strictly follow the search query format and only respond in that format when I request you.\n' +
                'A typical query describes content or filenames to find across all repositories.\n' +
                'QUERY_CONTENT is a regular expression used to match content in the repositories. QUERY_CONTENT in put inside the <contents></contents> block. The default value of QUERY_CONTENT is .* and you can use it like <contents>.*</contents>.\n' +
                'QUERY_FILTER has the following structure FILTER_NAME:FILTER_VALUE. You can put multiple QUERY_FILTER inside the <filters> block like <filters>FILTER_NAME_1:FILTER_VALUE_1 FILTER_NAME_2:FILTER_VALUE_2 FILTER_NAME_3:FILTER_VALUE_3</fitlers>. QUERY_FILTER is used to limit search to specific repositories or files.',
        },
        {
            speaker: 'assistant',
            text: "Understood, I'm ready to help you write Sourcegraph search queries in the format you've specified. Please send me your request in the format <request>HUMAN_REQUEST</request> and I'll respond with a search query in the format <contents>QUERY_CONTENT</contents><filters>QUERY_FILTER</filters>.",
        },

        // teaching filters
        {
            speaker: 'human',
            text:
                'Following is the list of QUERY_FILTER along with their descriptions. I want you to use only those QUERY_FILTER that are mentioned in the list below:\n' +
                '1. <filters>repo:REPO_NAME_REGEXP</filters> only include results from repositories whose path matches the REPO_NAME_REGEXP regular expression.\n' +
                '2. <filters>repo:REPO_NAME_REGEXP rev:REVISION</filters> only include results from specific REVISION of repositories whose path matches the REPO_NAME_REGEXP regular express. REVISION refers to repository revisions namely branch, commit, or tags. If you want results from specific branch my_branch on a repo use <fitlers>repo:REPO_NAME_REGEXP rev:my_branch</fitlers>. If you want results upto specific commit 1735d48 in a repo use <filters>repo:REPO_NAME_REGEXP rev:1735d48</fitlers>.\n' +
                "3. <filters>-repo:REPO_NAME_REGEXP</filters>. The '-' sign signifies that this QUERY_FILTER excludes results from repositories whose path matches the REPO_NAME_REGEXP regular expression. It can be used in combination of repo:REPO_NAME_REGEXP filter as well. If you want to include frontend repo and exclude backend repo use <filters>repo:frontend -repo:backend</filters>.\n" +
                '4. <filters>file:PATH-REGEXP</filters> only include results in files whose full path matches the PATH_REGEXP regular expression. \n' +
                "5. <filters>-file:PATH-REGEXP</filters>, the '-' sigh signifies that this QUERY_FILTER excludes results from files whose path matches the PATH_REGEXP regular expression. It can be used in combination of file:PATH-REGEXP filter. If you want to include \\.go files and exclude _test\\.go files use <filters>file:\\.go -file:_test\\.go</filters>.\n" +
                '6. <filters>file:has.owner(OWNER_NAME)</filters> only include results from files owned by a user with name matching OWNER_NAME regular expression. To only include files owned by james-connor use <filters>file.has.owner(james-connor).\n' +
                '7. <filters>file:has.owner(OWNER_EMAIL)</filters> only include results from files owned by a user with email matching OWNER_EMAIL regular expression. To only include files owner by peter@gmail.com use <filters>file.has.owner(peter@gmail.com)</filters>.\n' +
                '8. <filters>lang:LANGUAGE</filters> only include results from files in the specified LANGUAGE. LANGUAGE can be typescript, javascript, go, css, scss, html, markdown, rust, c++, java, or other programming language names.\n' +
                '9. <filters>type:SEARCH_TYPE</filters>. SEARCH_TYPE Specifies the type of search. SEARCH_TYPE can have 3 values diff, commit or symbol. If you want to search over changes to code use type:diff. If you want to search over a commit message use type:commit. If you want to search for a symbol use type:symbol.\n' +
                '10. <filters>type:diff</filters> seaches over changes done under a commit, including adding code, deleting code, deleting files, new files created, removal of code or any other modifications made. If you want to search over any such changes use type:diff.\n' +
                '11. <filters>author:"AUTHOR_PATTERN"</filters> only include results from diffs or commits authored by git author matching AUTHOR_PATTERN regular expression. AUTHOR_PATTERN is of the form "Full Name <user@example.com>", where Full Name is name of the author and user@example.com is the email of the author. To include commits from authors from a specific domain, use <filters>type:commit author:"example.com$"</filters>. To include commits from author having USERNAME as name use <filters>type:commit author:"USERNAME"</filters>. You must use type:diff or type:commit with author QUERY_FILTER. You must not use author QUERY_FILTER with type:symbol. \n' +
                '12. <filters>author:@SOURCEGRAPH_USERNAME</filters>. Only include results from diffs or commits authored by the Sourcegraph user having either email or username matching with SOURCEGRAPH_USERNAME regular expression. SOURCEGRAPH_USERNAME is the username of user in the sourcegraph instance. To include commits from sourcegraph user having username as SG_USER use <filters>type:commit author:@SG_USER</filters>. You must use type:diff or type:commit with author QUERY_FILTER. You must not use author QUERY_FILTER with type:symbol.\n' +
                '13. <filters>-author:"AUTHOR_PATTERN"</filters> the \'-\' sign signifies that this QUERY_FILTER excludes results from diffs or commits authored by AUTHOR_PATTERN regular expression. AUTHOR_PATTERN is of the form "Full Name <user@example.com>", where Full Name is name of the author and user@example.com is the email of the author. To exclude commits from authors from a specific domain, use <filters>type:commit -author:"example.com>$"</filters>. To exclude commits from author having USERNAME as name but include commit from author having ANOTHER_USER as name use <filters>type:commit author:"ANOTHER_USER" -author:"USERNAME"</filters>. You must use type:diff or type:commit with author QUERY_FILTER. You must not use author QUERY_FILTER with type:symbol.\n' +
                '14. <filters>-author:@SOURCEGRAPH_USERNAME</filters>. Exclude results from commits authored by the Sourcegraph user having either email or username matching with SOURCEGRAPH_USERNAME. SOURCEGRAPH_USERNAME is the username of user in the sourcegraph instance. To exclude diffs or commits from sourcegraph user having username as SG_USER but include commits from sourcegraph user having username as ANOTHER_SG_USER use <filters>type:commit author:@ANOTHER_SG_USER -author:@SG_USER</filters>. You must use type:diff or type:commit with -author QUERY_FILTER. You must not use -author QUERY_FILTER with type:symbol.\n' +
                '15. <filters>before:"TIME_FRAME"</filters> only include results from diffs or commits which have a commit date before the specified TIME_FRAME. To only include changes before last thursday use <filters>type:diff before:"last thursday"</filters>. To only include diffs before 1 november 2020 use <filters>type:diff before:"1 november 2020".\n' +
                '16. <filters>after:"TIME_FRAME"</filters> only include results from diffs or commits which have a commit date after the specified TIME_FRAME. To only include diffs after last thursday use <filters>type:diff after:"last thursday"</filters>. To only include changes after 1 november 2020 use <filters>type:diff after:"1 november 2020"</filters>.\n' +
                '17. <filters>message:"COMMIT_MESSAGE"</filters> only include results from diffs or commits which have a commit message containing the string COMMIT_MESSAGE. To include diffs or commits including bugfix use <fitlers>message:"bugfix"</filters>.\n' +
                '18. <filters>select:repo</filters> lists only distinct repo paths from the search results. If you want to list repos having package.json file use <filters>select:repo file:package.json</filters>.\n' +
                '19. <filters>select:file.owners</filters> lists owners of files from the search results. If you want to find owners of ci directory use <contents>.*</contents><filters>select:file.owners file:ci/</filters>.\n' +
                '20. <filters>select:file</filters> lists file paths from the search results. If you want to list all files inside backend directory use <filters>select:file file:backend/</filters>.\n' +
                '21. <filters>select:commit.diff.added</filters> lists additions (new lines of code added) made in a commit, matching the QUERY_CONTENT regular expression. For example to search for recent commits that added TODOs in your code use <contents>TODOs</contents><filters>type:diff select:commit.diff.added</filters>. This QUERY_FILTER can be used to search for added lines of code, added references to symbols, added function calls etc. If the QUERY_CONTENT is null, always use .* as default value of QUERY_CONTENT. \n' +
                '22. <filters>select:commit.diff.removed</filters> lists removals (lines of code removed) made in a commit, matching the QUERY_CONTENT regular expression. For example to search for recent commits that removed TODOs in your code use <contents>TODOs</contents><filters>type:diff select:commit.diff.removed</filters>. This QUERY_FILTER can be used to search for removed lines of code, removed references to symbols, removed function calls etc. If the QUERY_CONTENT is null, always use .* as default value of QUERY_CONTENT.\n' +
                '23. <filters>select:symbol.SYMBOL_KIND</filters> only finds symbols of specific SYMBOL_KIND based on the QUERY_CONTENT. SYMBOL_KIND can be of following types: file, module, namespace, package, class, method, property, field, constructor, enum, interface, function, variable, constant, string, number, boolean, array, object, key, null, enum-number, struct, event, operator, type-parameter. To find functions with name like use.*, the Sourcegraph search query will be <contents>use.*</contents><filters>select.symbol.function type:symbol</filters>. If the QUERY_CONTENT is null, always use .* as default value of QUERY_CONTENT. To list all classes in typescript files use <contents>.*</contents><filters>select:symbol.class type:symbol lang:typescript</filters>. You must not use type:diff QUERY_FILTER or type:commit QUERY_FILTER with select:symbol.SYMBOL_KIND QUERY_FILTER. You must use type:symbol QUERY_FILTER when using select:symbol.SYMBOL_KIND QUERY_FILTER. <filters>select:symbol.function</fitlers> is wrong. <filters>select:symbol.function type:symbol</filters is correct.',
        },
        {
            speaker: 'assistant',
            text: "Sure, I'll only use the QUERY_FILTER mentioned in the list you provided.",
        },

        // teach operators
        {
            speaker: 'human',
            text:
                'I want you to use the following 3 operators if necessary to build Sourcegraph search query:\n' +
                '1. AND operator returns results from files containing matches on the left and right side of the AND. To search files containing both foo and bar use <contents>foo AND bar</contents>.\n' +
                '2. OR operator returns results from files containing matches either on the left or right side, or both. To search files containing either of foo or bar use <contents>foo OR bar</contents>.\n' +
                '3. NOT operator excludes results from files containing matches with the term after NOT. To search files not containing foo use <contents>NOT foo</contents>.',
        },
        {
            speaker: 'assistant',
            text: 'Sure, I can help you build a Sourcegraph search query using the operators you provided.',
        },

        // teach no results
        {
            speaker: 'human',
            text: 'If you are unable to convert human request into a valid Sourcegraph search query using the info provided above then you must return <noresults></noresults>.',
        },
        {
            speaker: 'assistant',
            text: "Yes, if I'm unable to understand or convert a human request into a valid Sourcegraph search query, I will return <noresults></noresults> to indicate that no results were found.",
        },

        // repo filter basic
        { speaker: 'human', text: 'What is the query for <request>multierror repo</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>repo:multierror</filters>' },

        // repo filter with branch name tag
        { speaker: 'human', text: 'What is the query for <request>branch user/bug-fix on react repo</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>repo:react rev:user/bug-fix</filters>' },

        // repo filter with commit tag
        { speaker: 'human', text: 'What is the query for <request>commit 1g246ih in express repo</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>repo:express rev:1g246ih</filters>' },

        // excluding repo fitler
        {
            speaker: 'human',
            text: 'What is the query for <request>featureFlags and flagSets excluding dotenv repo</request>?',
        },
        { speaker: 'assistant', text: '<contents>featureFlags AND flagSets</contents><filters>-repo:dotenv</filters>' },

        // both excluding & including repo fitler
        { speaker: 'human', text: 'What is the query for <request>all microsoft repo excluding windows98</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>repo:microsoft -repo:windows98</filters>' },

        // file filter with npm package file name
        { speaker: 'human', text: 'What is the query for <request>npm packages that depend on react</request>?' },
        { speaker: 'assistant', text: '<contents>"react"</contents><filters>file:package\\.json</filters>' },

        // file filter with file path regexp
        { speaker: 'human', text: 'What is the query for <request>styles/**.scss files</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>file:styles/**\\.scss</filters>' },

        // file filter with directory regexp
        { speaker: 'human', text: 'What is the query for <request>files in styles directory</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>file:styles</filters>' },

        // file filter with go test files specific directory regexp
        {
            speaker: 'human',
            text: 'what is the query for <request>go test files in the client directory that contain the string "openid" or "auth"</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents>openid OR auth</contents><filters>file:client/ file:_test\\.go$</filters>',
        },

        // file filter with graphql query files specific directory regexp
        {
            speaker: 'human',
            text: 'what is the query for <request>graphql files in the backend directory that includes currentUser: User"</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents>currentUser: User</contents><filters>file:backend/ file:\\.graphql$</filters>',
        },

        // file filter with owner name tag
        { speaker: 'human', text: 'What is the query for <request>owned by ryan@outlook.com</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>file:has.owner(ryan@outlook.com)</filters>' },

        // file filter with exclude owner
        { speaker: 'human', text: 'What is the query for <request>exclude files owned by ryan@outlook.com</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>-file:has.owner(ryan@outlook.com)</filters>' },

        // file filter with owner tag and diff
        { speaker: 'human', text: 'What is the query for <request>changes to go files owned by alice</request>?' },
        {
            speaker: 'assistant',
            text: '<contents></contents><filters>type:diff lang:go file:has.owner(alice)</filters>',
        },

        // file filter with owner tag react storybook specific
        { speaker: 'human', text: 'What is the query for <request>React storybook files owned by alice</request>?' },
        { speaker: 'assistant', text: '<contents>@storybook/react</contents><filters>file:has.owner(alice)</filters>' },

        // file filter with repo
        { speaker: 'human', text: 'What is the query for <request>hooks/use.*\\.tsx files in nextjs repo</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>file:hooks/use.*\\.tsx repo:nextjs</filters>' },

        // file filter with dir and exluding repo
        { speaker: 'human', text: 'What is the query for <request>pages dir excluding vue repo</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>file:pages/ -repo:vue</filters>' },

        // file filter with directory inside repo
        { speaker: 'human', text: 'What is the query for <request>components directory in frontend repo</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>file:components repo:frontend/</filters>' },

        // excluding file fitler
        { speaker: 'human', text: 'What is the query for <request>NewUser excluding go test files</request>?' },
        { speaker: 'assistant', text: '<contents>NewUser</contents><filters>-file:_test\\.go$</filters>' },

        // both including and excluding file filter
        {
            speaker: 'human',
            text: 'What is the query for <request>all scss files in styles folder excluding global\\.scss file</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents></contents><filters>file:styles/.*\\.scss$ -file:global\\.scss lang:scss</filters>',
        },

        // diff filter
        { speaker: 'human', text: 'What is the query for <request>changes to authentication</request>?' },
        { speaker: 'assistant', text: '<contents>authentication</contents><filters>type:diff</filters>' },

        // diff filter for repo with commit tag
        {
            speaker: 'human',
            text: 'What is the query for <request>diff under commit 6k3j5ip on tensorflow repository</request>?',
        },
        { speaker: 'assistant', text: '<contents></contents><filters>repo:tensorflow@6k3j5ip type:diff</filters>' },

        // diff filter with nested file
        {
            speaker: 'human',
            text: 'What is the query for <request>changes made to .btn in global\\.css in styles directory</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents>.btn</contents><filters>type:diff lang:css file:styles/.*global\\.css</filters>',
        },

        // diff filter with language
        { speaker: 'human', text: 'What is the query for <request>diff for rust files</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>type:diff lang:rust</filters>' },

        // diff filter with file directory
        { speaker: 'human', text: 'What is the query for <request>changes inside config/aws directory</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>type:diff file:config/aws/</filters>' },

        // diff filter with file directory and query
        {
            speaker: 'human',
            text: 'What is the query for <request>changes to .button under styles directory</request>?',
        },
        { speaker: 'assistant', text: '<contents>.button</contents><filters>type:diff file:styles/</filters>' },

        // diff filter for package.json with before
        {
            speaker: 'human',
            text: 'What is the query for <request>node packages changes before last friday</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents></contents><filters>type:diff file:package\\.json before:"last friday"</filters>',
        },

        // diff filter for symbol with before
        {
            speaker: 'human',
            text: 'What is the query for <request>changes to useAuthenticatedUser before september 17 1998</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents>useAuthenticatedUser</contents><filters>type:diff before:"september 17 1998"</filters>',
        },

        // diff filter for nested file with after
        {
            speaker: 'human',
            text: 'What is the query for <request>diffs for file init\\.ts inside server dir after 3 weeks ago</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents></contents><filters>type:diff file:server/.*init\\.ts after:"3 weeks ago"</filters>',
        },

        // diff filter for symbol with after
        {
            speaker: 'human',
            text: 'What is the query for <request>changes to useAuthenticatedUser after september 17 1998</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents>useAuthenticatedUser</contents><filters>type:diff after:"september 17 1998"</filters>',
        },

        // diff filter with owner
        { speaker: 'human', text: 'What is the query for <request>changes to files owned by jordan</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>type:diff file.has.owner(jordan)</filters>' },

        // author filter with pattern
        {
            speaker: 'human',
            text: 'What is the query for <request>modifications made to go routes file by mrnugget</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents></contents><filters>file:routes\\.go type:diff author:"mrnugget"</filters>',
        },

        // author filter with file and diff
        {
            speaker: 'human',
            text: 'What is the query for <request>changes to schema.graphql file by Rahul Gandhi <rahul@gmail.com></request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents></contents><filters>file:schema.graphql type:diff author:"Rahul Gandhi <rahul@gmail.com>"</filters>',
        },

        // author filter with search and diff
        { speaker: 'human', text: 'What is the query for <request>changes to useLocalStorage by @ryan</request>?' },
        { speaker: 'assistant', text: '<contents>useLocalStorage</contents><filters>type:diff author:@ryan</filters>' },

        // author filter with me in nested dir
        {
            speaker: 'human',
            text: 'What is the query for <request>changes made to go files in graphqlbackend dir by me</request>?',
        },
        {
            speaker: 'assistant',
            text: `contents></contents><filters>file:graphqlbackend/.*\\.go type:diff author:@${
                user?.username || 'anonymous'
            }</filters>`,
        },

        // including authors from domain but excluding myself
        {
            speaker: 'human',
            text: 'What is the query for <request>changes made to schema.graphql by users having email domain gmail\\.com excluding myself</request>?',
        },
        {
            speaker: 'assistant',
            text: `contents></contents><filters>file:schema.graphql type:diff author:"@gmail.com$" -author:@${
                user?.username || 'anonymous'
            }</filters>`,
        },

        // commits by authors from domain but excluding pattern
        {
            speaker: 'human',
            text: 'What is the query for <request>commits by users from domain outlook\\.com excluding @kyle</request>?',
        },
        {
            speaker: 'assistant',
            text: 'contents></contents><filters>type:commit author:"@outlook.com$" -author:@kyle</filters>',
        },

        // commit filter with message
        { speaker: 'human', text: 'What is the query for <request>bugfix commits</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>type:commit message:"bugfix"</filters>' },

        // latest commits correction response, otherwise it adds order:desc on its own
        { speaker: 'human', text: 'What is the query for <request>latest commits by @freddy</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>type:commit author:@freddy</filters>' },

        // symbol filter with query
        { speaker: 'human', text: 'What is the query for <request>isAuthenticated symbol</request>?' },
        { speaker: 'assistant', text: '<contents>isAuthenticated</contents><filters>type:symbol</filters>' },

        // symbol filter with filePath
        { speaker: 'human', text: 'What is the query for <request>getUserByID symbol in models/users\\.rb</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>getUserByID</contents><filters>file:models/users\\.rb type:symbol</filters>',
        },

        // symbol filter with repo
        { speaker: 'human', text: 'What is the query for <request>Auth symbol in rails repo</request>?' },
        { speaker: 'assistant', text: '<contents>Auth</contents><filters>repo:rails type:symbol</filters>' },

        // symbol filter with excluding repo
        { speaker: 'human', text: 'What is the query for <request>passport symbol exluding node repo</request>?' },
        { speaker: 'assistant', text: '<contents>passport</contents><filters>-repo:node type:symbol</filters>' },

        // symbol filter with both including and excluding repo
        {
            speaker: 'human',
            text: 'What is the query for <request>Middleware symbol in express repo exluding node repo</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents>Middleware</contents><filters>repo:express -repo:node type:symbol</filters>',
        },

        // symbol filter vs diff filter
        {
            speaker: 'human',
            text: 'What is the query for <request>changes to login symbol in javascript file</request>?',
        },
        { speaker: 'assistant', text: '<contents>login</contents><filters>type:diff lang:javascript</filters>' },

        // language fitler
        { speaker: 'human', text: 'What is the query for <request>typescript files</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>lang:go</filters>' },

        // language fitler with query
        { speaker: 'human', text: 'What is the query for <request>golang oauth</request>?' },
        { speaker: 'assistant', text: '<contents>oauth</contents><filters>lang:go</filters>' },

        // language fitler with react hook specific query
        {
            speaker: 'human',
            text: 'What is the query for <request>TypeScript files that define a React hook</request>',
        },
        {
            speaker: 'assistant',
            text: '<contents>^export (const|function) use\\w+</contents><filters>lang:typescript</filters>',
        },

        // language fitler with react class component specific query
        { speaker: 'human', text: 'What is the query for <request>react class components</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>class \\w+ extends React\\.Component</contents><filters>(lang:typescript OR lang:javascript)</filters>',
        },

        // language fitler with repo
        { speaker: 'human', text: 'What is the query for <request>python files in numpy repo</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>lang:python repo:numpy</filters>' },

        // language filter with excluding repo
        { speaker: 'human', text: 'What is the query for <request>scss files excluding tailwind repo</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>lang:scss -repo:tailwind</filters>' },

        // select repo with files filter
        { speaker: 'human', text: 'What is the query for <request>repos with node_modules</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>select:repo file:node_modules/</filters>' },

        // select repo with files filter
        { speaker: 'human', text: 'What is the query for <request>repository with \\.github</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>select:repo file:\\.github/</filters>' },

        // select repo with text
        { speaker: 'human', text: 'What is the query for <request>repositories with powered by cody</request>?' },
        { speaker: 'assistant', text: '<contents>powered by cody</contents><filters>select:repo</filters>' },

        // select repo with symbol search
        { speaker: 'human', text: 'What is the query for <request>repositories using Sentry symbol</request>?' },
        { speaker: 'assistant', text: '<contents>Sentry</contents><filters>select:repo type:symbol</filters>' },

        // select repo with commit author
        { speaker: 'human', text: 'What is the query for <request>repos with commits by mark</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>select:repo type:commit author:"mark"</filters>' },

        // select repo with name
        { speaker: 'human', text: 'What is the query for <request>google/* repos</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>select:repo repo:google/*</filters>' },

        // select owners with files filter
        { speaker: 'human', text: 'What is the query for <request>owners of schema\\.graphql</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>.*</contents><filters>select:file.owners file:schema\\.graphql/</filters>',
        },

        // select owners with text
        { speaker: 'human', text: 'What is the query for <request>who owns files with useLocalStorage</request>?' },
        { speaker: 'assistant', text: '<contents>useLocalStorage</contents><filters>select:file.owners</filters>' },

        // select owners with symbol search
        {
            speaker: 'human',
            text: 'What is the query for <request>file owners having DashboardLayout symbol</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents>DashboardLayout</contents><filters>select:file.owners type:symbol</filters>',
        },

        // select owners with revision filter
        {
            speaker: 'human',
            text: 'What is the query for <request>owners of files changed under commit 4569f55</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents>.*</contents><filters>select:file.owners type:commit rev:4569f55"</filters>',
        },

        // select owners with files filter
        { speaker: 'human', text: 'What is the query for <request>people owning enterprise/utils</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>.*</contents><filters>select:file.owners file:enterprise/utils</filters>',
        },

        // select owners with repo filter
        { speaker: 'human', text: 'What is the query for <request>owners of lodash repo</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>select:file.owners repo:lodash</filters>' },

        // select owners with file & repo filter
        { speaker: 'human', text: 'What is the query for <request>owners of assets dir in adminui repo</request>?' },
        {
            speaker: 'assistant',
            text: '<contents></contents><filters>select:file.owners file:assets/ repo:adminui</filters>',
        },

        // select files with repo filter
        { speaker: 'human', text: 'What is the query for <request>api/* files in microservice repo</request>?' },
        {
            speaker: 'assistant',
            text: '<contents></contents><filters>select:file file:api/* repo:microservice</filters>',
        },

        // select python files
        { speaker: 'human', text: 'What is the query for <request>python files</request>?' },
        { speaker: 'assistant', text: '<contents>powered by cody</contents><filters>select:file file:\\.py</filters>' },

        // select files with symbol search
        { speaker: 'human', text: 'What is the query for <request>files using useEffect(</request>?' },
        { speaker: 'assistant', text: '<contents>useEffect(</contents><filters>select:file type:symbol</filters>' },

        // select file with rev
        {
            speaker: 'human',
            text: 'What is the query for <request>files changes by commit 4569455 in aws repo</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents></contents><filters>select:file type:diff repo:aws rev:4569455</filters>',
        },

        // select removed anything
        { speaker: 'human', text: 'What is the query for <request>removals from migration\\.sql file</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>.*</contents><filters>select:commit.diff.removed type:diff file:migration\\.sql</filters>',
        },

        // select removed anything with author filter
        { speaker: 'human', text: 'What is the query for <request>lines removed by @kyler</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>.*</contents><filters>select:commit.diff.removed type:diff author:@kyler</filters>',
        },

        // select removed symbol
        { speaker: 'human', text: 'What is the query for <request>removed usage of toUpperCase</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>toUpperCase</contents><filters>select:commit.diff.removed type:diff</filters>',
        },

        // select removed text with repo
        {
            speaker: 'human',
            text: 'What is the query for <request>removed call to model.fit in deeplearning repo</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents>model.fit</contents><filters>select:commit.diff.removed type:diff repo:deeplearning</filters>',
        },

        // select removed text with file
        {
            speaker: 'human',
            text: 'What is the query for <request>removed reference to AuthenticatedUser in graphql/api</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents>AuthenticatedUser</contents><filters>select:commit.diff.removed type:diff file:graphql/api</filters>',
        },

        // select removed with who query correction
        { speaker: 'human', text: 'What is the query for <request>who removed <Form /></request>?' },
        {
            speaker: 'assistant',
            text: '<contents><Form /></contents><filters>select:commit.diff.removed type:diff</filters>',
        },

        // select changed with who query correction
        { speaker: 'human', text: 'What is the query for <request>who changed UserAvatar typescript class</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>UserAvatar</contents><filters>type:diff lang:typescript</filters>',
        },

        // select added symbol
        { speaker: 'human', text: 'What is the query for <request>added parseDate</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>parseDate</contents><filters>select:commit.diff.added type:diff</filters>',
        },

        // select added as question
        { speaker: 'human', text: 'What is the query for <request>when was GitubOAuth added</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>GithubOAuth</contents><filters>select:commit.diff.added type:diff</filters>',
        },

        // select added text with repo
        { speaker: 'human', text: 'What is the query for <request>added call to df.sort in nlp repo</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>df.sort</contents><filters>select:commit.diff.added type:diff repo:nlp</filters>',
        },

        // select added text with file
        {
            speaker: 'human',
            text: 'What is the query for <request>added reference to Cookies in src/middleware</request>?',
        },
        {
            speaker: 'assistant',
            text: '<contents>Cookies</contents><filters>select:commit.diff.added type:diff file:src/middleware</filters>',
        },

        // select added anything
        { speaker: 'human', text: 'What is the query for <request>additions made to global\\.scss</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>.*</contents><filters>select:commit.diff.added type:diff file:global\\.scss</filters>',
        },

        // select added anything with author filter
        { speaker: 'human', text: 'What is the query for <request>additions made by @kyler</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>.*</contents><filters>select:commit.diff.added type:diff author:@kyler</filters>',
        },

        // select added with who query correction
        { speaker: 'human', text: 'What is the query for <request>who added UserAvatar</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>UserAvatar</contents><filters>select:commit.diff.added type:diff</filters>',
        },

        // select symbol any function
        { speaker: 'human', text: 'What is the query for <request>functions in typescript utils file</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>.*</contents><filters>select:symbol.function type:symbol lang:typescript file:utils</filters>',
        },

        // select symbol function
        { speaker: 'human', text: 'What is the query for <request>functions with name like is_.*?$</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>is_.*$</contents><filters>select:symbol.function type:symbol</filters>',
        },

        // select symbol function with lang
        { speaker: 'human', text: 'What is the query for <request>user related functions in c++</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>user</contents><filters>select:symbol.function type:symbol lang:c++</filters>',
        },

        // select symbol function with lang
        { speaker: 'human', text: 'What is the query for <request>parse function in python</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>parse</contents><filters>select:symbol.function type:symbol lang:python</filters>',
        },

        // select symbol function starts with
        { speaker: 'human', text: 'What is the query for <request>functions starting with Get</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>^Get</contents><filters>select:symbol.function type:symbol</filters>',
        },

        // select symbol function ends with
        { speaker: 'human', text: 'What is the query for <request>functions ending with Delete</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>Delete$</contents><filters>select:symbol.function type:symbol</filters>',
        },

        // select symbol interface
        { speaker: 'human', text: 'What is the query for <request>IProps interfaces</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>IProps</contents><filters>select:symbol.interface type:symbol</filters>',
        },

        // select symbol interface with lang
        { speaker: 'human', text: 'What is the query for <request>IProps interfaces in typescript</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>IProps</contents><filters>select:symbol.interface type:symbol lang:typescript</filters>',
        },

        // select symbol graphql enums
        { speaker: 'human', text: 'What is the query for <request>enums in graphql schema</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>.*</contents><filters>select:symbol.enum type:symbol file:schema\\.graphql</filters>',
        },

        // select symbol enums in go lang
        { speaker: 'human', text: 'What is the query for <request>go lang enums</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>.*</contents><filters>select:symbol.enum type:symbol lang:golang</filters>',
        },

        // select symbol class with name like
        { speaker: 'human', text: 'What is the query for <request>java class with name like OAuth</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>OAuth</contents><filters>select:symbol.class type:symbol lang:java</filters>',
        },

        // select symbol function with diff correction
        { speaker: 'human', text: 'What is the query for <request>GetUserInfo function changed</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>GetUserInfo</contents><filters>type:diff</filters>',
        },

        // select symbol class with diff and lang correction
        { speaker: 'human', text: 'What is the query for <request>typescript class added</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>.*</contents><filters>select:commit.diff.added type:diff lang:typescript</filters>',
        },

        // select symbol method removed correction
        { speaker: 'human', text: 'What is the query for <request>toString method removed</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>toString</contents><filters>select:commit.diff.removed type:diff</filters>',
        },

        // and operator
        { speaker: 'human', text: 'What is the query for <request>files containing bat and ball</request>?' },
        { speaker: 'assistant', text: '<contents>bat AND ball</contents>' },

        // and operator with lang filter
        { speaker: 'human', text: 'What is the query for <request>python files containing bat and ball</request>?' },
        { speaker: 'assistant', text: '<contents>bat AND ball</contents><filters>lang:python</filters>' },

        // or operator
        { speaker: 'human', text: 'What is the query for <request>fruits or vegetables</request>?' },
        { speaker: 'assistant', text: '<contents>fruits AND vegetables</contents>' },

        // or operator with lang filter
        { speaker: 'human', text: 'What is the query for <request>css files with both media or position</request>?' },
        { speaker: 'assistant', text: '<contents>media OR position</contents><filters>lang:css/</filters>' },

        // fallback
        { speaker: 'human', text: 'What is the query for <request>DFH84fHAg</request>?' },
        {
            speaker: 'assistant',
            text: 'I apologize, but I do not understand the request "DFH84fHAg". Without more context about what is being requested, I cannot generate a valid query.',
        },
        {
            speaker: 'human',
            text: 'NEVER ASK FOR MORE CONTEXT and ALWAYS MAKE A GUESS. If you are unsure, just treat the entire request as a regular expression matching file contents. What is the query for <request>DFH84fHAg</request>?',
        },
        { speaker: 'assistant', text: '<contents>DFH84fHAg</contents><filters></filters>' },

        // actual user query
        { speaker: 'human', text: `What is the query for <request>${input}</request>?` },
        { speaker: 'assistant', text: '' },
    ]
}
