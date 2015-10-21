# 브라우저에서

브라우저에서 [http://css.github.com/csso/csso.html](http://css.github.com/csso/csso.html)를 연다.

**CSSO는 브라우저 상의 동작을 보장하지 않습니다. 커맨드라인상에서 실행하거나 npm 모듈을 통해서 실행하는 것을 추천합니다.**

# 커맨드라인에서

git을 통해서 설치한 경우 `bin/csso`을 실행합니다. nodejs 0.4.x&nbsp;— [http://nodejs.org](http://nodejs.org)가 설치되어 있어야 합니다.

npm을 통해서 설치한 경우 `csso`을 실행합니다.

사용방법:

    csso
        도움말을 표시합니다.
    csso <filename>
        <filename>의 파일명을 가진 CSS파일을 최소화하고 표준출력(stdout)으로 결과를 출력합니다.
    csso <in_filename> <out_filename>
    csso -i <in_filename> -o <out_filename>
    csso --input <in_filename> --output <out_filename>
        <in_filename>의 파일명을 가진 CSS파일을 최소화하여 <out_filename>의 파일로 저장합니다.
    csso -off
    csso --restructure-off
        구조의 최적화를 실행하지 않습니다.
    csso -h
    csso --help
        도움말을 표시합니다.
    csso -v
    csso --version
        버전을 표시합니다.

예시:

    $ echo ".test { color: red; color: green }" > test.css
    $ csso test.css
    .test{color:green}

# npm 모듈로

예시 (`test.js`):
```js
    var csso = require('csso'),
        css = '.test, .test { color: rgb(255, 255, 255) }';

    console.log(csso.justDoIt(css));
```
결과 (`> node test.js`):
```css
    .test{color:#fff}
```
구조의 최적화를 실행시키지 않으려면 `csso.justDoIt(css, true)`을 사용합니다.
