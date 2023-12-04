#include <stdlib.h>
#include <sys/types.h>

#define BUFSIZE 4096

const char *name = "com.horsegraph.connection";
const char *author = "Petri & Thorsten";
const int count;
int count2;
const int age = 28;
static uint sweet_sweet_numbers[5] = {23, 420, 69, 42, 7};

const int *ptr1 = &count;
const int **ptrptr1 = &ptr1;
const int ***ptrptrptr1 = &ptrptr1;

enum { BLACK, RED };

enum animal {
  ANIMAL_TOUCAN,
  ANIMAL_TIGER = 1,
  ANIMAL_TIGGER = ANIMAL_TIGER,
  ANIMAL_HORSE,
  ANIMAL_GIRAFFE,
  ANIMAL_GOPHER = 99,
  ANIMAL_ORANGUTAN
};

typedef enum instrument {
  INSTRUMENT_GUITAR,
  INSTRUMENT_KEYTAR,
  INSTRUMENT_SITAR
} Instrument;

union {
  char *hobby;
  int age; // <-- TODO: this will be tagged as a Constant because it's the same
           // symbol as the `const int age` above.
} person;

union object {
  char *name;
  int value;
  int age;
} obj1, obj2, *obj3;

const union object2 {
  char name;
} obj4;

struct connection {
  int complete;
  int fd;
  int bufsize;
  char *buffer;
  char *url;
};

typedef struct connection Connection;

typedef struct computer {
  int cores;
} Computer;

typedef struct {
  int number;
} NoName;

struct outer {
  struct inner {
    int x;
    int y;
  } b;
};

// Prototype
struct connection *new_connection(int fd);
// Implementation
struct connection *new_connection(int fd) {
  struct connection *c = malloc(sizeof(struct connection));
  if (c == NULL) {
    return NULL;
  }

  c->url = NULL;
  c->complete = 0;
  c->fd = fd;

  c->buffer = calloc(BUFSIZE, sizeof(char));
  if (c->buffer == NULL) {
    free(c);
    return NULL;
  }
  c->bufsize = BUFSIZE;

  return c;
}

static void free_connection(struct connection *c) {
  if (c->buffer) {
    free(c->buffer);
  }

  if (c->url) {
    free(c->url);
  }

  free(c);
}

// Prototype
int returns_int();
// Implementation
int returns_int() { return 12; }

int main() { return 0; }

int k_and_r(s, f)
char *s;
float f;
{ return 5; }
