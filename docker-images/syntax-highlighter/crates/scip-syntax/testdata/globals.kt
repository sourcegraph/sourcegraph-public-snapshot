package org.example

// Top level constant property
const val PI = 3.14

// Top level property with getter
val version: String
    get() = "1.0.0"

// Top level function
fun printHello() {
    println("Hello, Kotlin!")
}

// Class with properties and methods
class MyKotlinClass {
    var prop: String = "property"

    fun method() {
        println("This is a method")
    }
}

// Data class
data class User(val name: String, val age: Int)

// Enum class
enum class Days {
    MONDAY,
    TUESDAY,
    WEDNESDAY,
    THURSDAY,
    FRIDAY,
    SATURDAY,
    SUNDAY
}

// Object (singleton)
object MyObject {
    val property = "Object property"
}

// Interface
interface MyInterface {
    fun interfaceMethod(): String
}

object SimpleSingleton {
    val answer = 42;
    fun greet(name: String) = "Hello, $name!"
}

// Type alias
typealias UserList = List<User>

// Extension function
fun String.print() {
    println(this)
}

// Sealed class
sealed class Result {
    data class Success(val message: String) : Result()
    data class Error(val error: Exception) : Result()
}

// Inline class
inline class Password(val value: String)

// Companion object
class MyClassWithCompanion {
    companion object ConstantCompanion {
        const val CONSTANT = "Companion constant"
    }
}

fun `Escaped`() {}

// Multi-variable declaration
val (left, right) = directions()

// Anonymous function
fun(x: Int, y: Int): Int = x + y
