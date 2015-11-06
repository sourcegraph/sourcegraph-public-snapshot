CSSO (CSSオプティマイザー)は、他とは違ったCSS縮小化ツールです。 一般的な縮小化テクニックに加えてCSSファイルの構造的な最適化も行うので、 他の縮小化ツールと比べてより軽量にすることが可能です。

# 1. 縮小化

縮小化というのはCSSファイルをより軽量なサイズに、それも不具合無く、変換するプロセスのことを言います。基本的なテクニックに関しては以下のとおりです。

* 不必要な要素（例：末尾のセミコロン）の削除や、値をよりコンパクトな表記に変更（例： `0px` を `0` に）するといったような基本的な変換
* 上書きされたプロパティの削除やブロックのマージなどのような構造的な変換

## 1.1. 基本的な変換

### 1.1.1. ホワイトスペースの削除

このようなホワイトスペース (` `, `\n`, `\r`, `\t`, `\f`) はレンダリングに影響与えないため必要ありません。

* 変換前:
```css
        .test
        {
            margin-top: 1em;

            margin-left  : 2em;
        }
```

* 変換後:
```css
        .test{margin-top:1em;margin-left:2em}
```

これ以降の例に関しては、可読性を考慮してホワイトスペースを残したままにしておきます。

### 1.1.2. 末尾の ';'　削除

最後のセミコロンは必要ではなく、レンダリングに影響与えません。

* 変換前:
```css
        .test {
            margin-top: 1em;;
        }
```

* 変換後:
```css
        .test {
            margin-top: 1em
        }
```

### 1.1.3. コメントの削除

