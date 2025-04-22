[Jira]: https://www.atlassian.com/software/jira
[github.com/ankitpokhrel/jira-cli]: https://github.com/ankitpokhrel/jira-cli
[`dot`]: https://graphviz.org/docs/layouts/dot
[`jira-cli`]: #jira-cli-advertisment
[`jiraboss.go`]: internal/myj/jiraboss.go
[github.com/ankitpokhrel/jira-cli]: https://github.com/ankitpokhrel/jira-cli
[REST API]: https://developer.atlassian.com/server/jira/platform/rest
[snips]: https://github.com/monopole/snips
[doomed]: https://github.com/search?q=gojira+in%3Aname&type=Repositories

# gojira


`gojira` helps me do specific [Jira] tasks.
<img src="internal/utils/gojira.jpg" align="right" height="180" width="140">
I'm not trying to generalize `gojira` or make the underlying
code publishable as a pkg.

Everyone who knows some Go and must use jira is [doomed] to write a
CLI called `gojira`.  For a general jira CLI, use [`jira-cli`].

The API is wide enough that there's more than one way to do a
particular task, so making this public. Feel free to fork this and
adapt [`jiraboss.go`].

### Why write this?

It seems that Jira evolved from simple issue tracking
to project management. Jira takes no sides, opting to
offer more customization options rather than constrain
users to a specific management practice.  Jira now has
an organic, sprawling [REST API].

Allowing so many options means complex, slow or perhaps
missing functionality in the canonical Jira web UX,
especially if using custom fields or doing multistep
ops like generating dependency graphs and cascading
date management. [`jira-cli`], although broad and fast,
obviously can't know how or why you've customized Jira.

To meet the reporting demands of a specific team with
legacy Jira customizations I often needed to do
multistep operations using custom fields.  This was
slow and error-prone in the Web app.  I started writing
shell to drive `jira-cli`, `curl`, `awk` etc. which was
better than the web UX, but the code got ugly.
Eventually the ugly drove me to re-use some Go I
already had ([snips]) to hit the Jira API directly and do
processing solely in Go.

### Installation

The parts of this that are interesting to me won't work
for you.  It relies on custom fields to define
start-end dates and group stories into epics.  Maybe
fork and adapt it; just change where it gets dates and
you're mostly done getting the `epic cal` and [`dot`]
commands to work.

Nevertheless you can see if it compiles:

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
