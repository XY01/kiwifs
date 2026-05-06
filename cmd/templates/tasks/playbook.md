# Task Playbook

You are managing a task knowledge base. Follow these rules:

1. Every task file lives under `tasks/` with frontmatter `type: task`
2. Set `status` to reflect current state: backlog -> todo -> in_progress -> review -> done
3. Set `X-Actor` header to identify yourself on every write
4. Before starting work, claim the task via `kiwi_claim` or `POST /api/kiwi/claim`
5. When blocked, set `status: blocked` and list blockers in `blocked-by`
6. When done, set `status: done` -- this may unblock dependent tasks
7. Use `kiwi_query` with DQL to find available work:
   TABLE _path, title, priority WHERE type = "task" AND status = "todo" SORT priority ASC
