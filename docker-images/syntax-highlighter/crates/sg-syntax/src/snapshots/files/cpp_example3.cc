#include <iostream>
#include <map>
#include <vector>

void rangeBasedLoops() {
    std::vector<int> nums = {1, 2, 3, 4, 5};

    // Iterate over vector using range-based for loop
    for (int num : nums) {
        std::cout << num << " ";
    }
    std::cout << std::endl;

    std::map<std::string, int> ages = {{"Alice", 25}, {"Bob", 30}, {"Charlie", 35}};

    // Iterate over map using range-based for loop
    for (const auto& [name, age] : ages) {
        std::cout << name << " is " << age << " years old." << std::endl;
    }
}

void closures() {
    int x = 5;

    // Define closure with capture
    auto add_x = [x](int y) {
        return x + y;
    };

    // Call closure with different arguments
    std::cout << add_x(10) << std::endl; // prints 15
}

// Define user-defined literal for string suffix "_s"
std::string operator"" _s(const char* str, std::size_t len) {
    return std::string(str, len);
}

// Define user-defined literal for number suffix "_m"
constexpr int operator"" _m(unsigned long long x) {
    return x * 1000;
}

void userDefinedLiterals() {
    // Use custom string literal suffix "_s"
    std::string s = "hello"_s;
    std::cout << s << std::endl; // prints "hello"

    int meters = 5_m;
    std::cout << meters << " meters" << std::endl; // prints "5000 meters"
}

#pragma once

#include <iostream>

#pragma warning(push)
#pragma warning(disable: 4996)

int pragmas() {
    #pragma message("This is a message from the preprocessor!")
    std::cout << "Hello, world!" << std::endl;

    return 0;
}

#pragma warning(pop)

#define MESSAGE "Hello, world! This is a very long message that spans multiple lines. " \
                "It is being continued using the line continuation character."

#include <iostream>

void preprocessorContinuations() {
    std::cout << MESSAGE << std::endl;
}


int andOrOperators() {
    int x = 5;
    int y = 10;
    bool b1 = x > 3 and y < 20; // b1 is true
    bool b2 = x < 3 or y > 20;  // b2 is false
    bool b3 = x > 3 and y > 20; // b3 is false
    bool b4 = x < 3 or y < 20;  // b4 is true
}


int multibyteCharacters() {
    std::string s1 = "Hello, world!";    // single-byte characters
    std::string s2 = "こんにちは世界！";  // multi-byte characters in UTF-8 encoding

    char c1 = 'A';          // single-byte character
    char c2 = u8'あ';       // multi-byte character encoded in UTF-8


    std::cout << s1 << std::endl;
    std::cout << s2 << std::endl;

    return 0;
}

void literals() {
    auto intLit = 1234;
    auto floatLit = 1.23f;
    auto boolLit = true;
    char charLit = 'x';
    auto stringLit = "Hello";
    auto wideStringLit = L"World";
    auto utf8StringLit = u8"World";
    auto utf16StringLit = u"World";
    auto utf32StringLit = U"World";
    auto rawStringLit = R"sequence(Hello
    \n
    \r
    \t
    World)sequence";
    auto nullptrLit = nullptr;
    auto nullLit = NULL;

    // Examples from https://en.cppreference.com/w/cpp/language/string_literal
    const wchar_t* sC = LR"--(STUV)--"; // ok, raw string literal
    const wchar_t* s4 = L"ABC" L"DEF"; // ok, same as
    const wchar_t* s5 = L"ABCDEF";
    const char32_t* s6 = U"GHI" "JKL"; // ok, same as
    const char32_t* s7 = U"GHIJKL";
    const char16_t* s9 = "MN" u"OP" "QR"; // ok, same as
    const char16_t* sA = u"MNOPQR";
}


// Structs, enums, unions
struct Point {
    int x;
    int y;
    Point() : x(0), y(0) { } // Default constructor
    Point(int x, int y) : x(x), y(y) { } // Initialization constructor
};
enum Color {
    Red,
    Green,
    Blue
};
union IntOrFloat {
    int i;
    float f;
};

void structsEnumsUnions() {
    Point p {5, 10};
    Color c = Blue;
    IntOrFloat iof;
    iof.i = 10; // Uses int member
    iof.f = 3.14f; // Now uses float member
}

// Templates
template <typename T>
void swap(T& a, T& b) {
    T temp = a;
    a = b;
    b = temp;
}
template <typename T, unsigned N>
struct Array {
    T data[N];
};

void templates() {
    int x = 5, y = 10;
    swap(x, y); // x is now 10, y is now 5
    double d1 = 1.2, d2 = 3.4;
    swap(d1, d2); // d1 is now 3.4, d2 is now 1.2
    Array<int, 10> ints; // Holds 10 ints
    Array<char, 100> chars; // Holds 100 chars
}

// Classes
class Person {
public:
    Person(std::string name, int age); // Constructor
    ~Person(); // Destructor
    std::string getName(); // Accessor method
    // Virtual method
    virtual void greeting();
    // Overloaded operator
    friend std::ostream& operator<<(std::ostream& os, const Person& p);
private:
    std::string name;
    int age;
};
class Employee : public Person { // Inherits from Person
public:
    Employee(std::string name, int age, std::string company);
    // Overridden virtual method
    void greeting() override;
};

// Attributes
[[my_attribute]]
struct my_attribute {
    int value;
};

namespace std {
    template <>
    struct has_attribute<my_attribute> : true_type { };
}

void attributes() {
    [[my_attribute(5)]] int data;
    if (has_attribute<my_attribute>(myData)) {
        // data has the my_attribute attribute
    }
}
