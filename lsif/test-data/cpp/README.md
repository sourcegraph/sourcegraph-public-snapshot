# LSIF data for C++ tests (via DXR plugin)

Getting the DXR plugin to work with macOS is an ongoing project. For now, we just version control some output. The reference source is given below.

### main.cpp

```
#include "five.h"

#define TABLE_SIZE 100

int x = TABLE_SIZE;

int four(int y) {
  return y;
}

int main() {
  five(x);
  four(x);
  five(six);
}
```

### five.h

```cpp
#include "five.h"

int five(int x) {
  return 5;
}
```

### five.cpp

```cpp
int six;
int five(int x);
```
