#/bin/bash
#
#  First install the jira-cli:
#
#    go install github.com/ankitpokhrel/jira-cli/cmd/jira@latest
#
#  then source this file (jira-cli.sh) from your shell to provide
#  some handy jira-cli driver functions.
#
#  Enter jira<TAB> to see completions.

# export JIRA_PROJECT=BOB
# export JIRA_HOST=issues.acmecorp.com

function jiraTestBashBehavior {
    if [ $# -lt 2 ]; then
        echo "Need a label and quoted string"
        return 0
    fi
    echo "Got \"$1\" and \"$2\""
}

function jiraStoryCreate {
    if [ $# -ne 2 ]; then
        echo "Need a label and quoted summary string"
        return 0
    fi
    local label=$1
    local summary=$2
    local body="${3:-TBD}"
    jira issue create \
         --no-input \
         --type Story \
         --label "${label}" \
         --summary "${summary}" \
         --body "${body}"
}

# In epics, the "name" and "summary" are the same
# https://community.atlassian.com/t5/Jira-questions/Epic-name-vs-Epic-Summary-Do-we-need-both/qaq-p/850442
#
function jiraEpicCreate {
    if [ $# -ne 1 ]; then
        echo "Need a quoted summary string"
        return 0
    fi
    local summary=$1
    tmp=$( jira epic create \
         --no-input \
         --name "${summary}" \
         --summary "${summary}" \
         --body "TBD" )
    num=$(echo $tmp | sed "s|.*/${JIRA_PROJECT}-\(.*\)|\1|")
    echo $num
}

function jiraDelete {
    for issue in "$@"; do
        jira issue delete ${JIRA_PROJECT}-${issue}
    done
}

function jiraMoveToBacklog {
  for issue in "$@"; do
    jira issue move ${JIRA_PROJECT}-${issue} Backlog
  done
}

function jiraMoveToReady {
  for issue in "$@"; do
    jira issue move ${JIRA_PROJECT}-${issue} "Ready for Development"
  done
}

function jiraMoveToDone {
  for issue in "$@"; do
    jira issue move ${JIRA_PROJECT}-${issue} "Done"
  done
}

function jiraLabelAdd {
    local label=$1
    shift
    for issue in "$@"; do
        jira issue edit ${JIRA_PROJECT}-${issue} --no-input --label $label
    done
}

function jiraLabelRemove {
    local label=$1
    shift
    for issue in "$@"; do
        jira issue edit ${JIRA_PROJECT}-${issue} --no-input --label -$label
    done
}

function jiraLabelList {
    jira --project ${JIRA_PROJECT} issue list \
         --columns type,key,summary,status,assignee,reporter,labels,created \
         --plain \
         --status ~"Done" --status ~"Closed Without Action" \
         --label $1
}

function jiraLabelNot {
    local nope=""
    for lab in "$@"; do
        nope+=" --label ~$lab"
    done
    jira issue list \
         --status ~"Done" --status ~"Closed Without Action" \
         $nope
}

function jListNoOkrLabel {
  jLabNot \
    okrCadCleanup \
    okrTestFramework \
    okrSupportFramework \
    okrSearchPerformance \
    okrDataGovernance \
    okrQuantifyUtilization \
    okrReact \
    okrNone
}

function jiraDoABunchOfLabelling {
  jiraLabelAdd  okrCadCleanup \
           273 272 271 270 269 268 250 \
           293 292 291

  jiraLabelAdd  okrTestFramework \
           275 274 228 227 225 224 \
           218 215 85 142

  jiraLabelAdd  okrSupportFramework \
           290 287 286 285 260 254 \
           244 243 242 231 230 229 226 \
           210 75 73 6

  jiraLabelAdd  okrSearchPerformance \
           38 36 245 241 237 236 234 233

  # jiraLabelAdd  okrDataGovernance none

  jiraLabelAdd  okrQuantifyUtilization \
           204 10 255

  jiraLabelAdd  okrReact \
           289 288 282 281 279 277 259 258 \
           257 252 251 235 232 222 \
           202 163 123

  jiraLabelAdd  okrNone \
           284 283 280 278 276 267 266 265 264 263 \
           262 261 256 253 249 248 247 240 238 \
           221 212 191 175 136  135 113 112 81 29
}

function jiraDeleteByLabel {
    IFS=$'\n' issues=( $(jira issue list \
         --no-headers --plain \
         --columns key \
         --label $1) )
    # declare -p issues
    for issue in "${issues[@]}"; do
        echo "jira issue delete ${issue}"
    done
}


# FOR jql see: https://support.atlassian.com/jira-service-management-cloud/docs/jql-fields/#Summary

function jiraDeleteBigMess {
    IFS=$'\n' issues=( $(jira issue list \
         --no-headers --plain \
         --columns key \
         --jql "reporter = asivanadi AND created >= -9d and summary ~ \"Release 21\"") )
    for issue in "${issues[@]}"; do
        echo "jira issue delete ${issue}"
        jira issue delete ${issue}
    done
}

function jiraInit {
  if [ x"${JIRA_API_TOKEN}" == "x" ]; then
    export JIRA_API_TOKEN=$JIRA_TOKEN
  fi
  if [ x"${JIRA_API_TOKEN}" == "x" ]; then
    echo "You forgot:"
    echo "  "
    echo "  export JIRA_API_TOKEN=whatever"
    echo "  "
    return
  fi
  if [ x"${JIRA_HOST}" == "x" ]; then
    echo "You forgot:"
    echo "  "
    echo "  export JIRA_HOST=whatever"
    echo "  "
    return
  fi
  if [ x"${JIRA_PROJECT}" == "x" ]; then
    echo "You forgot:"
    echo "  "
    echo "  export JIRA_PROJECT=whatever"
    echo "  "
    return
  fi

  jira init \
      --force \
      --installation local \
      --server https://${JIRA_HOST}  \
      --auth-type bearer \
      --login "${JIRA_EMAIL}" \
      --project "${JIRA_PROJECT}" \
      --board "${JIRA_PROJECT}-1"


  echo "Example commands:"
  echo " "
  echo "  jira project list | egrep '[dD]esign|PLM'"
  echo "  jira epic list --table --plain"
  echo "  jira board list"
  echo "  jira --project CIA board list"
  echo "  jira --project ${JIRA_PROJECT} issue list --updated-after 2024/08/01"
}

function jiraTransitions {
  curl -X GET \
       -H "Authorization: Bearer ${JIRA_API_TOKEN}" \
       -H "Accept: application/json" \
       "https://${JIRA_HOST}/rest/api/2/issue/${JIRA_PROJECT}-1/transitions" |\
       jq .
}

function jiraStatus {
  curl -X GET \
       -H "Authorization: Bearer ${JIRA_API_TOKEN}" \
       -H "Accept: application/json" \
       "https://${JIRA_HOST}/rest/api/2/status" |\
       jq .
}

# Lastly, run the init function to set up defaults.
# jiraInit
