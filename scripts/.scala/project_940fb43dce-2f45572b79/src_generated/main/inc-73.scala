



object `inc-73` {
/*<script>*/import java.util.concurrent.TimeUnit
import java.util.concurrent.Executors
import java.util.concurrent.ArrayBlockingQueue
import java.io.PrintStream
import java.util.Base64
import java.nio.file.Paths
import java.nio.file.Files
import java.nio.ByteBuffer
import scala.jdk.CollectionConverters._
import scala.sys.process._


def base64(id: String): String =new String(Base64.getEncoder.encode(s"Repository:$id".getBytes))
val lines = Files.readAllLines(Paths.get("/Users/olafurpg/Downloads/repo_id_versions.csv")).asScala.iterator.map(_.split(","))
lines.drop(1) // skip header
val donePath = Paths.get("/Users/olafurpg/Downloads/done.txt")
val isDone = Files.readAllLines(donePath).asScala.toSet
val done = new PrintStream(Files.newOutputStream(donePath))


val sh = Executors.newSingleThreadScheduledExecutor.scheduleAtFixedRate(() => {
lines.collect {
    case Array(id, version) if !isDone(id) =>
        val repo = base64(id)
        val rev = s"v$version"
        val result = List("src", "api", "-query='mutation QueueAutoIndexJobsForRepo($id: ID!, $rev: String) { queueAutoIndexJobsForRepo(repository: $id, rev: $rev) { id } }'", s"id=$repo", s"rev=$rev").!!
        done.println(id)
        // repo -> result
        (repo, id, rev, result)
}.take(1).foreach(println)
}, 0, 1, TimeUnit.SECONDS)


sh.wait()

/*</script>*/ /*<generated>*/
def args = `inc-73_sc`.args$
  /*</generated>*/
}
object `inc-73_sc` {
  private var args$opt0 = Option.empty[Array[String]]
  def args$set(args: Array[String]): Unit = {
    args$opt0 = Some(args)
  }
  def args$opt: Option[Array[String]] = args$opt0
  def args$: Array[String] = args$opt.getOrElse {
    sys.error("No arguments passed to this script")
  }
  def main(args: Array[String]): Unit = {
    args$set(args)
    `inc-73`.hashCode() // hasCode to clear scalac warning about pure expression in statement position
  }
}

