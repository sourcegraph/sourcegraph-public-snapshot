CSSO (CSS Optimizer) является минимизатором CSS, выполняющим как минимизацию без изменения структуры, так и структурную минимизацию с целью получить как можно меньший текст.

## Минимизация

Цель минимизации заключается в трансформации исходного CSS в CSS меньшего размера. Наиболее распространёнными стратегиями в достижении этой цели являются:

* минимизация без изменения структуры&nbsp;— удаление необязательных элементов (например, `;` у последнего свойства в блоке), сведение значений к меньшим по размеру (например, `0px` к `0`) и т.п.;
* минимизация с изменением структуры&nbsp;— удаление перекрываемых свойств, полное или частичное слияние блоков.

### Минимизация без изменения структуры

#### Удаление whitespace

В ряде случаев символы ряда whitespace (` `, `\n`, `\r`, `\t`, `\f`) являются необязательными и не влияют на результат применения таблицы стилей.

* Было:
```css
        .test
        {
            margin-top: 1em;

            margin-left  : 2em;
        }
```

* Стало:
```css
        .test{margin-top:1em;margin-left:2em}
```

Для большего удобства чтения текст остальных примеров приводится с пробелами (переводом строки и т.п.).

#### Удаление концевых ';'

Символ `;`, завершающий перечисление свойств в блоке, является необязательным и не влияет на результат применения таблицы стилей.

* Было:
```css
        .test {
            margin-top: 1em;;
        }
```

* Стало:
```css
        .test {
            margin-top: 1em
        }
```

#### Удаление комментариев