コメントはレンダリングに影響与えません: \[[CSS 2.1 / 4.1.9 Comments](http://www.w3.org/TR/CSS21/syndata.html#comments)\].

* 変換前:
```css
        /* comment */

        .test /* comment */ {
            /* comment */ margin-top: /* comment */ 1em;
        }
```

* 変換後:
```css
        .test {
            margin-top: 1em
        }
```

もしコメントを保持しておきたい場合は、最初のコメントだけですが `!` から記述するとコメントが保持されます。

* 変換前:
```css
        /*! MIT license */
        /*! will be removed */

        .test {
            color: red
        }
```

* 変換後:
```css
        /*! MIT license */
        .test {
            color: red
        }
```

### 1.1.4.  不正な @charset と @import 宣言の削除

仕様書によれば、 `@charset` はスタイルシートの先頭に置かなければなりません: \[[CSS 2.1 / 4.4 CSS style sheet representation](http://www.w3.org/TR/CSS21/syndata.html#charset)\].

CSSOはゆるやかにこのルールを適用します - スタイルシートの上部にあり、ホワイトスペースやコメントのすぐ後にある `@charset` を保持します。

\[[CSS 2.1 / 6.3 The @import rule](http://www.w3.org/TR/CSS21/cascade.html#at-import)\] の仕様に従って、間違った場所に置かれた `@import` は削除します。

* 変換前:
```css
        /* comment */
        @charset 'UTF-8';
        @import "test0.css";
        @import "test1.css";
        @charset 'wrong';

        h1 {
            color: red
        }

        @import "wrong";
```

* 変換後:
```css
        @charset 'UTF-8';
        @import "test0.css";
        @import "test1.css";
        h1 {
            color: red
        }
```

### 1.1.5. color プロパティの縮小化

\[[CSS 2.1 / 4.3.6 Colors](http://www.w3.org/TR/CSS21/syndata.html#color-units)\] の仕様に従って、色の値を変縮小化します。

* 変換前:
```css
        .test {
            color: yellow;
            border-color: #c0c0c0;
            background: #ffffff;
            border-top-color: #f00;
            outline-color: rgb(0, 0, 0);
        }
```

* 変換後:
```css
        .test {
            color: #ff0;
            border-color: silver;
            background: #fff;
            border-top-color: red;
            outline-color: #000
        }
```

### 1.1.6. 0 の縮小化

あるケースにおいて、数値は `0` にすることでコンパクトにできますし、ときには削除さえします。

`0%` の値は次のようなケースを考えると縮小化できません。 `rgb(100%, 100%, 0)`

* 変換前:
```css
        .test {
            fakeprop: .0 0. 0.0 000 00.00 0px 0.1 0.1em 0.000em 00% 00.00% 010.00
        }
```

* 変換後:
```css
        .test {
            fakeprop: 0 0 0 0 0 0 .1 .1em 0 0% 0% 10
        }
```

### 1.1.7. 複数行文字列の縮小化

\[[CSS 2.1 / 4.3.7 Strings](http://www.w3.org/TR/CSS21/syndata.html#strings)\] の仕様に従って、複数行文字列は縮小化されます。

* 変換前:
```css
        .test[title="abc\
        def"] {
            background: url("foo/\
        bar")
        }
```

* 変換後:
```css
        .test[title="abcdef"] {
            background: url("foo/bar")
        }
```

### 1.1.8. font-weight プロパティの縮小化

\[[CSS 2.1 / 15.6 Font boldness: the 'font-weight' property](http://www.w3.org/TR/CSS21/fonts.html#font-boldness)\] の仕様に従って、`font-weight` プロパティの `bold` と `normal` は縮小化されます。

* 変換前:
```css
        .test0 {
            font-weight: bold
        }

        .test1 {
            font-weight: normal
        }
```

* 変換後:
```css
        .test0 {
            font-weight: 700
        }

        .test1 {
            font-weight: 400
        }
```

## 1.2. 構造的な最適化

### 1.2.1. 同一セレクタブロックのマージ

同一のセレクタで隣接するブロックはマージされます。

* 変換前:
```css
        .test0 {
            margin: 0
        }

        .test1 {
            border: none
        }

        .test1 {
            background-color: green
        }

        .test0 {
            padding: 0
        }
```

* 変換後:
```css
        .test0 {
            margin: 0
        }

        .test1 {
            border: none;
            background-color: green
        }

        .test0 {
            padding: 0
        }
```

### 1.2.2. ブロック内の同一プロパティのマージ

隣接するブロック内の同一プロパティはマージされます。

* 変換前:
```css
        .test0 {
            margin: 0
        }

        .test1 {
            border: none
        }

        .test2 {
            border: none
        }

        .test0 {
            padding: 0
        }
```

* 変換後:
```css
        .test0 {
            margin: 0
        }

        .test1, .test2 {
            border: none
        }

        .test0 {
            padding: 0
        }
```

### 1.2.3. 上書きされたプロパティの削除

次のルールにより、ブラウザーによって無視されるプロパティは削除されます。:

* もし、`!important` 宣言がなければ、CSSルールないの最後のプロパティが適用されます。
* `!important` が宣言されたプロパティが複数あれば、最後のものが適用されます。

* 変換前:
```css
        .test {
            color: red;
            margin: 0;
            line-height: 3cm;
            color: green;
        }
```

* 変換後:
```css
        .test {
            margin: 0;
            line-height: 3cm;
            color: green
        }
```

#### 1.2.3.1. 上書きされたショートハンドプロパティの削除

`border`, `margin`, `padding`, `font`, `list-style` プロパティの場合、 次の削除ルールが適用されます: もし最後のプロパティが 'general' であれば (例： `border`), すべての先行の上書きされたプロパティは削除されます（例：`border-top-width` または `border-style`)。

* 変換前:
```css
        .test {
            border-top-color: red;
            border-color: green
        }
```

* 変換後:
```css
        .test {
            border-color:green
        }
```

### 1.2.4. 繰り返されているセレクタの削除

繰り返されているセレクタは削除されます。

* 変換前:
```css
        .test, .test {
            color: red
        }
```

* 変換後:
```css
        .test {
            color: red
        }
```

### 1.2.5. ブロックの部分的なマージ

2つの隣接するブロックがあり、片方がもう片方のサブセットの場合、次の最適化が考えられます:

* 重複するプロパティは、ブロックから削除されます。
* ブロックの残りのプロパティは受け手のブロックにコピーされます。

重複プロパティの文字数よりもコピーするプロパティの文字数が少なければ、縮小化が実行されます。

* 変換前:
```css
        .test0 {
            color: red
        }

        .test1 {
            color: red;
            border: none
        }

        .test2 {
            border: none
        }
```

* 変換後:
```css
        .test0, .test1 {
            color: red
        }

        .test1, .test2 {
            border: none
        }
```

重複プロパティの文字数よりもコピーするプロパティの文字数が多いので、縮小化が実行されません。

* 変換前:
```css
        .test0 {
            color: red
        }

        .longlonglong {
            color: red;
            border: none
        }

        .test1 {
            border: none
        }
```

* 変換後:
```css
        .test0 {
            color: red
        }

        .longlonglong {
            color: red;
            border: none
        }

        .test1 {
            border: none
        }
```

### 1.2.6. ブロックの部分的な分割

隣接する2つのブロックに重複するプロパティがあれば、縮小化が行われます:

* 新しいブロックには2つのブロックの重複プロパティが含まれています。

文字数の節約が期待できるのであれは縮小化が実行されます。

* 変換前:
```css
        .test0 {
            color: red;
            border: none;
            margin: 0
        }

        .test1 {
            color: green;
            border: none;
            margin: 0
        }
```

* 変換後:
```css
        .test0 {
            color: red
        }

        .test0, .test1 {
            border: none;
            margin: 0
        }

        .test1 {
            color: green
        }
```

文字数が増えるので縮小化が実行されません。

* 変換前:
```css
        .test0 {
            color: red;
            border: none;
            margin: 0
        }

        .longlonglong {
            color: green;
            border: none;
            margin: 0
        }
```

* 変換後:
```css
        .test0 {
            color: red;
            border: none;
            margin: 0
        }

        .longlonglong {
            color: green;
            border: none;
            margin: 0
        }
```

### 1.2.7. 空のルールセット・ルールの削除

空のルールセットとルールは削除されます。

* 変換前:
```css
        .test {
            color: red
        }

        .empty {}

        @font-face {}

        @media print {
            .empty {}
        }

        .test {
            border: none
        }
```

* 変換後:
```css
        .test{color:red;border:none}
```

### 1.2.8. margin と padding プロパティの縮小化

\[[CSS 2.1 / 8.3 Margin properties](http://www.w3.org/TR/CSS21/box.html#margin-properties)\] と \[[CSS 2.1 / 8.4 Padding properties](http://www.w3.org/TR/CSS21/box.html#padding-properties)\]の仕様に従って、`margin`と`padding`プロパティの縮小化がされます。

* 変換前:
```css
        .test0 {
            margin-top: 1em;
            margin-right: 2em;
            margin-bottom: 3em;
            margin-left: 4em;
        }

        .test1 {
            margin: 1 2 3 2
        }

        .test2 {
            margin: 1 2 1 2
        }

        .test3 {
            margin: 1 1 1 1
        }

        .test4 {
            margin: 1 1 1
        }

        .test5 {
            margin: 1 1
        }
```

* 変換後:
```css
        .test0 {
            margin: 1em 2em 3em 4em
        }

        .test1 {
            margin: 1 2 3
        }

        .test2 {
            margin: 1 2
        }

        .test3, .test4, .test5 {
            margin: 1
        }
```

# 2. リコメンド

Some stylesheets compress better than the others. Sometimes, one character difference can turn a well-compressible stylesheet to a very inconvenient one.

You can help the minimizer by following these recommendations.

## 2.1. Length of selectors

短いセレクタ名だと再グループが容易です。

## 2.2. プロパティの並び順

スタイルシート全体で同じプロパティ順を守る、つまりガードしなくてもよくなります。手動による介入が減ることで、それは縮小化の効率を高めることになります。

## 2.3. 同様なブロックの配置

似たようなルールセットのブロックは互いに近くに配置すると良いです。

悪い例:

* 変換前:
```css
        .test0 {
            color: red
        }

        .test1 {
            color: green
        }

        .test2 {
            color: red
        }
```

* 変換後 (53 characters):
```css
        .test0{color:red}.test1{color:green}.test2{color:red}
```

良い例:

* 変換前:
```css
        .test1 {
            color: green
        }

        .test0 {
            color: red
        }

        .test2 {
            color: red
        }
```

* 変換後 (43 characters):
```css
        .test1{color:green}.test0,.test2{color:red}
```

## 2.4. !important の使用

言うまでもなく `!important` 宣言は縮小化に悪影響を与えます.

悪い例:

* 変換前:
```css
        .test {
            margin-left: 2px !important;
            margin: 1px;
        }
```

* 変換後 (43 characters):
```css
        .test{margin-left:2px!important;margin:1px}
```

良い例:

* 変換前:
```css
        .test {
            margin-left: 2px;
            margin: 1px;
        }
```

* 変換後 (17 characters):
```css
        .test{margin:1px}
```
