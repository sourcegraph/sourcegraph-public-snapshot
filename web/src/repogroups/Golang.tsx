import { CodeHosts, RepogroupMetadata } from './types'
import { SearchPatternType } from '../../../shared/src/graphql/schema'
import * as React from 'react'

export const golang: RepogroupMetadata = {
    title: 'Golang',
    name: 'golang',
    url: '/golang',
    repositories: [
        { name: 'golang/go', codehost: CodeHosts.GITHUB },
        { name: 'kubernetes/kubernetes', codehost: CodeHosts.GITHUB },
        { name: 'moby/moby', codehost: CodeHosts.GITHUB },
        { name: 'avelino/awesome-go', codehost: CodeHosts.GITHUB },
        { name: 'gohugoio/hugo', codehost: CodeHosts.GITHUB },
        { name: 'gin-gonic/gin', codehost: CodeHosts.GITHUB },
        { name: 'fatedier/frp', codehost: CodeHosts.GITHUB },
        { name: 'astaxie/build-web-application-with-golang', codehost: CodeHosts.GITHUB },
        { name: 'gogs/gogs', codehost: CodeHosts.GITHUB },
        { name: 'v2ray/v2ray-core', codehost: CodeHosts.GITHUB },
        { name: 'syncthing/syncthing', codehost: CodeHosts.GITHUB },
        { name: 'etcd-io/etcd', codehost: CodeHosts.GITHUB },
        { name: 'prometheus/prometheus', codehost: CodeHosts.GITHUB },
        { name: 'junegunn/fzf', codehost: CodeHosts.GITHUB },
        { name: 'containous/traefik', codehost: CodeHosts.GITHUB },
        { name: 'caddyserver/caddy', codehost: CodeHosts.GITHUB },
        { name: 'ethereum/go-ethereum', codehost: CodeHosts.GITHUB },
        { name: 'FiloSottile/mkcert', codehost: CodeHosts.GITHUB },
        { name: 'astaxie/beego', codehost: CodeHosts.GITHUB },
        { name: 'iikira/BaiduPCS-Go', codehost: CodeHosts.GITHUB },
        { name: 'pingcap/tidb', codehost: CodeHosts.GITHUB },
        { name: 'istio/istio', codehost: CodeHosts.GITHUB },
        { name: 'hashicorp/terraform', codehost: CodeHosts.GITHUB },
        { name: 'minio/minio', codehost: CodeHosts.GITHUB },
        { name: 'rclone/rclone', codehost: CodeHosts.GITHUB },
        { name: 'unknwon/the-way-to-go_ZH_CN', codehost: CodeHosts.GITHUB },
        { name: 'drone/drone', codehost: CodeHosts.GITHUB },
        { name: 'wagoodman/dive', codehost: CodeHosts.GITHUB },
        { name: 'go-gitea/gitea', codehost: CodeHosts.GITHUB },
        { name: 'github/hub', codehost: CodeHosts.GITHUB },
        { name: 'hashicorp/consul', codehost: CodeHosts.GITHUB },
        { name: 'influxdata/influxdb', codehost: CodeHosts.GITHUB },
        { name: 'inconshreveable/ngrok', codehost: CodeHosts.GITHUB },
        { name: 'jinzhu/gorm', codehost: CodeHosts.GITHUB },
        { name: 'kubernetes/minikube', codehost: CodeHosts.GITHUB },
        { name: 'mattermost/mattermost-server', codehost: CodeHosts.GITHUB },
        { name: 'cockroachdb/cockroach', codehost: CodeHosts.GITHUB },
        { name: 'kataras/iris', codehost: CodeHosts.GITHUB },
        { name: 'nsqio/nsq', codehost: CodeHosts.GITHUB },
        { name: 'openfaas/faas', codehost: CodeHosts.GITHUB },
        { name: 'labstack/echo', codehost: CodeHosts.GITHUB },
        { name: 'helm/helm', codehost: CodeHosts.GITHUB },
        { name: 'go-kit/kit', codehost: CodeHosts.GITHUB },
        { name: 'spf13/cobra', codehost: CodeHosts.GITHUB },
        { name: 'yeasy/docker_practice', codehost: CodeHosts.GITHUB },
        { name: 'jesseduffield/lazygit', codehost: CodeHosts.GITHUB },
        { name: 'hashicorp/vault', codehost: CodeHosts.GITHUB },
        { name: 'jesseduffield/lazydocker', codehost: CodeHosts.GITHUB },
        { name: 'sirupsen/logrus', codehost: CodeHosts.GITHUB },
        { name: 'joewalnes/websocketd', codehost: CodeHosts.GITHUB },
        { name: 'tsenart/vegeta', codehost: CodeHosts.GITHUB },
        { name: 'rancher/rancher', codehost: CodeHosts.GITHUB },
        { name: 'go-delve/delve', codehost: CodeHosts.GITHUB },
        { name: 'yudai/gotty', codehost: CodeHosts.GITHUB },
        { name: 'urfave/cli', codehost: CodeHosts.GITHUB },
        { name: 'cayleygraph/cayley', codehost: CodeHosts.GITHUB },
        { name: 'golang/dep', codehost: CodeHosts.GITHUB },
        { name: 'zyedidia/micro', codehost: CodeHosts.GITHUB },
        { name: 'helm/charts', codehost: CodeHosts.GITHUB },
        { name: 'dgraph-io/dgraph', codehost: CodeHosts.GITHUB },
        { name: 'micro/go-micro', codehost: CodeHosts.GITHUB },
        { name: 'buger/goreplay', codehost: CodeHosts.GITHUB },
        { name: 'rancher/k3s', codehost: CodeHosts.GITHUB },
        { name: 'tmrts/go-patterns', codehost: CodeHosts.GITHUB },
        { name: 'chai2010/advanced-go-programming-book', codehost: CodeHosts.GITHUB },
        { name: 'coreybutler/nvm-windows', codehost: CodeHosts.GITHUB },
        { name: 'valyala/fasthttp', codehost: CodeHosts.GITHUB },
        { name: 'spf13/viper', codehost: CodeHosts.GITHUB },
        { name: 'ehang-io/nps', codehost: CodeHosts.GITHUB },
        { name: 'gorilla/websocket', codehost: CodeHosts.GITHUB },
        { name: 'gorilla/mux', codehost: CodeHosts.GITHUB },
        { name: 'xtaci/kcptun', codehost: CodeHosts.GITHUB },
        { name: 'goharbor/harbor', codehost: CodeHosts.GITHUB },
        { name: 'revel/revel', codehost: CodeHosts.GITHUB },
        { name: 'txthinking/brook', codehost: CodeHosts.GITHUB },
        { name: 'kubernetes/kops', codehost: CodeHosts.GITHUB },
        { name: 'wtfutil/wtf', codehost: CodeHosts.GITHUB },
        { name: 'grpc/grpc-go', codehost: CodeHosts.GITHUB },
        { name: 'julienschmidt/httprouter', codehost: CodeHosts.GITHUB },
        { name: 'CodisLabs/codis', codehost: CodeHosts.GITHUB },
        { name: 'quii/learn-go-with-tests', codehost: CodeHosts.GITHUB },
        { name: 'go-martini/martini', codehost: CodeHosts.GITHUB },
        { name: 'jaegertracing/jaeger', codehost: CodeHosts.GITHUB },
        { name: 'gocolly/colly', codehost: CodeHosts.GITHUB },
        { name: 'fogleman/primitive', codehost: CodeHosts.GITHUB },
        { name: 'google/cadvisor', codehost: CodeHosts.GITHUB },
        { name: 'boltdb/bolt', codehost: CodeHosts.GITHUB },
        { name: 'peterq/pan-light', codehost: CodeHosts.GITHUB },
        { name: 'stretchr/testify', codehost: CodeHosts.GITHUB },
        { name: 'iawia002/annie', codehost: CodeHosts.GITHUB },
        { name: 'hyperledger/fabric', codehost: CodeHosts.GITHUB },
        { name: 'hashicorp/packer', codehost: CodeHosts.GITHUB },
        { name: 'restic/restic', codehost: CodeHosts.GITHUB },
        { name: 'google/grumpy', codehost: CodeHosts.GITHUB },
        { name: 'vitessio/vitess', codehost: CodeHosts.GITHUB },
        { name: 'google/gvisor', codehost: CodeHosts.GITHUB },
        { name: 'bcicen/ctop', codehost: CodeHosts.GITHUB },
        { name: 'gizak/termui', codehost: CodeHosts.GITHUB },
        { name: 'go-kratos/kratos', codehost: CodeHosts.GITHUB },
        { name: 'uber-go/za', codehost: CodeHosts.GITHUB },
        { name: 'robcalcroft/react-use-lazy-load-image', codehost: CodeHosts.GITHUB },
        { name: 'intercaetera/react-use-message-bar', codehost: CodeHosts.GITHUB },
        { name: 'wowlusitong/react-use-modal', codehost: CodeHosts.GITHUB },
        { name: 'zhangkaiyulw/react-use-path', codehost: CodeHosts.GITHUB },
        { name: 'neo/react-use-scroll-position', codehost: CodeHosts.GITHUB },
        { name: 'lessmess-dev/react-use-trigger', codehost: CodeHosts.GITHUB },
        { name: 'perlin-network/react-use-wavelet', codehost: CodeHosts.GITHUB },
        { name: 'streamich/react-use', codehost: CodeHosts.GITHUB },
        { name: 'GeDiez/react-use-formless', codehost: CodeHosts.GITHUB },
        { name: 'venil7/react-usemiddleware', codehost: CodeHosts.GITHUB },
        { name: 'alex-cory/react-useportal', codehost: CodeHosts.GITHUB },
        { name: 'vardius/react-user-media', codehost: CodeHosts.GITHUB },
        { name: 'f/react-wait', codehost: CodeHosts.GITHUB },
        { name: 'AvraamMavridis/react-window-communication-hook', codehost: CodeHosts.GITHUB },
        { name: 'yesmeck/react-with-hooks', codehost: CodeHosts.GITHUB },
        { name: 'mfrachet/reaktion', codehost: CodeHosts.GITHUB },
        { name: 'iusehooks/redhooks', codehost: CodeHosts.GITHUB },
        { name: 'ianobermiller/redux-react-hook', codehost: CodeHosts.GITHUB },
        { name: 'regionjs/region-core', codehost: CodeHosts.GITHUB },
        { name: 'imbhargav5/rehooks-visibility-sensor', codehost: CodeHosts.GITHUB },
        { name: 'pedronasser/resynced', codehost: CodeHosts.GITHUB },
        { name: 'brn/rrh', codehost: CodeHosts.GITHUB },
        { name: 'LeetCode-OpenSource/rxjs-hooks', codehost: CodeHosts.GITHUB },
        { name: 'dejorrit/scroll-data-hook', codehost: CodeHosts.GITHUB },
        { name: 'style-hook/style-hook', codehost: CodeHosts.GITHUB },
        { name: 'vercel/swr', codehost: CodeHosts.GITHUB },
        { name: 'jaredpalmer/the-platform', codehost: CodeHosts.GITHUB },
        { name: 'danieldelcore/trousers', codehost: CodeHosts.GITHUB },
        { name: 'mauricedb/use-abortable-fetch', codehost: CodeHosts.GITHUB },
        { name: 'awmleer/use-action', codehost: CodeHosts.GITHUB },
        { name: 'awmleer/use-async-memo', codehost: CodeHosts.GITHUB },
        { name: 'lowewenzel/use-autocomplete', codehost: CodeHosts.GITHUB },
        { name: 'sergey-s/use-axios-react', codehost: CodeHosts.GITHUB },
        { name: 'zcallan/use-browser-history', codehost: CodeHosts.GITHUB },
        { name: 'samjbmason/use-cart', codehost: CodeHosts.GITHUB },
        { name: 'CharlesStover/use-clippy', codehost: CodeHosts.GITHUB },
        { name: 'dai-shi/use-context-selector', codehost: CodeHosts.GITHUB },
        { name: 'oktaysenkan/use-countries', codehost: CodeHosts.GITHUB },
        { name: 'xnimorz/use-debounce', codehost: CodeHosts.GITHUB },
        { name: 'sandiiarov/use-deep-compare', codehost: CodeHosts.GITHUB },
        { name: 'kentcdodds/use-deep-compare-effect', codehost: CodeHosts.GITHUB },
        { name: 'gregnb/use-detect-print', codehost: CodeHosts.GITHUB },
        { name: 'CharlesStover/use-dimensions', codehost: CodeHosts.GITHUB },
        { name: 'zattoo/use-double-click', codehost: CodeHosts.GITHUB },
        { name: 'sandiiarov/use-events', codehost: CodeHosts.GITHUB },
        { name: 'CharlesStover/use-force-update', codehost: CodeHosts.GITHUB },
        { name: 'sandiiarov/use-hotkeys', codehost: CodeHosts.GITHUB },
        { name: 'therealparmesh/use-hovering', codehost: CodeHosts.GITHUB },
        { name: 'ava/use-http', codehost: CodeHosts.GITHUB },
        { name: 'immerjs/use-immer', codehost: CodeHosts.GITHUB },
        { name: 'immerjs/immer', codehost: CodeHosts.GITHUB },
        { name: 'neighborhood999/use-input-file', codehost: CodeHosts.GITHUB },
        { name: 'helderburato/use-is-mounted-ref', codehost: CodeHosts.GITHUB },
        { name: 'davidicus/use-lang-direction', codehost: CodeHosts.GITHUB },
        { name: 'streamich/use-media', codehost: CodeHosts.GITHUB },
        { name: 'dimitrinicolas/use-mouse-action', codehost: CodeHosts.GITHUB },
        { name: 'jschloer/use-multiselect', codehost: CodeHosts.GITHUB },
        { name: 'wellyshen/use-places-autocomplete', codehost: CodeHosts.GITHUB },
        { name: 'sandiiarov/use-popper', codehost: CodeHosts.GITHUB },
        { name: 'alex-cory/use-react-modal', codehost: CodeHosts.GITHUB },
        { name: 'CharlesStover/use-react-router', codehost: CodeHosts.GITHUB },
        { name: 'tedstoychev/use-reactive-state', codehost: CodeHosts.GITHUB },
        { name: 'dai-shi/use-reducer-async', codehost: CodeHosts.GITHUB },
        { name: 'flepretre/use-redux', codehost: CodeHosts.GITHUB },
        { name: 'tudorgergely/use-scroll-to-bottom', codehost: CodeHosts.GITHUB },
        { name: 'sandiiarov/use-simple-undo', codehost: CodeHosts.GITHUB },
        { name: 'mfrachet/use-socketio', codehost: CodeHosts.GITHUB },
        { name: 'iamgyz/use-socket.io-client', codehost: CodeHosts.GITHUB },
        { name: 'kmoskwiak/useSSE', codehost: CodeHosts.GITHUB },
        { name: 'alex-cory/use-ssr', codehost: CodeHosts.GITHUB },
        { name: 'haydn/use-state-snapshots', codehost: CodeHosts.GITHUB },
        { name: 'philipp-spiess/use-substate', codehost: CodeHosts.GITHUB },
        { name: 'octet-stream/use-suspender', codehost: CodeHosts.GITHUB },
        { name: 'streamich/use-t', codehost: CodeHosts.GITHUB },
        { name: 'homerchen19/use-undo', codehost: CodeHosts.GITHUB },
        { name: 'donavon/use-dark-mode', codehost: CodeHosts.GITHUB },
        { name: 'KATT/use-is-typing', codehost: CodeHosts.GITHUB },
        { name: 'pranesh239/use-key-capture', codehost: CodeHosts.GITHUB },
        { name: 'tranbathanhtung/usePosition', codehost: CodeHosts.GITHUB },
        { name: 'Tweries/useReducerWithLocalStorage', codehost: CodeHosts.GITHUB },
        { name: 'pankod/react-hooks-screen-type', codehost: CodeHosts.GITHUB },
        { name: 'Purii/react-use-scrollspy', codehost: CodeHosts.GITHUB },
        { name: 'JCofman/react-hook-use-service-worker', codehost: CodeHosts.GITHUB },
        { name: 'bboydflo/use-value-after', codehost: CodeHosts.GITHUB },
        { name: 'wednesday-solutions/react-screentype-hoo', codehost: CodeHosts.GITHUB },
    ],
    description: 'Interesting examples of Go.',
    examples: [
        {
            title: 'Search for usages of the Retry-After header in non-vendor Go files:',
            exampleQuery: (
                <>
                    <span className="repogroup-page__keyword-text">file:</span>.go{' '}
                    <span className="repogroup-page__keyword-text">-file:</span>vendor/ Retry-After
                </>
            ),
            patternType: SearchPatternType.literal,
            rawQuery: 'file:.go -file:vendor/ Retry-After',
        },
        {
            title: 'Find examples of sending JSON in a HTTP POST request:',
            exampleQuery: (
                <>
                    repogroup:goteam <span className="repogroup-page__keyword-text">file:</span>.go http.Post json
                </>
            ),
            rawQuery: 'repogroup:goteam file:.go http.Post json',
            patternType: SearchPatternType.literal,
        },
        {
            title: 'Find error handling examples in Go',
            exampleQuery: (
                <>
                    {'if err != nil {:[_]}'} <span className="repogroup-page__keyword-text">lang:</span>go
                </>
            ),
            rawQuery: 'if err != nil {:[_]} lang:go',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Find usage examples of cmp.Diff with options',
            exampleQuery: (
                <>
                    <span className="repogroup-page__keyword-text">lang:go</span> cmp.Diff(:[_], :[_], :[opts])
                </>
            ),
            rawQuery: 'lang:go cmp.Diff(:[_], :[_], :[opts])',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Find examples for setting timeouts on http.Transport',
            exampleQuery: (
                <>
                    {'http.Transport{:[_], MaxIdleConns: :[idleconns], :[_]}'}{' '}
                    <span className="repogroup-page__keyword-text">-file:</span>vendor{' '}
                    <span className="repogroup-page__keyword-text">lang:</span>go
                </>
            ),
            rawQuery: 'http.Transport{:[_], MaxIdleConns: :[idleconns], :[_]} -file:vendor lang:go',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Find examples of Switch statements in Go',
            exampleQuery: (
                <>
                    {'switch :[_] := :[_].(type) { :[string] }'}{' '}
                    <span className="repogroup-page__keyword-text">lang:</span>go{' '}
                    <span className="repogroup-page__keyword-text">count:</span>1000
                </>
            ),
            rawQuery: 'switch :[_] := :[_].(type) { :[string] } lang:go count:1000',
            patternType: SearchPatternType.structural,
        },
    ],
    homepageDescription: 'Interesting examples of Go.',
    homepageIcon: 'https://code.benco.io/icon-collection/logos/go-lang.svg',
}
