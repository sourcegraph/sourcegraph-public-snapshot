CSSO (CSS Optimizer)는 다른 CSS압축툴과는 다릅니다. 일반적인 압축 기법에 더불어 CSS파일의 구조를 최적화하여 다른 툴에 비해 더 작은 파일로 만들 수 있습니다.

# 1. 최소화

최소화는 CSS파일을 손실없이 더 작은 용량으로 변환하는 과정입니다. 이를 위해 기본적으로 다음과 같이 접근합니다.

* 불필요한 요소(예: 맨 마지막에 위치한 세미콜론)를 삭제하거나 속성값을 더 간단하게 표기(예: `0px`을 `0`으로)하는 것과 같은 기본적인 변환
* 덮어쓰기된 속성을 삭제하거나 블럭을 합치는 구조의 최적화

## 1.1. 기본적인 변환

### 1.1.1. 공백을 삭제한다.

특정 공백(` `, `\n`, `\r`, `\t`, `\f`)은 불필요하며 렌더링에 영향을 주지 않습니다.

* 변환전:
```css
        .test
        {
            margin-top: 1em;

            margin-left  : 2em;
        }
```

* 변환후:
```css
        .test{margin-top:1em;margin-left:2em}
```

이후의 예시는 가독성을 위해서 공백을 남겨놓습니다.

### 1.1.2. 맨 마지막에 위치한 ';'을 삭제한다.

블럭의 마지막 세미콜론은 필요하지 않으며 렌더링에 영향을 주지 않습니다.

* 변환전:
```css
        .test {
            margin-top: 1em;;
        }
```

* 변환후:
```css
        .test {
            margin-top: 1em
        }
```

### 1.1.3. 주석을 삭제한다.

