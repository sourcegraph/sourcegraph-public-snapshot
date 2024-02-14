package foobar

// Comments
/* Multi
   line
   comment */

// Imports
import java.util.*

import java.nio.channels.FileChannel

fun Mat.put(indices: IntArray, data: UShortArray)  = this.put(indices, data.asShortArray())

/***
 *  Example use:
 *
 *  val (b, g, r) = mat.at<UByte>(50, 50).v3c
 *  mat.at<UByte>(50, 50).val = T3(245u, 113u, 34u)
 *
 */
@Suppress("UNCHECKED_CAST")
inline fun <reified T> Mat.at(row: Int, col: Int) : Atable<T> =
    when (T::class) {
        UShort::class -> AtableUShort(this, row, col) as Atable<T>
        else -> throw RuntimeException("Unsupported class type")
    }


/**
 * Implementation of [DataAccess] which handles access and interactions with file and data
 * under scoped storage via the MediaStore API.
 */
@RequiresApi(Build.VERSION_CODES.Q)
internal class MediaStoreData(context: Context, filePath: String, accessFlag: FileAccessFlags) :
	DataAccess(filePath) {

	private data class DataItem(
		val id: Long,
		val mediaType: String
	)

	companion object {

		private val PROJECTION = arrayOf(
			MediaStore.Files.FileColumns._ID
		)

		private const val SELECTION_BY_PATH = "${MediaStore.Files.FileColumns.DISPLAY_NAME} = ? " +
			" AND ${MediaStore.Files.FileColumns.RELATIVE_PATH} = ?"

		private fun getSelectionByPathArguments(path: String): Array<String> {
			return arrayOf(getMediaStoreDisplayName(path), getMediaStoreRelativePath(path))
		}
	}
	override val fileChannel: FileChannel

	init {
		val contentResolver = context.contentResolver
		val dataItems = queryByPath(context, filePath)


		id = dataItem.id
		uri = dataItem.uri
	}
}


// Variables
var a = 1         // Mutable
val b = 2         // Immutable
var c: Int = 3     // Type specified

// Nullable types
var nullable: String? = null

// Functions
fun doSomething() { }
fun sum(x: Int, y: Int): Int { return x + y }

// Lambdas/Anonymous functions
fun exampleLambda(a: Int, func: (Int) -> Int) {
    println(func(a))
}

// String interpolation
var name = "John"
println("Hello $name!")

// Conditional expressions
var max = a > b ? a : b

// Range
for (i in 1..10) { print(i) }   // 1 to 10

// Collections
val list = listOf(1, 2, 3)
val set = setOf("a", "b", "c")
val map = mapOf(1 to "a", 2 to "b")

// Null safety
var length: Int? = null
val l = length ?: -1   // Elvis operator

// Smart casts
fun example(x: Any) {
    if (x is String) {  // Smart cast to String
        print(x.length)
    }
}

// Extension functions
fun String.isUppercase() = this.toUpperCase() == this


// Class
class Person(var name: String) {
    // Constructor
    constructor(name: String, age: Int) : this(name) {
        // ...
    }

    // Methods
    fun printName() { println(name) }
}

// Abstract class
abstract class Animal {
    abstract fun makeSound()
}

// Interface
interface Flyer {
    fun fly()
}

// Object (singleton)
object DataProvider {
    val name = "John"
}

// Inheritance
class Dog(name: String) : Animal() {
    override fun makeSound() { print("Bark!") }
}

// Implementing interface
class Bird : Animal(), Flyer {
    override fun makeSound() { print("Chirp!") }
    override fun fly() { println("Flutter flutter!") }
}

// Annotation
annotation class UseCase(val name: String)

// Apply to function
@UseCase("Check user balance")
fun checkBalance() { /*  ... */ }

// Apply to class
@UseCase("Provide login functionality")
class LoginService { /* ... */ }

// Use reflection to read annotation metadata
checkBalance::class.java.getAnnotation(UseCase::class.java).name
// Prints "Check user balance"

// Annotations can have parameters
annotation class Name(val first: String, val last: String)
@Name(first = "John", last = "Doe")
fun doSomething() { ... }

// Built-in annotations
@JvmStatic  // Java static equivalent
fun doSomething() {}

@Throws(IOException::class)  // Declares exceptions thrown
fun readFile() {}


// Usage
fun main() {
    val person = Person("Jack")
    person.printName()

    val dog = Dog("Max")
    dog.makeSound()

    val bird = Bird()
    bird.makeSound()
    bird.fly()

    println(DataProvider.name)
}


class Person(val name: String) {
    var lastName: String

    init {
        lastName = "Unknown"
    }
    companion object {
        val defaultName = "Default Name"

        fun from(name: String): Person {
            return Person(name)
        }
    }
}

val john = Person.from("John")

object UserStore {
    val users = arrayListOf<Person>()

    init {
        File("users.txt").readLines().forEach {
            users.add(Person(it))
        }
    }
}
