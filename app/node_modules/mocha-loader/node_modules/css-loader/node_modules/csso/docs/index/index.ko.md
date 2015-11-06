CSSO (CSS Optimizer)는 다른 CSS압축툴과는 다릅니다. 일반적인 압축 기법에 더불어 CSS파일의 구조를 최적화하여 다른 툴에 비해 더 작은 파일로 만들 수 있습니다.

# 최소화（요약）

안전한 변환:

* 공백을 삭제한다.
* 맨 마지막에 위치한 ';'을 삭제한다.
* 주석을 삭제한다.
* 유효하지 않은 @charset과 @import선언을 삭제한다.
* 색상값을 최소화한다.
* `0`을 최소화한다.
* 여러 줄의 문자열을 최소화한다.
* `font-weight`속성을 최소화한다.

구조의 최적화:

* 동일한 선택자의 블럭을 병합한다.
* 동일한 속성을 가진 블럭을 병합한다.
* 덮어쓰기된 속성을 삭제한다.
* 덮어쓰기된 단축속성을 삭제한다.
* 반복된 선택자를 삭제한다.
* 블럭을 부분적으로 병합한다.
* 블럭을 부분적으로 분할한다.
* 비어있는 룰셋과 @룰을 삭제한다.
* `margin`과 `padding`속성을 최소화한다.

최소화 기법의 상세한 내용은 [detailed description](../description/description.ko.md)에 있습니다.

# 저자

* 초안&nbsp;— Vitaly Harisov (<vitaly@harisov.name>)
* 구현&nbsp;— Sergey Kryzhanovsky (<skryzhanovsky@ya.ru>)
* 영어 번역&nbsp;— Leonid Khachaturov (<leonidkhachaturov@gmail.com>)
* 일본어 번역&nbsp;— Koji Ishimoto (<ijok.ijok@gmail.com>)
* 한국어 번역&nbsp;— Wankyu Kim (<wankyu19@gmail.com>)

# 피드백

문제가 있을 경우 [Github](https://github.com/css/csso/issues)에 올려주시기 바랍니다.

피드백, 제안 등은 <skryzhanovsky@ya.ru>로 전달바랍니다.

# 라이센스

* CSSO는 [MIT](https://github.com/css/csso/blob/master/MIT-LICENSE.txt)라이센스를 따릅니다.
