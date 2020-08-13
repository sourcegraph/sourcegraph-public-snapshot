+package db
+
+import "github.com/pkg/errors"
+
+// sessions provides access to the sessions table.
+type sessions struct{}
+
+// ErrorNoSuchSession is returned when we can't retrieve the specific session object
+var ErrorNoSuchSession = errors.New("failed to find session")
+
+func GetByUserID() {}
+
+func DeleteByUserID() {}
+
+func DeleteBySessionName() {}
+
+func GetBySessionName() {}
+
+func InsertSession() {}
