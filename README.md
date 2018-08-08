# Randomizer

```
go get go.alexhamlin.co/randomizer/...
```

The randomizer is, quite simply, a tool that chooses a random option from a
list.

The most interesting thing about this particular randomizer is that it can work
on the command line *and* as a Slack slash command, using the same underlying
implementation (which is seriously great for testing).

The second most interesting thing is that it lets you save groups of options
for re-use later. So if you tend to randomize from the same set of things all
the time, it's now way easier.

Right now this is in sort of a prototype / beta state.

## Upcoming

* Get feedback on the current implementation / usage patterns (e.g. is
  per-channel really the best way to save groups, or is there a better option?)
* Try to get the service running on AWS Lambda, just like the current
  randomizer (makes maintenance easier, especially for others; the non-Lambda
  HTTP server will still be available though)
