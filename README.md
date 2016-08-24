roll
----

`roll` is a simple Rollbar client for Go that makes it easy to report errors and
messages to Rollbar. It supports all Rollbar error and message types, stack
traces, and automatic grouping by error type and location. For more advanced
functionality, check out [heroku/rollbar](https://github.com/heroku/rollbar).

All errors and messages are sent to Rollbar synchronously.

[API docs on godoc.org](http://godoc.org/github.com/stvp/roll)

Notes
=====

* Critical-, Error-, and Warning-level messages include a stack trace. However,
  Go's `error` type doesn't include stack information from the location the
  error was set or allocated. Instead, `roll` uses the stack information from
  where the error was reported.
* Info- and Debug-level Rollbar messages do not include stack traces.  
* When calling `roll` away from where the error actually occurred, `roll`'s
  stack walking won't represent the actual stack trace at the time the error
  occurred. The `WithStack` variants of Critical, Error, and Warning take a
  `[]uintptr`, allowing the stack to be provided, rather than walked.

Running Tests
=============

`go test` will run tests against a fake server by default.

If the environment variable `TOKEN` is a Rollbar access token, running
`go test` will produce errors using an environment named `test`. 

    TOKEN=f0df01587b8f76b2c217af34c479f9ea go test

Verify the reported errors manually in the Rollbar dashboard.

Contributors
============

This library was forked from [stvp/rollbar](https://github.com/stvp/rollbar),
which had contributions from:

* @kjk
* @Soulou
* @paulmach