Комментарии не влияют на результат применения таблицы стилей: \[[CSS 2.1 / 4.1.9 Comments](http://www.w3.org/TR/CSS21/syndata.html#comments)\].

* Было:
```css
        /* comment */

        .test /* comment */ {
            /* comment */ margin-top: /* comment */ 1em;
        }
```

* Стало:
```css
        .test {
            margin-top: 1em
        }
```

Если вам требуется сохранить комментарий, CSSO позволяет это сделать только с одним первым комментарием, если его текст начинается с `!`.

* Было:
```css
        /*! MIT license */
        /*! will be removed */

        .test {
            color: red
        }
```

* Стало:
```css
        /*! MIT license */
        .test {
            color: red
        }
```

#### Удаление неправильных @charset и @import

Единственно верным расположением `@charset` является начало текста: \[[CSS 2.1 / 4.4 CSS style sheet representation](http://www.w3.org/TR/CSS21/syndata.html#charset)\].

Однако CSSO позволяет обходиться с этим правилом достаточно вольно, т.к. оставляет первый после whitespace и комментариев `@charset`.

Правило `@import` на неправильном месте удаляется согласно \[[CSS 2.1 / 6.3 The @import rule](http://www.w3.org/TR/CSS21/cascade.html#at-import)\].

* Было:
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

* Стало:
```css
        @charset 'UTF-8';
        @import "test0.css";
        @import "test1.css";
        h1 {
            color: red
        }
```

#### Минимизация цвета

Некоторые значения цвета минимизируются согласно \[[CSS 2.1 / 4.3.6 Colors](http://www.w3.org/TR/CSS21/syndata.html#color-units)\].

* Было:
```css
        .test {
            color: yellow;
            border-color: #c0c0c0;
            background: #ffffff;
            border-top-color: #f00;
            outline-color: rgb(0, 0, 0);
        }
```

* Стало:
```css
        .test {
            color: #ff0;
            border-color: silver;
            background: #fff;
            border-top-color: red;
            outline-color: #000
        }
```

#### Минимизация 0

В ряде случаев числовое значение можно сократить до `0` или же отбросить `0`.

Значения `0%` не сокращаются до `0`, чтобы избежать ошибок вида `rgb(100%, 100%, 0)`.

* Было:
```css
        .test {
            fakeprop: .0 0. 0.0 000 00.00 0px 0.1 0.1em 0.000em 00% 00.00% 010.00
        }
```

* Стало:
```css
        .test {
            fakeprop: 0 0 0 0 0 0 .1 .1em 0 0% 0% 10
        }
```

#### Слияние многострочных строк в однострочные

Многострочные строки минимизируются согласно \[[CSS 2.1 / 4.3.7 Strings](http://www.w3.org/TR/CSS21/syndata.html#strings)\].

* Было:
```css
        .test[title="abc\
        def"] {
            background: url("foo/\
        bar")
        }
```

* Стало:
```css
        .test[title="abcdef"] {
            background: url("foo/bar")
        }
```

#### Минимизация font-weight

Значения `bold` и `normal` свойства `font-weight` минимизируются согласно \[[CSS 2.1 / 15.6 Font boldness: the 'font-weight' property](http://www.w3.org/TR/CSS21/fonts.html#font-boldness)\].

* Было:
```css
        .test0 {
            font-weight: bold
        }

        .test1 {
            font-weight: normal
        }
```

* Стало:
```css
        .test0 {
            font-weight: 700
        }

        .test1 {
            font-weight: 400
        }
```

### Минимизация с изменением структуры

#### Слияние блоков с одинаковыми селекторами

В один блок сливаются соседние блоки с одинаковым набором селекторов.

* Было:
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

* Стало:
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

#### Слияние блоков с одинаковыми свойствами

В один блок сливаются соседние блоки с одинаковым набором свойств.

* Было:
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

* Стало:
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

#### Удаление перекрываемых свойств

Минимизация удалением перекрываемых свойств основана на том, что внутри блока применяется:

* последнее по порядку свойство, если все свойства не `!important`;
* последнее по порядку свойство `!important`.

Это позволяет избавиться от всех игнорируемых браузером свойств.

* Было:
```css
        .test {
            color: red;
            margin: 0;
            line-height: 3cm;
            color: green;
        }
```

* Стало:
```css
        .test {
            margin: 0;
            line-height: 3cm;
            color: green
        }
```

##### Удаление перекрываемых shorthand свойств

Для свойств `border`, `margin`, `padding`, `font` и `list-style` используется следующий алгоритм удаления: если последним по порядку свойством является более 'широкое' свойство (например, `border`), то все предыдущие перекрываемые им свойства удаляются (например, `border-top-width` или `border-style`).

* Было:
```css
        .test {
            border-top-color: red;
            border-color: green
        }
```

* Стало:
```css
        .test {
            border-color:green
        }
```

#### Удаление повторяющихся селекторов

Повторяющиеся селекторы излишни и потому могут быть удалены.

* Было:
```css
        .test, .test {
            color: red
        }
```

* Стало:
```css
        .test {
            color: red
        }
```

#### Частичное слияние блоков

Если рядом расположены блоки, один из которых набором свойств полностью входит в другой, возможна следующая минимизация:

* в исходном (наибольшем) блоке удаляется пересекающийся набор свойств;
* селекторы исходного блока копируются в принимающий блок.

Если в символах размер копируемых селекторов меньше размера пересекающегося набора свойств, минимизация происходит.

* Было:
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

* Стало:
```css
        .test0, .test1 {
            color: red
        }

        .test1, .test2 {
            border: none
        }
```

Если в символах размер копируемых селекторов больше размера пересекающегося набора свойств, минимизация не происходит.

* Было:
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

* Стало:
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

#### Частичное разделение блоков

Если рядом расположены блоки, частично пересекающиеся набором свойств, возможна следующая минимизация:

* из обоих блоков выделяется пересекающийся набор свойств;
* между блоками создаётся новый блок с выделенным набором свойств и с селекторами обоих блоков.

Если в символах размер копируемых селекторов меньше размера пересекающегося набора свойств, минимизация происходит.

* Было:
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

* Стало:
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

Если в символах размер копируемых селекторов больше размера пересекающегося набора свойств, минимизация не происходит.

* Было:
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

* Стало:
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

#### Удаление пустых ruleset и at-rule

Пустые ruleset и at-rule удаляются.

* Было:
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

* Стало:
```css
        .test{color:red;border:none}
```

#### Минимизация margin и padding

Свойства `margin` и `padding` минимизируются согласно \[[CSS 2.1 / 8.3 Margin properties](http://www.w3.org/TR/CSS21/box.html#margin-properties)\] и \[[CSS 2.1 / 8.4 Padding properties](http://www.w3.org/TR/CSS21/box.html#padding-properties)\].

* Было:
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

* Стало:
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

Минимизация не происходит в случаях, когда один набор `selector X / shorthands` прерывается другим набором `selector Y / shorthands`.

* Было:
```css
        .test1 {
            margin-top: 0
        }

        .test2 {
            margin-top: 100px
        }

        .test1 {
            margin-left: 0
        }

        .test1 {
            margin-bottom: 0
        }

        .test1 {
            margin-right: 0
        }
```

* Стало:
```css
        .test1 {
            margin-top: 0
        }

        .test2 {
            margin-top: 100px
        }

        .test1 {
            margin-left: 0;
            margin-bottom: 0;
            margin-right: 0
        }
```

* Могло быть (неправильно):
```css
        .test2 {
            margin-top: 100px
        }

        .test1 {
            margin: 0
        }
```

К сожалению, результат рендеринга последнего варианта отличается от рендеринга исходного стиля, потому такая минимизация недопустима.

#### Специальная минимизация псевдоклассов

Если в группе селекторов UA обнаружит неподдерживаемый селектор, он, согласно \[[CSS 3 / Selectors / 5. Groups of selectors](http://www.w3.org/TR/selectors/#grouping)\], посчитает неподдерживаемой всю группу и не применит стили к перечисленным в ней селекторам. Этим нередко пользуются для разграничения стилей по браузерам. Вот [пример](http://pornel.net/firefoxhack) метода:
```css
        #hackme, x:-moz-any-link { Firefox 2.0 here }
        #hackme, x:-moz-any-link, x:default { Firefox 3.0 and newer }
```

Чтобы сохранить такие блоки, но в то же время минимизировать то, что поддаётся оптимизации, CSSO применяет нижеперечисленные правила. Предполагается, что вместе эти правила составляют компромисс, удовлетворяющий большинство пользователей.

##### Сохранение группы

В общем случае (исключения описаны ниже) минимизация удалением перекрываемого селектора не происходит, если группа селекторов включает псевдокласс или псевдоэлемент.

* Было:
```css
        a {
            property: value0
        }

        a, x:-vendor-class {
            property: value1
        }
```

* Стало (структура не изменилась):
```css
        a {
            property: value0
        }

        a, x:-vendor-class {
            property: value1
        }
```

Если же группы селекторов образуют одинаковую "сигнатуру псевдоклассов" (исключается ситуация, в которой браузер поддерживает одну группу, но не поддерживает другую), минимизация происходит.

* Было:
```css
        a, x:-vendor-class {
            property: value0
        }

        a, b, x:-vendor-class {
            property: value1
        }
```

* Стало:
```css
        a, b, x:-vendor-class {
            property: value1
        }
```

##### Минимизация общеподдерживаемых псевдоклассов и псевдоэлементов

Существуют псевдоклассы и псевдоэлементы, поддерживаемые большинством браузеров: `:link`, `:visited`, `:hover`, `:active`, `:first-letter`, `:first-line`. Для них минимизация происходит в общем порядке без сохранения группы.

* Было:
```css
        a, x:active {
            color: red
        }

        a {
            color: green
        }
```

* Стало:
```css
        x:active {
            color: red
        }

        a {
            color: green
        }
```

##### Минимизация :before и :after

Псевдоэлементы `:before` и `:after` обычно поддерживаются браузерами парно, потому объединение блоков с их участием безопасно.

* Было:
```css
        a, x:before {
            color: red
        }

        a, x:after {
            color: red
        }
```

* Стало:
```css
        a, x:before, x:after {
            color:red
        }
```

Тем не менее, блоки, в которых участвует только один из этих псевдоэлементов, не объединяются:

* Было:
```css
        a {
            color: red
        }

        a, x:after {
            color: red
        }
```

* Стало:
```css
        a {
            color: red
        }

        a, x:after {
            color: red
        }
```

В примере выше можно заметить, что удаление селектора `a` из второго блока не повлияло бы на итоговый рендеринг, но в общем случае это опасная минимизация, потому не применяется.

## Рекомендации

С точки зрения минимизации таблицы стилей можно разделить на две группы: удобные и неудобные. Разница даже в один символ может превратить вполне сокращаемый исходный текст в минимально обрабатываемый.

Если вы хотите помочь минимизатору хорошо выполнить работу, следуйте рекомендациям.

### Длина селекторов

Чем короче селектор (whitespace не учитываются), тем больше вероятность удачного группирования.

### Порядок свойств

Придерживайтесь во всём CSS одного порядка, в котором перечисляются свойства, так вам не потребуется защита от смены порядка. Соответственно, меньше вероятность допустить ошибку и помешать минимизатору излишним управлением.

### Расположение схожих блоков

Располагайте блоки со схожим набором свойств как можно ближе друг к другу.

Плохо:

* Было:
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

* Стало (53 символа):
```css
        .test0{color:red}.test1{color:green}.test2{color:red}
```

Хорошо:

* Было:
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

* Стало (43 символа):
```css
        .test1{color:green}.test0,.test2{color:red}
```

### Использование !important

Очевидно, `!important` оказывает серьёзное влияние на минимизацию, особенно заметно это может отразиться на минимизации `margin` и `padding`, потому им лучше не злоупотреблять.

Плохо:

* Было:
```css
        .test {
            margin-left: 2px !important;
            margin: 1px;
        }
```

* Стало (43 символа):
```css
        .test{margin-left:2px!important;margin:1px}
```

Хорошо:

* Было:
```css
        .test {
            margin-left: 2px;
            margin: 1px;
        }
```

* Стало (17 символов):
```css
        .test{margin:1px}
```
