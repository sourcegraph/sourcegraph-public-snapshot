# ブラウザーから

ブラウザーで[http://css.github.com/csso/csso.html](http://css.github.com/csso/csso.html)を開く。

**CSSOはブラウザーでの動作を保証していません。このツールを使い方の推奨はコマンドラインまたはnpmモジュールから使用する方法です。**

# コマンドラインから

gitからインストールした場合 `bin/csso` 、nodejs 0.4.x&nbsp;— [http://nodejs.org](http://nodejs.org)をインストールしておく必要があります。

npmからインストールした場合 `csso`

使用方法:

    csso
        使用方法の表示
    csso <filename>
        <filename> のCSSを縮小化し結果を標準出力する
    csso <in_filename> <out_filename>
    csso -i <in_filename> -o <out_filename>
    csso --input <in_filename> --output <out_filename>
        <in_filename> のCSSを縮小化し <out_filename> に出力する
    csso -off
    csso --restructure-off
        構造的な最適化を行わない
    csso -h
    csso --help
        使用方法の表示
    csso -v
    csso --version
        バージョンナンバーの表示

使用例:

    $ echo ".test { color: red; color: green }" > test.css
    $ csso test.css
    .test{color:green}

# npmモジュールから

サンプル (`test.js`):
```js
    var csso = require('csso'),
        css = '.test, .test { color: rgb(255, 255, 255) }';

    console.log(csso.justDoIt(css));
```
出力結果 (`> node test.js`):
```css
    .test{color:#fff}
```
`csso.justDoIt(css, true)` を使用すれば構造的な最適化を行わないようにすることできます。
