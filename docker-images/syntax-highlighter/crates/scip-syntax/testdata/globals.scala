// Top level package, symbol: com.example
package com.example

// Import statements (not typically symbol indexed)
import scala.collection.immutable._

// Top level class, symbol: com.example.MyClass
class MyClass {
  def method1(): Unit = ()
}

// Top level abstract class, symbol: com.example.MyAbstractClass
abstract class MyAbstractClass {
  def abstractMethod: Int
}

// Top level case class, symbol: com.example.MyCaseClass
case class MyCaseClass(a: Int, b: String)

// Top level object, symbol: com.example.MyObject
object MyObject {
  def method2(): Unit = ()
}

// Top level case object, symbol: com.example.MyCaseObject
case object MyCaseObject {
  def method3(): Unit = ()
}

// Top level trait, symbol: com.example.MyTrait
trait MyTrait {
  def method4(): String
}

// Another way to declare package, symbol: com.example.inner
package inner {
  // Top level class within package, symbol: com.example.inner.InnerClass
  class InnerClass {
    def innerMethod(): Unit = ()
  }

  // Nested object within package, symbol: com.example.inner.NestedObject
  object NestedObject {
    def nestedMethod(): Unit = ()
  }
}

// Top level type alias, symbol: com.example.MyAlias
type MyAlias = MyCaseClass

// Top level implicit class, symbol: com.example.MyImplicitClass
implicit class MyImplicitClass(val s: String) {
def method5(): String = s.toUpperCase
}

// Top level implicit def, symbol: com.example.stringToInt
implicit def stringToInt(s: String): Int = s.toInt

case class MinimizedCaseClass(value: String) {
  def this() = this(value = "value")
}
object MinimizedCaseClass {
  def main(): Unit = {
    println(MinimizedCaseClass.apply(value = "value1").copy(value = "value2").value)
  }
}

trait MinimizedTrait[T] extends AutoCloseable {
  def add(e: T): T
  final def +(e: T): T = add(e)
}

class MinimizedScalaSignatures extends AutoCloseable with java.io.Serializable {
  def close(): Unit = ()
}

object MinimizedScalaSignatures extends MinimizedScalaSignatures with Comparable[Int] {
  @inline def annotation(x: Int): Int = x + 1
  @deprecated("2020-07-27") def annotationMessage(x: Int): Int = x + 1
  def compareTo(x: Int): Int = ???
  def identity[T](e: T): T = e
  def tuple(): (Int, String) = null
  def function0(): () => String = null
  def function1(): Int => String = null
  def function2(): (Int, String) => Unit = null
  def typeParameter(): Map[Int, String] = null
  def termParameter(a: Int, b: String): String = null
  def singletonType(e: String): e.type = e
  def thisType(): this.type = this
  def constantInt(): 1 = 1
  def constantString(): "string" = "string"
  def constantBoolean(): true = true
  def constantFloat(): 1.2f = 1.2f
  def constantChar(): 'a' = 'a'
  def structuralType(): { val x: Int; def foo(a: Int): String } = null
  def byNameType(a: => Int): Unit = ()
  def repeatedType(a: Int*): Unit = ()

  type TypeAlias = Int
  type ParameterizedTypeAlias[A] = () => A
  type ParameterizedTypeAlias2[A, B] = A => B
  type TypeBound
  type TypeUpperBound <: String
  type TypeLowerBound >: CharSequence
  type TypeLowerUpperBound >: String <: CharSequence
}
