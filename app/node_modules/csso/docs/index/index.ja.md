CSSO (CSSオプティマイザー)は、他とは違ったCSS縮小化ツールです。 一般的な縮小化テクニックに加えてCSSファイルの構造的な最適化も行うので、 他の縮小化ツールと比べてより軽量にすることが可能です。

# 縮小化（要約版）

安全な変換:

* ホワイトスペースの
* 末尾の `;` 削除
* コメントの削除
* 不正な `@charset` と `@import` 宣言の削除
* color プロパティの縮小化
* `0` の縮小化
* 複数行文字列の縮小化
* `font-weight` プロパティの縮小化

構造的な最適化:

* 同一セレクタブロックのマージ
* ブロック内の同一プロパティのマージ
* 上書きされたプロパティの削除
* 上書きされたショートハンドプロパティの削除
* 繰り返されているセレクタの削除
* ブロックの部分的なマージ
* ブロックの部分的な分割
* 空のルールセット・ルールの削除
* `margin` と `padding` プロパティの縮小化

縮小化テクニックの詳細は [detailed description](../description/description.ja.md) で解説されています。

# 著者

* 発案&nbsp;— Vitaly Harisov (<vitaly@harisov.name>)
* 実装&nbsp;— Sergey Kryzhanovsky (<skryzhanovsky@ya.ru>)
* 英語翻訳&nbsp;— Leonid Khachaturov (<leonidkhachaturov@gmail.com>)
* 日本語翻訳&nbsp;— Koji Ishimoto (<ijok.ijok@gmail.com>)
* 韓国語翻訳&nbsp;— Wankyu Kim (<wankyu19@gmail.com>)

# フィードバック

問題の報告は [Github](https://github.com/css/csso/issues) まで。

フィードバック、提案、その他は <skryzhanovsky@ya.ru> まで。

# ライセンス

* CSSOは [MIT](https://github.com/css/csso/blob/master/MIT-LICENSE.txt) ライセンスに基づいています。
