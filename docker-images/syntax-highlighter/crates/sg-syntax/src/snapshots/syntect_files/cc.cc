#include "common/common.h"
#include <csignal>

int main() {
    syntect::Exception::raise("oops");
}
