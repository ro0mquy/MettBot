MettBot
=======

This is the bot serving #mett on the EosinNet IRC Network (`irc://irc.ps0ke.de:2342/#mett`).

Commands
--------

* `!quote <$nick> $quote` -- add a new quote to the database, timestamp is added automagically
* `!print $interger` -- print a quote from the database
* `!help seriöslich` -- show help text

Misc
----

### Topic Diffing

Whenever the topic is changed, the bot diffs the new and the old one using [`wdiff`](https://www.gnu.org/software/wdiff/).

### Bugs and Improvements

Feel free to crash the bot or make a pull request.
