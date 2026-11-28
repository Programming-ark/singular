# Singularity

**Governance-first source control built on top of Git.**

Singularity introduces a virtual repository layer that sits between developers and Git, enabling organizations to enforce repository governance, domain ownership, branch policies, and cross-team workflows without depending on discipline, conventions, or manual review.

Instead of allowing contributors to interact directly with Git branches, Singularity manages repository state through logical branches, governance policies, and a hidden Git context. Developers work within a projected workspace while Singularity orchestrates the underlying Git operations.

---

## Why Singularity Exists

Modern repositories often suffer from common governance problems:

- Teams accidentally modify areas outside their ownership.
- Feature branches contain unrelated changes.
- Architectural boundaries become blurred over time.
- Domain experts become bottlenecks for reviewing unrelated work.
- Repository rules rely on conventions rather than enforcement.

Git provides excellent version control but has limited understanding of repository governance.

For example, Git can protect:

- branches
- tags
- merges

But Git cannot naturally enforce policies such as:

> A finance branch may only modify files inside the finance domain.

> User-service contributors may not modify billing-service code without an approved handoff.

> A feature branch may only contain changes related to a single repository domain.

Singularity solves this by introducing a governance layer above Git.

---

## Core Philosophy

Developers interact with Singularity:

```bash
singularity checkout finance
singularity commit
singularity handoff users
singularity merge
```

Developers do **not** interact with the underlying Git repository.

Git becomes an implementation detail.

Governance becomes the primary interface.

---

## Virtual Repository Architecture

Unlike traditional repositories, Singularity separates the developer workspace from the physical Git repository.

Example:

```text
project/
├── finance/
├── users/
├── billing/
└── .singularity/
    ├── repo/
    │   └── .git/
    ├── policy.yaml
    ├── domains.json
    ├── branch.db
    └── snapshot.db
```

The workspace contains the files developers edit.

The `.singularity` directory contains:

- governance policies
- domain definitions
- branch mappings
- repository metadata
- workspace snapshots
- the hidden Git repository

Direct Git interaction is intentionally abstracted away to prevent governance bypass.

---

## Managed Physical Branches

Managed Physical Branches are mandatory and always enabled.

Developers work with **logical branches**:

```text
main
finance
users
billing
```

Internally, Singularity manages **physical branches**:

```text
sg-main-d3f2c
sg-finance-a82f4c
sg-users-c91ab1
sg-billing-f72a9d
```

Logical branches are the stable identifiers visible to developers and governance policies.

Physical branches are implementation details managed exclusively by Singularity.

This separation allows Singularity to:

- enforce governance consistently
- orchestrate cross-domain workflows
- maintain branch metadata
- support recovery operations
- preserve branch topology history
- evolve workflows without exposing complexity

Contributors never interact directly with physical branches.

---

## Domains

Domains represent ownership boundaries within a repository.
Both standard and advanced initialization require at least one domain.

### Define domains during initialization

- `singularity init` asks for the root branch and repository type, then starts domain setup.
- `singularity init --advanced` asks for the root branch and optional protected branches, then starts domain setup. It also asks about consolidation, direct merges to the root branch, and trusted domains.

Domains defined during initialization are first-level domains under the repository
root. Policy inheritance is reserved for later creation of a child domain
under an existing domain with governance rules.

Scopes are intended for work on branches created after initialization. They
describe the area of work within a domain, while the domain defines ownership.
See [Scope-based reflection](../README.md#scope-based-reflection) for the planned
workflow.

Name each domain for the ownership boundary it represents, such as a directory,
service, component, or team. For example, entering `users` prefills **Owned paths**
with `users/**`. Press Enter to accept it or edit the field to enter comma-separated
path patterns. Choose **Create another domain?** to define additional boundaries.

After confirmation, the wizard writes `.singularity/policy.yaml` in the current
directory. It preserves any existing policy file. The root branch is always
included in the protected branch list. In standard setup, choosing **Create
Handoff** requires approval and leaves handoffs active indefinitely.

Advanced initialization asks whether to enable dependency governance, which
controls which domains may depend on each other. Governance event recording is
always enabled by design. Domain inboxes and health metrics are derived from
governance events and repository state.

Example:

```yaml
domains:
  finance:
    paths:
      - finance/**

  users:
    paths:
      - users/**

  billing:
    paths:
      - billing/**
```

Domains define:

- ownership
- allowed file locations
- governance boundaries
- handoff destinations

---

## Branch Governance

Logical branches are bound to one or more domains.

Example:

```text
finance
    └── finance/**
```

Allowed:

```text
finance/payroll.ts
finance/tax.ts
```

Rejected:

```text
users/profile.ts
billing/invoice.ts
```

Every commit is validated against repository policy before entering the managed Git context.

---

## Cross-Domain Handoffs

Many changes naturally cross organizational boundaries.

Instead of allowing unrestricted modification of another domain, contributors initiate handoffs.

Example:

```bash
singularity handoff users
```

This creates a governance event indicating that work must continue in another domain.

Example topology:

```text
finance
    │
    ▼
handoff-users
    │
    ▼
users
```

Organizations may configure:

- approval requirements
- ownership workflows
- handoff expiration policies
- merge requirements
- audit requirements

Every handoff becomes a traceable governance event.

When you select **Handoff** in advanced initialization, the wizard asks whether
handoff merges require approval and whether handoffs expire. If expiration is
enabled, it asks for the number of days. Singularity manages physical handoff
branch names internally.

---

## Synchronization Model

Singularity intentionally avoids permanent background daemons.

Repository synchronization occurs on demand.

When commands such as:

```bash
singularity status
singularity commit
```

are executed, Singularity:

1. Scans the workspace.
2. Detects modified files.
3. Validates governance policies.
4. Synchronizes approved changes into the hidden Git repository.
5. Executes the required Git operations.
6. Updates workspace snapshots.

This approach minimizes overhead while maintaining strict governance.

---

## Auditability

Every governance action is recorded.

Examples include:

- branch creation
- branch mapping changes
- domain creation
- domain updates
- policy changes
- handoff initiation
- handoff approval
- merges
- recovery actions

This creates a complete governance history independent of individual Git commits.

---

## Example Workflow

Initialize a repository:

```bash
singularity init
```

Create or switch to a logical branch:

```bash
singularity checkout finance
```

View repository status:

```bash
singularity status
```

Commit validated changes:

```bash
singularity commit
```

Initiate a cross-domain workflow:

```bash
singularity handoff users
```

View governance activity:

```bash
singularity audit
```

---

## Design Goals

- Enforce repository boundaries technically rather than socially.
- Prevent unauthorized cross-domain modifications.
- Make ownership explicit and enforceable.
- Provide traceable collaboration workflows.
- Scale repository governance without increasing complexity.
- Preserve Git's strengths while adding governance abstractions.

---

## Vision

Singularity is not a replacement for Git.

Git remains the storage engine.

Singularity becomes the governance engine.

Just as Kubernetes abstracts infrastructure into declarative resources and policies, Singularity abstracts repository governance into domains, logical branches, policies, handoffs, and managed workflows.

The long-term goal is to make repository governance explicit, enforceable, auditable, and scalable while preserving the flexibility developers expect from distributed version control systems.
