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

# gojira


`gojira` helps me do specific [Jira] tasks.
<img src="internal/utils/gojira.jpg" align="right" height="180" width="140">

Everyone who knows some Go, must use Jira,
and must use Jia customizations is [doomed] to write a
CLI called `gojira`.  The Jira web UX
simply cannot offer good and fast UX for all the
things one might do (or be forced to do by circumstance)
with the organic, sprawling Jira [REST API].

I'm not trying to generalize `gojira` or make the underlying
code publishable as a pkg, but I am trying to keep
it tidy. For a great _general purpose_ Jira CLI see [`jira-cli`].

### Installation

The parts of this that are interesting to me won't work
for you.  It relies on [custom fields] to define
start-end dates and group stories into epics.
Feel free to fork and
adapt [`jiraboss.go`]. Change where it gets dates and
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
export JIRA_HOST=jira.acmecorp.com      # domain of the host running the Jira web UX and REST API
export JIRA_API_TOKEN=whatever          # an API token obtained from JIRA_HOST
export JIRA_PROJECT=BOB                 # your favorite Jira project
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

### Why do yet another gojira?

Because of particular reporting needs (dependency
graphs), a need for flexible yet safe epic date management,
and a need for [custom fields].

This effort started as
shell to drive `jira-cli`, `curl`, `awk` etc. to do
what I needed to do.
`jira-cli`, although broad and fast,
can't know how and why you've customized Jira and
serve those needs well.

The shell code got unwieldly, so I
switched to Go to hit the Jira API directly reusing
some old code ([snips]).
