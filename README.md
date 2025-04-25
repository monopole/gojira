[Jira]: https://www.atlassian.com/software/jira
[github.com/ankitpokhrel/jira-cli]: https://github.com/ankitpokhrel/jira-cli
[`dot`]: https://graphviz.org/docs/layouts/dot
[`jira-cli`]: #jira-cli-advertisment
[`jiraboss.go`]: internal/myj/jiraboss.go
[custom fields]: internal/myj/customfields.go
[github.com/ankitpokhrel/jira-cli]: https://github.com/ankitpokhrel/jira-cli
[REST API]: https://developer.atlassian.com/server/jira/platform/rest
[snips]: https://github.com/monopole/snips
[doomed]: https://github.com/search?q=jira+in%3Areadme+and+gojira+in%3Aname+and+language%3Ago&type=repositories&ref=advsearch
[yet another]: https://github.com/search?q=jira+in%3Areadme+and+gojira+in%3Aname+and+language%3Ago&type=repositories&ref=advsearch

# gojira


`gojira` helps me do specific [Jira] tasks.
<img src="internal/utils/gojira.jpg" align="right" height="180" width="140">

Everyone who knows some Go, must use Jira,
and must use Jira customizations is [doomed] to write a
CLI called `gojira`.  The Jira web UX
cannot offer good and fast UX for all the
things one might do (or by circumstance be encouraged to do)
with the organic, sprawling Jira [REST API].

I'm not trying to generalize `gojira` or make the underlying
code publishable as a pkg, but I am trying to keep
it tidy. For a fine general purpose Jira CLI try [`jira-cli`].

### Installation

The parts of this that are interesting to me won't work
for you.  It relies on [custom fields] to define
start-end dates and group stories into epics.
Feel free to fork and
adapt [`jiraboss.go`]. Change where it gets _dates_ and
you're mostly done getting the `epic cal` and [`dot`]
commands to work.

Nevertheless, it should install with:

```bash
go install github.com/monopole/gojira@latest
```
See capabilities with

```bash
gojira help
```

To avoid specifying flags, use:
```bash
# domain of the host running the Jira web UX and REST API
export JIRA_HOST=jira.acmecorp.com
# an API token obtained from JIRA_HOST
export JIRA_API_TOKEN=whatever
# your favorite Jira project
export JIRA_PROJECT=BOB
```

Get a value for `JIRA_API_TOKEN` from

> https://${JIRA_HOST}/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens


### jira-cli (_advertisment_)

> [github.com/ankitpokhrel/jira-cli]

is a super useful general jira tool and can be installed with

```bash
go install github.com/ankitpokhrel/jira-cli/cmd/jira@latest
```

The app installs as a binary called  `jira`.

Create an epic list:

```bash
jira epic list --order-by key --reverse --plain
```

Look at one issue:
```bash
jira issue view LEMON-33
```

### Why [yet another] gojira?
Once you learn the API, it's easier to use 
than the Jira web UX.

I needed
* particular reporting:
  * dependency graph images for long term sanity check and
  * fast terminal-based calendars for daily sanity checks,
* [custom fields],
* safe date checking and repair.

`jira-cli`, although broad and fast,
can't know how and why you've customized Jira and
serve those needs well.
This effort started as
shell to drive `jira-cli`, `curl`, `awk` etc.
Unwieldy shell code led to a
switch to Go to hit the Jira API directly reusing
some old code ([snips]).
