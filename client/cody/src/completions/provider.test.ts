import { sliceUntilFirstNLinesOfSuffixMatch } from './provider'

describe('sliceUntilFirstNLinesOfSuffixMatch', () => {
    it('returns the right text', () => {
        const suggestion = `foo\nbar\nbaz\nline-1\nline-2\noh no\nline-1\nline-2\nline-3`
        const suffix = `line-1\nline-2\nline-3\nline-4\nline-5`

        expect(sliceUntilFirstNLinesOfSuffixMatch(suggestion, suffix, 3)).toMatchInlineSnapshot(`
            "foo
            bar
            baz
            line-1
            line-2
            oh no"
        `)
    })

    it('works with the example suggested by Cody', () => {
        const suggestion = 'foo\nbar\nbaz\nqux\nquux'
        const suffix = 'baz\nqux\nquux'

        expect(sliceUntilFirstNLinesOfSuffixMatch(suggestion, suffix, 3)).toMatchInlineSnapshot(`
            "foo
            bar"
        `)
    })

    it('works with a real-world suggestion', () => {
        const suggestion =
            ' \n]; \n\nexport default function LogKit() {\n\n  const data = useLoaderData(); \n\n  return (\n    <div className="h-screen w-full flex bg-zinc-900 [color-scheme:dark] text-white">\n      <div className="w-[500px] bg-zinc-900 border-r border-zinc-800 p-4"></div>\n      <div className="flex-1"> \n        <div className="w-full border-b border-zinc-800 h-12 flex items-center p-4"> \n          <h2 className="font-semibold text-lg">Event Log</h2> \n        </div> \n        <div> \n          <table className="w-full font-mono text-xs"> \n            <thead>\n              <tr> \n                <th>Event</th> \n                <th>Payload</th> \n                <th>Timestamp</th> \n              </tr>'
        const suffix =
            '\n];\n\nexport async function action({ request, context }: ActionArgs) {\n  const payload: any = await request.json();\n\n  data.push({ ...payload, timestamp: Date.now() });\n\n  return null;\n}\n\nexport async function loader() {\n  return json(data);\n}\n\nexport default function LogKit() {\n  const data = useLoaderData();\n\n  return (\n    <div className="h-screen w-full flex bg-zinc-900 [color-scheme:dark] text-white">\n      <div className="w-[500px] bg-zinc-900 border-r border-zinc-800 p-4"></div>\n      <div className="flex-1">\n        <div className="w-full border-b border-zinc-800 h-12 flex items-center p-4">\n          <h2 className="font-semibold text-lg">Event Log</h2>\n        </div>\n        <div>\n          <table className="w-full font-mono text-xs">'

        expect(sliceUntilFirstNLinesOfSuffixMatch(suggestion, suffix, 5)).toMatchInlineSnapshot()
    })
})
