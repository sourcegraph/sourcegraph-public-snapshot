// Copyright Sam Xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otelsql

// Method specifics operation in the database/sql package.
type Method string

// Event specifics events in the database/sql package.
type Event string

const (
	MethodConnectorConnect Method = "sql.connector.connect"
	MethodConnPing         Method = "sql.conn.ping"
	MethodConnExec         Method = "sql.conn.exec"
	MethodConnQuery        Method = "sql.conn.query"
	MethodConnPrepare      Method = "sql.conn.prepare"
	MethodConnBeginTx      Method = "sql.conn.begin_tx"
	MethodConnResetSession Method = "sql.conn.reset_session"
	MethodTxCommit         Method = "sql.tx.commit"
	MethodTxRollback       Method = "sql.tx.rollback"
	MethodStmtExec         Method = "sql.stmt.exec"
	MethodStmtQuery        Method = "sql.stmt.query"
	MethodRows             Method = "sql.rows"
)

const (
	EventRowsNext Event = "sql.rows.next"
)
