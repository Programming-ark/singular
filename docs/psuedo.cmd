# Singularity Command Reference

## Repository

```bash
singularity init
singularity init --advanced

singularity validate

singularity status

singularity health

singularity audit

singularity history

singularity version

singularity doctor

singularity repair

singularity config

singularity config get
singularity config set

singularity policy

singularity export
singularity import
```

---

## Domain Management

```bash
singularity domain create

singularity domain delete

singularity domain update

singularity domain rename

singularity domain list

singularity domain show

singularity domain tree

singularity domain validate
```

### Domain Ownership

```bash
singularity domain paths

singularity domain paths add

singularity domain paths remove

singularity domain paths list
```

### Domain Governance

```bash
singularity domain trust

singularity domain trust add

singularity domain trust remove

singularity domain trust list

singularity domain approve

singularity domain unapprove

singularity domain lock

singularity domain unlock
```

### Domain Relationships

```bash
singularity domain parent

singularity domain children

singularity domain ancestry

singularity domain descendants
```

---

## Branch Context

```bash
singularity checkout

singularity current

singularity context

singularity contexts

singularity switch

singularity root
```

---

## Reflection

```bash
singularity reflect

singularity reflect create

singularity reflect list

singularity reflect show

singularity reflect rename

singularity reflect archive

singularity reflect restore

singularity reflect delete

singularity reflect consolidate
```

### Reflection Metadata

```bash
singularity reflect scopes

singularity reflect scopes add

singularity reflect scopes remove

singularity reflect scopes list
```

### Reflection Lineage

```bash
singularity reflect parent

singularity reflect children

singularity reflect ancestry

singularity reflect descendants
```

---

## Lineage

```bash
singularity lineage

singularity lineage show

singularity lineage tree

singularity lineage graph

singularity lineage parents

singularity lineage children

singularity lineage roots

singularity lineage validate
```

---

## Consolidation

```bash
singularity consolidate

singularity consolidate approve

singularity consolidate reject

singularity consolidate list

singularity consolidate show

singularity consolidate history
```

---

## Handoffs

```bash
singularity handoff

singularity handoff create

singularity handoff show

singularity handoff list

singularity handoff update

singularity handoff cancel

singularity handoff archive
```

### Handoff Review

```bash
singularity handoff approve

singularity handoff reject

singularity handoff reopen

singularity handoff claim

singularity handoff release
```

### Handoff Discovery

```bash
singularity handoff incoming

singularity handoff outgoing

singularity handoff pending

singularity handoff active

singularity handoff expired

singularity handoff completed
```

### Handoff Lifecycle

```bash
singularity handoff history

singularity handoff lineage

singularity handoff audit
```

---

## Inbox

```bash
singularity inbox

singularity inbox list

singularity inbox show

singularity inbox clear
```

### Inbox Filters

```bash
singularity inbox handoffs

singularity inbox approvals

singularity inbox violations

singularity inbox consolidations
```

---

## Governance

```bash
singularity governance

singularity governance status

singularity governance validate

singularity governance audit

singularity governance check
```

### Governance Policies

```bash
singularity governance policy

singularity governance policy create

singularity governance policy update

singularity governance policy delete

singularity governance policy show

singularity governance policy list
```

### Governance Violations

```bash
singularity governance violations

singularity governance violations list

singularity governance violations show

singularity governance violations resolve
```

---

## Ownership

```bash
singularity ownership

singularity ownership show

singularity ownership validate

singularity ownership transfer

singularity ownership history
```

---

## Scope Management

```bash
singularity scope create

singularity scope delete

singularity scope update

singularity scope list

singularity scope show
```

### Scope Assignment

```bash
singularity scope assign

singularity scope unassign

singularity scope assignments
```

---

## Dependency Governance

```bash
singularity dependency

singularity dependency validate

singularity dependency graph

singularity dependency audit
```

### Dependency Rules

```bash
singularity dependency allow

singularity dependency deny

singularity dependency list

singularity dependency remove
```

### Dependency Analysis

```bash
singularity dependency incoming

singularity dependency outgoing

singularity dependency violations
```

---

## Events

```bash
singularity events

singularity events list

singularity events show

singularity events export
```

### Event Filters

```bash
singularity events handoffs

singularity events governance

singularity events lineage

singularity events domains

singularity events violations
```

---

## Branch Registry

```bash
singularity branches

singularity branches list

singularity branches show

singularity branches active

singularity branches archived
```

### Branch Metadata

```bash
singularity branches lineage

singularity branches scopes

singularity branches domains

singularity branches history
```

---

## Approval Workflows

```bash
singularity approvals

singularity approvals pending

singularity approvals approve

singularity approvals reject

singularity approvals history
```

---

## Analytics

```bash
singularity metrics

singularity metrics domains

singularity metrics handoffs

singularity metrics lineage

singularity metrics governance
```

---

## Diagnostics

```bash
singularity diagnose

singularity verify

singularity integrity

singularity consistency

singularity doctor
```

---

## Search

```bash
singularity search

singularity search branches

singularity search domains

singularity search handoffs

singularity search events

singularity search scopes
```

---

## Export

```bash
singularity export policy

singularity export lineage

singularity export events

singularity export domains

singularity export handoffs
```

---

## Import

```bash
singularity import policy

singularity import lineage

singularity import domains

singularity import handoffs
```


One architectural observation: based on everything we've discussed, I suspect the final CLI will be much smaller than this. Commands such as:

singularity branches
singularity ownership
singularity governance


may eventually become internal concepts because the primary user-facing workflows seem to be converging around:

singularity init
singularity checkout
singularity reflect
singularity status
singularity commit
singularity consolidate
singularity handoff
singularity inbox
singularity lineage
singularity audit
singularity health


with most other actions exposed as subcommands of those core concepts. That's usually a sign of a strong CLI design.