주석은 렌더링에 영향을 주지 않습니다.: \[[CSS 2.1 / 4.1.9 Comments](http://www.w3.org/TR/CSS21/syndata.html#comments)\].

* 변환전:
```css
        /* comment */

        .test /* comment */ {
            /* comment */ margin-top: /* comment */ 1em;
        }
```

* 변환후:
```css
        .test {
            margin-top: 1em
        }
```

만약 주석을 남겨두고 싶은 경우 `!`로 시작하는 처음의 주석 하나만 남겨둘 수 있습니다.

* 변환전:
```css
        /*! MIT license */
        /*! will be removed */

        .test {
            color: red
        }
```

* 변환후:
```css
        /*! MIT license */
        .test {
            color: red
        }
```

### 1.1.4.  유효하지 않은 @charset과 @import선언을 삭제한다.

표준에 따르면 `@charset`은 스타일시트가 시작하는 부분에 위치해야 합니다.: \[[CSS 2.1 / 4.4 CSS style sheet representation](http://www.w3.org/TR/CSS21/syndata.html#charset)\].

CSSO는 이 규칙을 조금 유연하게 적용하여 스타일시트의 시작 부분에 있는 공백과 주석 바로 다음의 `@charset`을 유지시킵니다.

\[[CSS 2.1 / 6.3 The @import rule](http://www.w3.org/TR/CSS21/cascade.html#at-import)\]의 규칙에 따라서 올바르지 않은 곳에 위치한 `@import`를 삭제합니다.

* 변환전:
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

* 변환후:
```css
        @charset 'UTF-8';
        @import "test0.css";
        @import "test1.css";
        h1 {
            color: red
        }
```

### 1.1.5. 색상값을 최소화한다.

\[[CSS 2.1 / 4.3.6 Colors](http://www.w3.org/TR/CSS21/syndata.html#color-units)\]의 규칙에 따라 일부 색상값을 최소화합니다.

* 변환전:
```css
        .test {
            color: yellow;
            border-color: #c0c0c0;
            background: #ffffff;
            border-top-color: #f00;
            outline-color: rgb(0, 0, 0);
        }
```

* 변환후:
```css
        .test {
            color: #ff0;
            border-color: silver;
            background: #fff;
            border-top-color: red;
            outline-color: #000
        }
```

### 1.1.6. 0을 최소화한다.

경우에 따라 숫자로 된 값은 `0`으로 간략하게 하거나 생략할 수 있습니다. 

`0%`의 값은 다음과 같은 경우를 고려하여 간략화하지 않습니다.: `rgb(100%, 100%, 0)`

* 변환전:
```css
        .test {
            fakeprop: .0 0. 0.0 000 00.00 0px 0.1 0.1em 0.000em 00% 00.00% 010.00
        }
```

* 변환후:
```css
        .test {
            fakeprop: 0 0 0 0 0 0 .1 .1em 0 0% 0% 10
        }
```

### 1.1.7. 여러 줄의 문자열을 최소화한다.

\[[CSS 2.1 / 4.3.7 Strings](http://www.w3.org/TR/CSS21/syndata.html#strings)\]의 규칙에 따라 여러 줄의 문자열을 최소화합니다.

* 변환전:
```css
        .test[title="abc\
        def"] {
            background: url("foo/\
        bar")
        }
```

* 변환후:
```css
        .test[title="abcdef"] {
            background: url("foo/bar")
        }
```

### 1.1.8. font-weight속성을 최소화한다.

\[[CSS 2.1 / 15.6 Font boldness: the 'font-weight' property](http://www.w3.org/TR/CSS21/fonts.html#font-boldness)\]의 규칙에 따라 `font-weight`속성의 `bold`와 `normal`값을 최소화합니다.

* 변환전:
```css
        .test0 {
            font-weight: bold
        }

        .test1 {
            font-weight: normal
        }
```

* 변환후:
```css
        .test0 {
            font-weight: 700
        }

        .test1 {
            font-weight: 400
        }
```

## 1.2. 구조의 최적화

### 1.2.1. 동일한 선택자의 블럭을 병합한다.

동일한 선택자에 지정된 블럭이 인접해있을 경우 병합합니다.

* 변환전:
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

* 변환후:
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

### 1.2.2. 동일한 속성을 가진 블럭을 병합한다.

동일한 속성을 가진 블럭이 인접해있을 경우 병합합니다.

* 변환전:
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

* 변환후:
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

### 1.2.3. 덮어쓰기된 속성을 삭제한다.

브라우저가 무시하는 속성을 다음 규칙에 따라 삭제합니다.:

* `!important`로 선언된 속성이 없다면 CSS 룰에서 맨 나중에 선언된 속성이 적용됩니다.
* `!important`로 선언된 속성 중 맨 나중에 선언된 속성이 적용됩니다.

* 변환전:
```css
        .test {
            color: red;
            margin: 0;
            line-height: 3cm;
            color: green;
        }
```

* 변환후:
```css
        .test {
            margin: 0;
            line-height: 3cm;
            color: green
        }
```

#### 1.2.3.1. 덮어쓰기된 단축속성을 삭제한다.

`border`, `margin`, `padding`, `font`, `list-style`속성의 경우 다음 규칙에 따라 삭제합니다.: 나중에 선언된 속성이 대표 속성(예: `border`)이라면 그 이전에 선언되어 덮어쓰기된 모든 속성을 삭제합니다.(예: `border-top-width` 이나 `border-style`)

* 변환전:
```css
        .test {
            border-top-color: red;
            border-color: green
        }
```

* 변환후:
```css
        .test {
            border-color:green
        }
```

### 1.2.4. 반복된 선택자를 삭제한다.

반복된 선택자를 삭제합니다.

* 변환전:
```css
        .test, .test {
            color: red
        }
```

* 변환후:
```css
        .test {
            color: red
        }
```

### 1.2.5. 블럭을 부분적으로 병합한다.

인접한 2개의 블럭중 한쪽이 다른 한쪽에 완전히 포함되는 속성을 가진 경우 다음과 같은 최적화가 가능합니다.:

* 겹치는 속성을 원래 블럭에서 삭제합니다.
* 원래 블럭에 남아있는 속성을 다른 블럭으로 복사합니다.

겹치는 속성의 문자 수가 복사할 속성의 문자 수보다 작을 경우에만 최소화를 실행합니다.

* 변환전:
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

* 변환후:
```css
        .test0, .test1 {
            color: red
        }

        .test1, .test2 {
            border: none
        }
```

겹치는 속성의 문자 수가 복사할 속성의 문자 수보다 클 경우에는 최소화를 실행하지 않습니다.

* 변환전:
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

* 변환후:
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

### 1.2.6. 블럭을 부분적으로 분할한다.

인접한 2개의 블럭이 공통된 속성을 가지고 있을 경우 다음과 같은 최적화가 가능합니다.:

* 공통된 속성을 확정합니다.
* 인접한 2블럭 사이에 공통된 속성을 가진 새로운 블럭을 생성합니다.

문자 수가 절약될 경우 최소화를 실행합니다.

* 변환전:
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

* 변환후:
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

문자 수가 절약되지 않을 경우 최소화를 실행하지 않습니다.

* 변환전:
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

* 변환후:
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

### 1.2.7. 비어있는 룰셋과 @룰을 삭제한다.

비어있는 룰셋과 @룰을 삭제합니다.

* 변환전:
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

* 변환후:
```css
        .test{color:red;border:none}
```

### 1.2.8. margin과 padding속성을 최소화한다.

\[[CSS 2.1 / 8.3 Margin properties](http://www.w3.org/TR/CSS21/box.html#margin-properties)\]과 \[[CSS 2.1 / 8.4 Padding properties](http://www.w3.org/TR/CSS21/box.html#padding-properties)\]의 규칙에 따라 `margin`과 `padding`속성을 최소화합니다.

* 변환전:
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

* 변환후:
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

# 2. 권장사항

스타일시트에 따라서 압축이 더 잘되는 경우도 있습니다. 가끔은 한 글자 차이로 잘 압축되던 스타일시트가 압축효율이 낮은 파일로 변하기도 합니다.

다음의 권장사항을 따르면 최소화를 더 효율적으로 할 수 있습니다.

## 2.1. 선택자의 길이

짧은 선택자가 재구성하기 더 쉽습니다.

## 2.2. 속성의 나열 순서

스타일시트 전체의 속성 나열 순서를 지킵니다. 임의로 조정한 내용이 적을 수록 효율적으로 최소화하기에 좋습니다.

## 2.3. 유사한 블럭의 배치

속성들의 집합이 유사한 블럭들을 서로 가까이 위치하게 합니다.

나쁜 예:

* 변환전:
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

* 변환후 (53 characters):
```css
        .test0{color:red}.test1{color:green}.test2{color:red}
```

좋은 예:

* 변환전:
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

* 변환후 (43 characters):
```css
        .test1{color:green}.test0,.test2{color:red}
```

## 2.4. !important의 사용

당연한 이야기지만 `!important`를 사용하는 것은 최소화에 악영향을 줍니다.

나쁜 예:

* 변환전:
```css
        .test {
            margin-left: 2px !important;
            margin: 1px;
        }
```

* 변환후 (43 characters):
```css
        .test{margin-left:2px!important;margin:1px}
```

좋은 예:

* 변환전:
```css
        .test {
            margin-left: 2px;
            margin: 1px;
        }
```

* 변환후 (17 characters):
```css
        .test{margin:1px}
```
