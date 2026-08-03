---
name: update-docs
description: Synchronize all repository documentation with the work completed through the current chapter, including newly added, renamed, and existing documents, then validate and commit the intended chapter changes. Use when the user invokes $update-docs, asks to update or sync docs at a chapter boundary, or wants the completed chapter documented and committed. Do not use for ordinary single-file documentation edits or when the user only asks for a report without repository changes.
---

# Update Docs

Bring the repository's full documentation set in line with verified current state, then create one reviewed chapter commit. Discover documents anew on every run; never rely on a fixed file list.

## 1. Establish scope

1. Resolve the repository root and read every applicable `AGENTS.md`.
2. Inspect `git status --short --branch`, recent commits, and `JOURNEY.md` when present.
3. Determine the chapter or milestone from `JOURNEY.md`, branch history, changed paths, and the user's request. If these sources conflict and choosing would change recorded history, ask the user before editing.
4. Find the previous documentation-sync boundary from commit history, such as a prior chapter commit. Compare all commits and working-tree changes after it. On the first run, inspect the repository's full history and current tree.
5. Preserve unrelated user changes. Include them in the commit only when they are clearly part of the completed chapter; otherwise leave them unstaged and report them.

## 2. Discover documentation dynamically

Build a fresh candidate inventory from both tracked and untracked, non-ignored files. Use `git ls-files`, `git status`, and `rg --files --hidden -g '!.git/**'` or equivalent.

Treat a file as documentation when its role or location indicates documentation, including:

- names such as `README*`, `AGENTS.md`, `JOURNEY.md`, `CHANGELOG*`, ADR indexes, runbooks, onboarding guides, and architecture records;
- documentation formats such as `.md`, `.mdx`, `.rst`, `.adoc`, `.txt`, `.mmd`, and `.puml`;
- files under documentation, context, runbook, or guardrail directories, even when introduced by a later chapter;
- new, renamed, or deleted documentation visible in Git history or working-tree status.

Exclude `.git/`, dependency/vendor trees, generated build output, caches, and this skill's own files unless the workflow itself changed. Do not assume every candidate needs editing: read its purpose and update it only when current work affects it.

## 3. Build an evidence ledger

Before writing, list the facts introduced or changed during the chapter:

- features, APIs, commands, configuration, infrastructure, and architecture;
- selected tools, rejected alternatives, and recorded decision reasons;
- installed versions, image references, namespaces, node pools, and resource state;
- operating rules, risks, recovery procedures, and troubleshooting outcomes;
- documents added, renamed, superseded, or removed.

Derive facts from Git diffs and history, source files, manifests, build records, and read-only runtime queries. Never invent completion, versions, dates, resources, or test results. For Kubernetes, always use the context required by `AGENTS.md`; in this repository use `--context gke-sysnet4admin_book_gitaiops` on every `kubectl` command.

## 4. Synchronize every affected document

Match the evidence ledger against the dynamic inventory. Handle both existing documents and documents introduced in the current or later chapters.

- `JOURNEY.md`: mark every actually completed subchapter, record the completion date, tool and architecture decisions, observed versions, current resources, and new troubleshooting lessons. Query live state for mutable facts when available.
- `README*` and onboarding or runbooks: update current capabilities, structure, prerequisites, commands, deployment state, and links. Remove language that describes completed work as future work.
- `AGENTS.md` and context or rule documents: update durable project facts, commands, safety rules, and validation expectations only when they changed.
- ADR documents: if present, add every new decision in chronological order without renumbering existing records. Follow the document's established format and verify ordering.
- Architecture or context snapshots: if present and architecture changed, update components, connections, namespaces, node pools, versions, and the snapshot milestone.
- Guardrails and operational procedures: if present, incorporate new hazards and verified recovery steps without weakening existing safety controls.
- Newly added documents: ensure their facts match the rest of the repository and add navigation links from an appropriate index or README when that improves discovery.
- Renamed or deleted documents: repair stale links and references throughout the repository.
- Any future document: infer its contract from its title, headings, neighboring files, and inbound links; update it when evidence affects that contract. Do not require a skill change merely because a new document type appears.

Retain intentional historical statements as history. Clearly distinguish current state from planned state.

## 5. Cross-check and validate

1. Search for stale versions, old names, obsolete commands, broken relative links, unchecked rows for completed work, and contradictory current-state claims.
2. Re-read all edited documents and inspect `git diff --check` plus the full diff.
3. Run available documentation checks. Validate referenced commands or runtime facts with the least invasive read-only checks practical.
4. Confirm secrets, credentials, private keys, generated artifacts, and unrelated changes are not staged.
5. If a required fact cannot be verified, state the limitation in the relevant document or stop and ask; do not fill it by inference.

## 6. Commit the chapter

1. Stage all and only intended completed-chapter changes, including implementation files when they belong to that chapter, not just documentation.
2. Review `git diff --cached --name-status`, `git diff --cached --stat`, and the staged diff.
3. If nothing changed, do not create an empty commit; report that documentation is already current.
4. Create one commit whose subject summarizes the chapter outcome, for example `ch4: add observability and sync documentation`. Follow an established repository convention when one exists.
5. Verify the new commit with `git status --short --branch` and `git log -1 --oneline`.
6. Do not push unless the user explicitly requests a push or an applicable repository instruction requires it.

Report the discovered document set, files updated, runtime facts verified, checks performed, commit hash, and any intentionally unstaged files.
