package foobar

import scala.collection.immutable.List

// Comment
case class Foo(x: Int, y: String) extends AnyVal
/** Docstring */
object Foo {
  val x: Int = 42
  val y: Double = 42.0
  val z = s"hello $x" + "hello world"
  val a: Map[Int, Int] = Map.empty
  val b: Foo = Foo(x = 42)
  lazy val c = 'a'
  var d = 1.4f
  val e = true
  val f = null
  def main(args: Array[String]): Unit = {
    println(args.toList)
    System.out.println(args.toList)
    args(1).indexOf("a")
    args.toList match {
        case 1 :: 2 :: Nil =>
        case a :: Nil =>
        case Some(x) =>
    }
  }
  private def privateMethod = 42
  protected def protectedMethod = 42
  private[this] def privateThisMethod = 42
  private[foobar] def privatePackageMethod = 42
  type MyMap[A, B] = Map[A, B]
  trait MyTrait[T] extends SuperTrait[T]
  enum X { case A, B }
}
