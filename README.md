MettBot
=======

This is the bot serving #mett on the EosinNet IRC Network (`irc://irc.ps0ke.de:2342/#mett`).

Commands
--------

* `!quote <$nick> $quote` -- add a new quote to the database, timestamp is added automagically
* `!print $interger` -- print a quote from the database
* `!search $regex` -- searches the quote database for the regex
* `!mett` -- post a random entry from the mett database
* `!mett $mettcontent` -- add new mettcontent to the mett database
* `!help seriöslich` -- show help text

Misc
----

### Topic Diffing

Whenever the topic is changed, the bot diffs the new and the old one using [`wdiff`](https://www.gnu.org/software/wdiff/).

### Mett Content

If there is no appearance of the word Mett for a certain time or number of messages, the bot posts a random entry from the mett database.

### Showing Tweets

The bot automagically detects and prints the content of tweets.

### Bugs and Improvements

Feel free to crash the bot or make a pull request.
