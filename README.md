# bloXroute test project

Consists of client and server.

Uses Amazon SQS.

Amason SQS message have attributes:
* `method`: string, can have values only `AddItem`, `RemoveItem`, `GetItem`, `GetAllItems`;
* `name`: string, name of data entry. Only for `AddItem`, `RemoveItem` and `GetItem`;
* *Message body*: string, contains value of data for `AddItem`.

Messages are sent, received and deleted in batches of 10 or less (SQS feature).

When `AddItem` overwrites data, order is preserverd.

Messages are processed by server in threads. Number of threads is defined by number of CPUs or by `NTHREADS` environment variable it it is set.

## Building

```
git clone git@github.com:cbytensky/bloxroute.git
cd bloxroute
go install ./...
```

## Running

### Server

```
export AWS_ACCESS_KEY_ID=EXAMPLEKEYID
export AWS_SECRET_ACCESS_KEY=AbCdEfExampleAccessKey
export AWS_REGION=us-east-2
export QUEUE_URL=https://sqs.us-east-2.amazonaws.com/302642057669/test
$GOPATH/bin/server
```

### Client

```
export AWS_ACCESS_KEY_ID=EXAMPLEKEYID
export AWS_SECRET_ACCESS_KEY=AbCdEfExampleAccessKey
export AWS_REGION=us-east-2
export QUEUE_URL=https://sqs.us-east-2.amazonaws.com/302642057669/test
# cmdline
$GOPATH/bin/client +a:A +b:B +c:C
# stdin
$GOPATH/bin/client <<EOF
.a
-b
!
EOF
```

## Client commands

Commands passed to client both by command line arguments and stdin (only if it is not TTY). Commands are in such format:
* **AddItem**: <code><strong>+</strong>⟨name⟩<strong>:</strong>⟨value⟩</code>
* **RemoveItem**: <code><strong>-</strong>⟨name⟩</code>
* **GetItem**: <code><strong>.</strong>⟨name⟩</code>
* **GetAllItems**: <code><strong>!</strong></code>


## Parameters

Both client and server uses `QUEUE_URL` environment variable which must me set to SQS message queue URL.

Also AWS API credentials and AWS region can be also defined by environment variables `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` and `AWS_REGION`.

S
